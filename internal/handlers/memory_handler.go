package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/blog/conversational-agent/internal/agents"
	"github.com/blog/conversational-agent/internal/logger"
	"github.com/labstack/echo/v4"
	"github.com/tmc/langchaingo/schema"
)

type ImportDatasetRequest struct {
	FilePath string `json:"file_path"`
}

// GetMemoryHandler retrieves the chat history for a specific thread.
func (h *AgentHandler) GetMemoryHandler(c echo.Context) error {
	ctx := c.Request().Context()

	// Get threadID from the request (e.g., query parameter or header)
	threadID := c.Param("thread_id")
	if threadID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing thread_id"})
	}

	log := logger.GetLogger()

	// Retrieve memory from AgentManager
	memory, err := h.AgentManager.RetrieveMemory(ctx, threadID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if len(memory) == 0 {
		log.Debug().Msgf("Handler: No memory found for thread ID: %s", threadID)
		return c.JSON(http.StatusOK, map[string]string{"memory": "Memory is empty"})
	}

	return c.JSON(http.StatusOK, map[string]any{"memory": memory})
}

// AddDocumentHandler processes and stores documents with chunking and metadata.
func (h *AgentHandler) AddDocumentHandler(c echo.Context) error {
	type AddDocumentRequest struct {
		Content  interface{}       `json:"page_content"` // Accepts string or JSON
		Metadata map[string]string `json:"metadata"`
	}

	var req AddDocumentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Convert Content to a string
	content, err := agents.ConvertContentToString(req.Content)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Create a document to add
	doc := schema.Document{
		PageContent: content,
		Metadata:    agents.ConvertMetadata(req.Metadata),
	}

	// Add document to Weaviate vector store
	ctx := c.Request().Context()
	docID, err := h.AgentManager.VectorStore.AddDocuments(ctx, []schema.Document{doc})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add document"})
	}

	return c.JSON(http.StatusOK, map[string]string{"doc_id": docID[0]})
}

// ImportMemoryHandler imports a dataset into the vector store.
func (h *AgentHandler) ImportMemoryHandler(c echo.Context) error {
	type ImportRequest struct {
		FilePath string `json:"file_path"`
		UserID   string `json:"user_id"`
		OrgID    string `json:"org_id"`
	}

	var req ImportRequest
	if err := c.Bind(&req); err != nil || req.FilePath == "" || req.UserID == "" || req.OrgID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload. 'file_path', 'user_id', and 'org_id' are required.",
		})
	}

	// Open the JSON file
	file, err := os.Open(req.FilePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to open the file."})
	}
	defer file.Close()

	// Parse the JSON file
	var dataset []map[string]interface{}
	if err := json.NewDecoder(file).Decode(&dataset); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON file format."})
	}

	// Convert dataset into schema.Document format
	docs := []schema.Document{}
	for _, item := range dataset {
		if summary, ok := item["summary"].(string); ok {
			doc := schema.Document{
				PageContent: summary,
				Metadata:    item,
			}
			docs = append(docs, doc)
		}
	}

	// Add documents to vector store with userID and orgID
	ctx := context.Background()
	if err := h.AgentManager.AddDocuments(ctx, docs, req.UserID, req.OrgID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add documents to vector store."})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Dataset imported successfully."})
}
