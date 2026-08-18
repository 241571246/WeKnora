package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSystemInfoExposesVONEACLDiagnostics(t *testing.T) {
	t.Setenv("WEKNORA_VONE_KB_ACL_MODE", "shadow")
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", "/api/v1/system/info", nil)
	(&SystemHandler{}).GetSystemInfo(c)

	var payload struct {
		Data GetSystemInfoResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.KBACLMode != "shadow" {
		t.Fatalf("kb_acl_mode=%q", payload.Data.KBACLMode)
	}
}
