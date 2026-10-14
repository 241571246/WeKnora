package router

import (
	"bytes"
	"io"
	"strings"

	apperrors "github.com/Tencent/WeKnora/internal/errors"
	"github.com/Tencent/WeKnora/internal/middleware"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/gin-gonic/gin"
)

// Keep the cross-KB search at the same AI-query capability as the existing
// single-KB wiki search, without consuming the JSON body needed by its handler.
func (g *rbacGuards) WikiSearchKBCapabilities() gin.HandlerFunc {
	return middleware.RequireKBCapabilities(func(c *gin.Context) ([]string, error) {
		var body struct {
			ID  string   `json:"knowledge_base_id"`
			IDs []string `json:"knowledge_base_ids"`
		}
		if err := c.ShouldBindBodyWithJSON(&body); err != nil {
			return nil, apperrors.NewBadRequestError("invalid wiki search request")
		}
		if cached, ok := c.Get(gin.BodyBytesKey); ok {
			c.Request.Body = io.NopCloser(bytes.NewReader(cached.([]byte)))
		}
		seen := make(map[string]bool)
		ids := make([]string, 0, len(body.IDs)+1)
		for _, id := range append(body.IDs, body.ID) {
			id = strings.TrimSpace(id)
			if id != "" && !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
		if len(ids) > 32 {
			return nil, apperrors.NewBadRequestError("at most 32 knowledge_base_ids are allowed")
		}
		return ids, nil
	}, types.KBCapabilityAIQuery, g.kbAuthorizer)
}
