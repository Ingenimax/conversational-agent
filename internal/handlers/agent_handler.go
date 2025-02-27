// internal/handlers/agent_handler.go
package handlers

import (
	"fmt"
	"net/http"

	"github.com/blog/conversational-agent/internal/agents"
	"github.com/labstack/echo/v4"
)

type AgentHandler struct {
	AgentManager *agents.AgentManager
}

// QueryHandler handles the query request from the client.
func (h *AgentHandler) QueryHandler(c echo.Context) error {
	type RequestBody struct {
		Query  string `json:"query"`
		Stream bool   `json:"stream"`
	}

	var req RequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Retrieve query parameters
	userID := c.Param("user_id")
	orgID := c.Param("org_id")
	threadID := c.Param("thread_id")

	// Validate required parameters
	if threadID == "" || userID == "" || orgID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "'thread_id', 'user_id', and 'org_id' are required query parameters",
		})
	}

	if req.Stream {
		// Streamed response
		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().WriteHeader(http.StatusOK)

		_, err := h.AgentManager.Query(
			c.Request().Context(),
			userID,
			orgID,
			threadID,
			req.Query,
			func(chunk []byte) {
				c.Response().Write([]byte(fmt.Sprintf("data: %s\n\n", chunk)))
				c.Response().Flush()
			},
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		c.Response().Write([]byte("data: [DONE]\n\n"))
		c.Response().Flush()

		return nil
	}

	// Non-streamed response
	response, err := h.AgentManager.Query(c.Request().Context(), userID, orgID, threadID, req.Query, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"response": response})
}
