package router

import (
	"github.com/blog/conversational-agent/internal/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers all application routes
func RegisterRoutes(e *echo.Echo, agentHandler *handlers.AgentHandler) {
	e.POST("/v1/agent/query/:org_id/:user_id/:thread_id", agentHandler.QueryHandler)
	e.GET("/v1/agent/memory/thread/:thread_id", agentHandler.GetMemoryHandler)
	e.POST("/v1/agent/memory/update", agentHandler.AddDocumentHandler)
	e.POST("/v1/agent/memory/import/:org_id/:user_id", agentHandler.ImportMemoryHandler)
}
