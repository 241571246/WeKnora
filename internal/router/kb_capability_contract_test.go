package router

import (
	"os"
	"strings"
	"testing"
)

func TestSensitiveKnowledgeRoutesDeclareExactKBCapabilities(t *testing.T) {
	routes, err := os.ReadFile("routes_knowledge.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(routes)
	contracts := []string{
		`GET("/:knowledge_id", g.Viewer(), g.KBCapabilityFromKnowledgeID("knowledge_id", types.KBCapabilityChunkPreview)`,
		`DELETE("/:knowledge_id/:id", g.KBCapabilityFromKnowledgeID("knowledge_id", types.KBCapabilityChunkDelete)`,
		`PUT("/:knowledge_id/:id", g.KBCapabilityFromKnowledgeID("knowledge_id", types.KBCapabilityChunkEdit)`,
		`POST("/file", g.KBCapability("id", types.KBCapabilityDocumentUpload)`,
		`GET("", g.Viewer(), g.KBCapability("id", types.KBCapabilityDocumentsList)`,
		`DELETE("", g.KBCapability("id", types.KBCapabilityDocumentDelete)`,
		`GET("/:id", g.Viewer(), g.KBCapabilityFromKnowledgeID("id", types.KBCapabilityDocumentPreview)`,
		`DELETE("/:id", g.KBCapabilityFromKnowledgeID("id", types.KBCapabilityDocumentDelete)`,
		`PUT("/:id", g.KBCapabilityFromKnowledgeID("id", types.KBCapabilityDocumentEdit)`,
		`POST("/:id/reparse", g.KBCapabilityFromKnowledgeID("id", types.KBCapabilityDocumentReparse)`,
		`GET("/:id/download", g.Viewer(), g.KBCapabilityFromKnowledgeID("id", types.KBCapabilityDocumentDownload)`,
		`GET("/:id/preview", g.Viewer(), g.KBCapabilityFromKnowledgeID("id", types.KBCapabilityDocumentPreview)`,
		`DELETE("/:id", g.KBCapability("id", types.KBCapabilityDelete)`,
		`POST("/:id/hybrid-search", g.Viewer(), g.KBCapability("id", types.KBCapabilityAIQuery)`,
	}
	for _, contract := range contracts {
		if !strings.Contains(source, contract) {
			t.Errorf("missing sensitive route capability contract: %s", contract)
		}
	}
}

func TestMemberAndCollectionMutationRoutesRemainCapabilityGuarded(t *testing.T) {
	routes, err := os.ReadFile("routes_kb_acl.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(routes)
	contracts := []string{
		`g.KBCapability("id", types.KBCapabilityMembersManage), h.ListMembers`,
		`g.KBCapability("id", types.KBCapabilityMembersManage), h.AddMember`,
		`g.KBCapability("id", types.KBCapabilityMembersManage), h.UpdateMember`,
		`g.KBCapability("id", types.KBCapabilityMembersManage), h.DeleteMember`,
		`r.POST("/knowledge-base-collections", g.Owner(), h.Create)`,
		`r.PATCH("/knowledge-base-collections/:id", g.Owner(), h.Update)`,
		`r.DELETE("/knowledge-base-collections/:id", g.Owner(), h.Delete)`,
		`g.KBCapability("id", types.KBCapabilitySettingsEdit), h.SetBinding`,
	}
	for _, contract := range contracts {
		if !strings.Contains(source, contract) {
			t.Errorf("missing ACL route contract: %s", contract)
		}
	}
}
