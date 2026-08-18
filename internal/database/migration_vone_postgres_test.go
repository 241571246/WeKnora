package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/lib/pq"
)

// TestVONEPostgresMigrationRehearsal is opt-in because it requires an isolated
// PostgreSQL database. It validates the real upstream schema, VONE backfill,
// retained cross-workspace shares, and an extension-only rollback.
func TestVONEPostgresMigrationRehearsal(t *testing.T) {
	dsn := os.Getenv("VONE_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set VONE_TEST_POSTGRES_DSN to an isolated PostgreSQL database")
	}

	_, filename, _, _ := runtime.Caller(0)
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	oldCWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldCWD) })

	if err := RunMigrationsWithOptions(dsn, MigrationOptions{}); err != nil {
		t.Fatalf("upstream migration: %v", err)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	seedSQL := `
INSERT INTO tenants (id, name, business) VALUES
  (70001, 'VONE migration rehearsal', 'test');
INSERT INTO users (id, username, email, password_hash, tenant_id, created_at) VALUES
  ('owner-00000000-0000-0000-0000000001', 'vone-owner', 'vone-owner@example.invalid', 'test', 70001, '2026-01-01'),
  ('creator-000000-0000-0000-0000000001', 'vone-creator', 'vone-creator@example.invalid', 'test', 70001, '2026-01-02');
INSERT INTO tenant_members (user_id, tenant_id, role, status, joined_at) VALUES
  ('owner-00000000-0000-0000-0000000001', 70001, 'owner', 'active', '2026-01-01'),
  ('creator-000000-0000-0000-0000000001', 70001, 'viewer', 'active', '2026-01-02');
INSERT INTO knowledge_bases (id, name, tenant_id, embedding_model_id, summary_model_id, creator_id) VALUES
  ('kb-valid-00000000-0000-0000-00000001', 'valid creator', 70001, '', '', 'creator-000000-0000-0000-0000000001'),
  ('kb-fallback-000000-0000-0000-000001', 'fallback creator', 70001, '', '', 'missing-0000000-0000-0000-0000000001');
INSERT INTO organizations (id, name, owner_id, owner_tenant_id) VALUES
  ('org-00000000-0000-0000-0000-00000001', 'VONE share rehearsal', 'owner-00000000-0000-0000-0000000001', 70001);
INSERT INTO kb_shares (id, knowledge_base_id, organization_id, shared_by_user_id, source_tenant_id, permission) VALUES
  ('share-0000000-0000-0000-0000-0000001', 'kb-valid-00000000-0000-0000-00000001', 'org-00000000-0000-0000-0000-00000001', 'owner-00000000-0000-0000-0000000001', 70001, 'viewer');`
	if _, err := db.Exec(seedSQL); err != nil {
		t.Fatalf("seed pre-VONE data: %v", err)
	}

	if err := RunVONEMigrationsWithOptions(dsn, MigrationOptions{}); err != nil {
		t.Fatalf("VONE migration: %v", err)
	}
	assertMembershipOwner(t, db, "kb-valid-00000000-0000-0000-00000001", "creator-000000-0000-0000-0000000001")
	assertMembershipOwner(t, db, "kb-fallback-000000-0000-0000-000001", "owner-00000000-0000-0000-0000000001")

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM kb_shares WHERE id = 'share-0000000-0000-0000-0000-0000001'").Scan(&count); err != nil || count != 1 {
		t.Fatalf("kb_shares preserved count=%d err=%v", count, err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM vone_schema_migrations WHERE version = 1 AND dirty = false").Scan(&count); err != nil || count != 1 {
		t.Fatalf("VONE migration ledger count=%d err=%v", count, err)
	}

	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}
	m, err := migrate.New("file://"+filepath.ToSlash(filepath.Join(repoRoot, "migrations", "vone", "versioned")), dsn+separator+"x-migrations-table="+voneMigrationTable)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Down(); err != nil {
		t.Fatal(err)
	}
	_, _ = m.Close()

	if err := db.QueryRow("SELECT COUNT(*) FROM kb_shares").Scan(&count); err != nil || count != 1 {
		t.Fatalf("kb_shares after rollback count=%d err=%v", count, err)
	}
	if _, err := db.Exec("SELECT 1 FROM vone_kb_memberships LIMIT 1"); err == nil {
		t.Fatal("vone_kb_memberships still exists after rollback")
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE dirty = false").Scan(&count); err != nil || count == 0 {
		t.Fatalf("upstream migration ledger after VONE rollback count=%d err=%v", count, err)
	}

	// A short ping catches asynchronous container/database shutdown mistakes in
	// local rehearsal scripts without adding a long-running health check.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("database unavailable after VONE rollback: %v", err)
	}
}

func assertMembershipOwner(t *testing.T, db *sql.DB, kbID, wantUserID string) {
	t.Helper()
	var got string
	if err := db.QueryRow("SELECT user_id FROM vone_kb_memberships WHERE kb_id = $1 AND role = 'owner' AND deleted_at IS NULL", kbID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != wantUserID {
		t.Fatalf("owner for %s=%q want=%q", kbID, got, wantUserID)
	}
}
