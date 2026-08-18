package service

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/Tencent/WeKnora/internal/application/repository"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type captureCollectionAudit struct {
	interfaces.AuditLogService
	entries []*types.AuditLog
}

func (a *captureCollectionAudit) Log(_ context.Context, entry *types.AuditLog) error {
	a.entries = append(a.entries, entry)
	return nil
}

func newKBCollectionTestService(t *testing.T) (*kbCollectionService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&types.KnowledgeBase{}, &types.KBCollection{}, &types.KBCollectionBinding{}); err != nil {
		t.Fatal(err)
	}
	return &kbCollectionService{repo: repository.NewKBCollectionRepository(db), kbs: repository.NewKnowledgeBaseRepository(db)}, db
}

func TestKBCollectionRejectsCycleAndNonEmptyDelete(t *testing.T) {
	svc, _ := newKBCollectionTestService(t)
	ctx := context.Background()
	root := &types.KBCollection{ID: "root", TenantID: 7, Name: "Root"}
	rootID := root.ID
	child := &types.KBCollection{ID: "child", TenantID: 7, Name: "Child", ParentID: &rootID}
	if err := svc.Create(ctx, root); err != nil {
		t.Fatal(err)
	}
	if err := svc.Create(ctx, child); err != nil {
		t.Fatal(err)
	}
	childID := child.ID
	root.ParentID = &childID
	if err := svc.Update(ctx, root); !errors.Is(err, ErrKBCollectionCycle) {
		t.Fatalf("cycle update error = %v, want %v", err, ErrKBCollectionCycle)
	}
	if err := svc.Delete(ctx, 7, root.ID); !errors.Is(err, ErrKBCollectionNotEmpty) {
		t.Fatalf("non-empty delete error = %v, want %v", err, ErrKBCollectionNotEmpty)
	}
}

func TestKBCollectionTreeFiltersBindingsAndKeepsAncestors(t *testing.T) {
	svc, db := newKBCollectionTestService(t)
	ctx := context.Background()
	root := &types.KBCollection{ID: "root", TenantID: 7, Name: "Root"}
	rootID := root.ID
	child := &types.KBCollection{ID: "child", TenantID: 7, Name: "Child", ParentID: &rootID}
	if err := svc.Create(ctx, root); err != nil {
		t.Fatal(err)
	}
	if err := svc.Create(ctx, child); err != nil {
		t.Fatal(err)
	}
	for _, kb := range []*types.KnowledgeBase{{ID: "visible", TenantID: 7, Name: "Visible"}, {ID: "hidden", TenantID: 7, Name: "Hidden"}} {
		if err := db.Create(kb).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, binding := range []*types.KBCollectionBinding{{TenantID: 7, KBID: "visible", CollectionID: "child"}, {TenantID: 7, KBID: "hidden", CollectionID: "root"}} {
		if err := svc.SetKnowledgeBaseCollection(ctx, binding); err != nil {
			t.Fatal(err)
		}
	}
	collections, bindings, err := svc.GetTree(ctx, 7, []string{"visible"})
	if err != nil {
		t.Fatal(err)
	}
	if len(collections) != 2 || len(bindings) != 1 || bindings[0].KBID != "visible" {
		t.Fatalf("filtered tree collections=%+v bindings=%+v", collections, bindings)
	}
}

func TestKBCollectionTreeUsesConstantQueryCount(t *testing.T) {
	svc, db := newKBCollectionTestService(t)
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		if err := svc.Create(ctx, &types.KBCollection{ID: fmt.Sprintf("c-%03d", i), TenantID: 7, Name: fmt.Sprintf("Collection %03d", i)}); err != nil {
			t.Fatal(err)
		}
	}
	var queries atomic.Int64
	if err := db.Callback().Query().Before("gorm:query").Register("vone:test_query_count", func(*gorm.DB) {
		queries.Add(1)
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.GetTree(ctx, 7, nil); err != nil {
		t.Fatal(err)
	}
	if got := queries.Load(); got != 2 {
		t.Fatalf("GetTree queries=%d, want 2 regardless of node count", got)
	}
}

func TestKBCollectionUnclassifiedMoveIsAudited(t *testing.T) {
	svc, db := newKBCollectionTestService(t)
	audit := &captureCollectionAudit{}
	svc.audit = audit
	ctx := context.WithValue(context.Background(), types.UserIDContextKey, "owner")
	if err := db.Create(&types.KnowledgeBase{ID: "kb-1", TenantID: 7, Name: "KB"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := svc.Create(ctx, &types.KBCollection{ID: "c-1", TenantID: 7, Name: "Collection"}); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetKnowledgeBaseCollection(ctx, &types.KBCollectionBinding{KBID: "kb-1", TenantID: 7, CollectionID: "c-1", UpdatedBy: "owner"}); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetKnowledgeBaseCollection(ctx, &types.KBCollectionBinding{KBID: "kb-1", TenantID: 7, UpdatedBy: "owner"}); err != nil {
		t.Fatal(err)
	}
	last := audit.entries[len(audit.entries)-1]
	if last.Action != types.AuditActionKBCollectionBound || last.ScopeID != "kb-1" || last.TargetID != "" {
		t.Fatalf("unexpected unclassified audit: %+v", last)
	}
}
