package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tencent/WeKnora/internal/application/repository"
	appservice "github.com/Tencent/WeKnora/internal/application/service"
	"github.com/Tencent/WeKnora/internal/middleware"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ownerGuardMembershipService struct {
	interfaces.KBMembershipService
	member  *types.KBMembership
	updates int
	deletes int
}

func TestKBAccessEndpointDoesNotCrossWorkspaceBoundary(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&types.KnowledgeBase{}, &types.KBMembership{}, &types.KBMemberCapability{}); err != nil {
		t.Fatal(err)
	}
	kbs := repository.NewKnowledgeBaseRepository(db)
	members := repository.NewKBMembershipRepository(db)
	ctx := context.Background()
	if err := kbs.CreateKnowledgeBase(ctx, &types.KnowledgeBase{ID: "tenant-8-kb", TenantID: 8, Name: "secret"}); err != nil {
		t.Fatal(err)
	}
	if err := members.Create(ctx, &types.KBMembership{
		TenantID: 8, KBID: "tenant-8-kb", UserID: "employee", Role: types.KBMemberRoleDocumentViewer, GrantedBy: "owner",
	}, nil); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewKBACLHandler(nil, appservice.NewKBAuthorizer(members, kbs, nil, nil))
	router.GET("/knowledge-bases/:id/access", h.GetAccess)
	req := httptest.NewRequest(http.MethodGet, "/knowledge-bases/tenant-8-kb/access?capability=kb.metadata.read", nil)
	requestContext := context.WithValue(req.Context(), types.TenantIDContextKey, uint64(7))
	requestContext = context.WithValue(requestContext, types.UserIDContextKey, "employee")
	requestContext = context.WithValue(requestContext, types.TenantRoleContextKey, types.TenantRoleViewer)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req.WithContext(requestContext))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Data types.KBPolicyDecision `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Allowed || response.Data.Source != types.KBPolicySourceNone {
		t.Fatalf("cross-workspace decision=%+v, want non-revealing denial", response.Data)
	}
}

func (s *ownerGuardMembershipService) Get(context.Context, uint64, string, string) (*types.KBMembership, error) {
	return s.member, nil
}

func (s *ownerGuardMembershipService) Update(context.Context, string, *types.KBMembership, []types.KBCapability) error {
	s.updates++
	return nil
}

func (s *ownerGuardMembershipService) Delete(context.Context, string, uint64, string, string) error {
	s.deletes++
	return nil
}

type ownerGuardAuthorizer struct {
	interfaces.KBAuthorizer
	allowOwnersManage bool
}

func (a *ownerGuardAuthorizer) Authorize(_ context.Context, req interfaces.KBPolicyRequest) (types.KBPolicyDecision, error) {
	return types.KBPolicyDecision{Allowed: a.allowOwnersManage, Capability: req.Capability}, nil
}

func runOwnerMutationRequest(t *testing.T, method, body string, service *ownerGuardMembershipService) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := NewKBACLHandler(service, &ownerGuardAuthorizer{})
	r.PATCH("/knowledge-bases/:id/members/:user_id", h.UpdateMember)
	r.DELETE("/knowledge-bases/:id/members/:user_id", h.DeleteMember)
	req := httptest.NewRequest(method, "/knowledge-bases/kb-1/members/co-owner", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), types.TenantIDContextKey, uint64(7))
	ctx = context.WithValue(ctx, types.UserIDContextKey, "member-manager")
	ctx = context.WithValue(ctx, types.TenantRoleContextKey, types.TenantRoleViewer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req.WithContext(ctx))
	return w.Code
}

func TestKBMemberManagerCannotDemoteOwnerWithoutOwnersManage(t *testing.T) {
	svc := &ownerGuardMembershipService{member: &types.KBMembership{Role: types.KBMemberRoleOwner}}
	status := runOwnerMutationRequest(t, http.MethodPatch, `{"role":"editor"}`, svc)
	if status != http.StatusForbidden || svc.updates != 0 {
		t.Fatalf("status=%d updates=%d, want 403 and no update", status, svc.updates)
	}
}

func TestKBMemberManagerCannotDeleteOwnerWithoutOwnersManage(t *testing.T) {
	svc := &ownerGuardMembershipService{member: &types.KBMembership{Role: types.KBMemberRoleOwner}}
	status := runOwnerMutationRequest(t, http.MethodDelete, "", svc)
	if status != http.StatusForbidden || svc.deletes != 0 {
		t.Fatalf("status=%d deletes=%d, want 403 and no delete", status, svc.deletes)
	}
}

func TestKBMemberManagerCanDeleteOrdinaryMemberWithoutOwnersManage(t *testing.T) {
	svc := &ownerGuardMembershipService{member: &types.KBMembership{Role: types.KBMemberRoleEditor}}
	status := runOwnerMutationRequest(t, http.MethodDelete, "", svc)
	if status != http.StatusOK || svc.deletes != 1 {
		t.Fatalf("status=%d deletes=%d, want 200 and one delete", status, svc.deletes)
	}
}
