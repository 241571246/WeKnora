package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	sqlite3migrate "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func TestVONESQLiteMigrationBackfillPreservesSharesAndRollsBack(t *testing.T) {
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

	dbPath := filepath.Join(t.TempDir(), "vone-migration.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	baseSQL := `
CREATE TABLE tenant_members (id INTEGER PRIMARY KEY, tenant_id INTEGER, user_id TEXT, role TEXT, status TEXT, joined_at DATETIME, deleted_at DATETIME);
CREATE TABLE knowledge_bases (id TEXT PRIMARY KEY, tenant_id INTEGER, creator_id TEXT, deleted_at DATETIME);
CREATE TABLE kb_shares (id INTEGER PRIMARY KEY, knowledge_base_id TEXT);
INSERT INTO tenant_members VALUES (1, 7, 'owner-1', 'owner', 'active', '2026-01-01', NULL);
INSERT INTO tenant_members VALUES (2, 7, 'creator-1', 'viewer', 'active', '2026-01-02', NULL);
INSERT INTO knowledge_bases VALUES ('kb-valid', 7, 'creator-1', NULL);
INSERT INTO knowledge_bases VALUES ('kb-fallback', 7, 'missing', NULL);
INSERT INTO kb_shares VALUES (1, 'kb-valid');`
	if _, err := db.Exec(baseSQL); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if err := RunVONEMigrationsWithOptions("", MigrationOptions{SQLiteDBPath: dbPath}); err != nil {
		t.Fatal(err)
	}
	if version, dirty, ok := CachedVONEMigrationVersion(); !ok || dirty || version != 1 {
		t.Fatalf("cached VONE migration state version=%d dirty=%v ok=%v", version, dirty, ok)
	}
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM vone_kb_memberships").Scan(&count); err != nil || count != 2 {
		t.Fatalf("membership count=%d err=%v", count, err)
	}
	var fallback string
	if err := db.QueryRow("SELECT user_id FROM vone_kb_memberships WHERE kb_id='kb-fallback'").Scan(&fallback); err != nil || fallback != "owner-1" {
		t.Fatalf("fallback owner=%q err=%v", fallback, err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM kb_shares").Scan(&count); err != nil || count != 1 {
		t.Fatalf("kb_shares count=%d err=%v", count, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	downDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	driver, err := sqlite3migrate.WithInstance(downDB, &sqlite3migrate.Config{MigrationsTable: voneMigrationTable})
	if err != nil {
		t.Fatal(err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(filepath.Join(repoRoot, "migrations", "vone", "sqlite")), "sqlite3", driver)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Down(); err != nil {
		t.Fatal(err)
	}
	_, _ = m.Close()

	checkDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer checkDB.Close()
	if err := checkDB.QueryRow("SELECT COUNT(*) FROM kb_shares").Scan(&count); err != nil || count != 1 {
		t.Fatalf("kb_shares after rollback count=%d err=%v", count, err)
	}
	if _, err := checkDB.Exec("SELECT 1 FROM vone_kb_memberships LIMIT 1"); err == nil {
		t.Fatal("vone_kb_memberships still exists after rollback")
	}
}
