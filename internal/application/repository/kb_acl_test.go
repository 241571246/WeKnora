package repository

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newKBACLTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&types.KBMembership{}, &types.KBMemberCapability{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestKBMembershipRepositoryProtectsFinalOwner(t *testing.T) {
	db := newKBACLTestDB(t)
	repo := NewKBMembershipRepository(db)
	ctx := context.Background()
	first := &types.KBMembership{TenantID: 7, KBID: "kb-1", UserID: "u-1", Role: types.KBMemberRoleOwner, GrantedBy: "u-1"}
	second := &types.KBMembership{TenantID: 7, KBID: "kb-1", UserID: "u-2", Role: types.KBMemberRoleOwner, GrantedBy: "u-1"}
	if err := repo.Create(ctx, first, nil); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, second, nil); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, 7, "kb-1", "u-2"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, 7, "kb-1", "u-1"); !errors.Is(err, ErrLastKBOwner) {
		t.Fatalf("last-owner delete error = %v, want %v", err, ErrLastKBOwner)
	}
}

func TestKBMembershipRepositoryConcurrentOwnerDeleteNeverOrphansKB(t *testing.T) {
	db := newKBACLTestDB(t)
	repo := NewKBMembershipRepository(db)
	ctx := context.Background()
	for _, userID := range []string{"u-1", "u-2"} {
		if err := repo.Create(ctx, &types.KBMembership{TenantID: 7, KBID: "kb-1", UserID: userID, Role: types.KBMemberRoleOwner, GrantedBy: "u-1"}, nil); err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, userID := range []string{"u-1", "u-2"} {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			<-start
			errs <- repo.Delete(ctx, 7, "kb-1", id)
		}(userID)
	}
	close(start)
	wg.Wait()
	close(errs)
	var successes int
	for err := range errs {
		if err == nil {
			successes++
		}
	}
	owners, err := repo.CountOwners(ctx, 7, "kb-1")
	if err != nil {
		t.Fatal(err)
	}
	if successes > 1 || owners < 1 {
		t.Fatalf("concurrent deletes successes=%d remaining owners=%d", successes, owners)
	}
}

func TestKBMembershipRepositoryPersistsOnlyCustomCapabilities(t *testing.T) {
	db := newKBACLTestDB(t)
	repo := NewKBMembershipRepository(db)
	ctx := context.Background()
	custom := &types.KBMembership{TenantID: 7, KBID: "kb-1", UserID: "u-1", Role: types.KBMemberRoleCustom, GrantedBy: "owner"}
	if err := repo.Create(ctx, custom, []types.KBCapability{types.KBCapabilityAIQuery}); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.Get(ctx, 7, "kb-1", "u-1")
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil || len(loaded.Capabilities) != 1 || loaded.Capabilities[0].Capability != types.KBCapabilityAIQuery {
		t.Fatalf("custom capabilities = %#v", loaded)
	}
}
