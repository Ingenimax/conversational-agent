package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blog/conversational-agent/internal/logger"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
)

// GetThreadMemory retrieves or initializes a ConversationBuffer for the given thread ID.
func (am *AgentManager) GetThreadMemory(threadID string) *memory.ConversationBuffer {
	am.memoryMutex.Lock()
	defer am.memoryMutex.Unlock()

	// Check if the memory for the thread already exists
	if mem, exists := am.AgentMemory[threadID]; exists {
		return mem
	}

	// Initialize a new memory buffer
	newMemory := memory.NewConversationBuffer(
		memory.WithMemoryKey(fmt.Sprintf("thread:%s", threadID)),
		memory.WithReturnMessages(false),
	)
	am.AgentMemory[threadID] = newMemory
	return newMemory
}

// RetrieveMemory retrieves the chat history for a specific thread
func (am *AgentManager) RetrieveMemory(ctx context.Context, threadID string) ([]map[string]string, error) {
	log := logger.GetLogger()
	log.Debug().Msgf("Starting RetrieveMemory function for thread ID: %s", threadID)

	// Get the specific memory for the thread
	threadMemory := am.GetThreadMemory(threadID)
	if threadMemory == nil {
		log.Debug().Msgf("AgentMemory for thread ID '%s' is not initialized.", threadID)
		return nil, fmt.Errorf("memory is not initialized for thread ID: %s", threadID)
	}

	// Retrieve messages from the chat history
	messages, err := threadMemory.ChatHistory.Messages(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve messages from chat history.")
		return nil, fmt.Errorf("failed to retrieve messages for thread_id %s: %w", threadID, err)
	}

	// Log retrieved messages
	log.Debug().Msgf("Retrieved messages for thread ID %s: %+v", threadID, messages)

	// Format the messages for output
	formattedMessages := formatMessages(messages)
	return formattedMessages, nil
}

// formatMessages formats the messages for output
func formatMessages(messages []llms.ChatMessage) []map[string]string {
	var formattedMessages []map[string]string
	for _, msg := range messages {
		role := "user"
		if msg.GetType() == llms.ChatMessageTypeAI {
			role = "ai"
		}
		formattedMessages = append(formattedMessages, map[string]string{
			"role":    role,
			"content": msg.GetContent(),
		})
	}
	return formattedMessages
}

// addToBuffer adds the input and response to the message buffer
func (am *AgentManager) addToBuffer(threadID, input, response, userID, orgID string) {
	am.bufferMutex.Lock()
	defer am.bufferMutex.Unlock()

	// Chunk the input and response for fine-grained storage
	chunks := ChunkContent(input+"\n"+response, 300) // Adjust chunk size
	for _, chunk := range chunks {
		doc := schema.Document{
			PageContent: chunk,
			Metadata: map[string]any{
				"thread_id": threadID,
				"user_id":   userID,
				"org_id":    orgID,
				"timestamp": time.Now().Format(time.RFC3339),
				"source":    "conversation",
			},
		}

		// Ensure uniqueness before adding
		am.messageBuffer = append(am.messageBuffer, doc)
	}

	// Flush the buffer if it exceeds max size
	if len(am.messageBuffer) >= am.maxBufferMessages {
		go am.flushBufferToWeaviate(context.Background())
	}
}

// flushBufferToWeaviate flushes the message buffer to the vector store
func (am *AgentManager) flushBufferToWeaviate(ctx context.Context) {
	am.bufferMutex.Lock()
	buffer := am.messageBuffer
	am.messageBuffer = nil
	am.bufferMutex.Unlock()

	if len(buffer) == 0 {
		return
	}

	// Deduplicate buffer entries before inserting into Weaviate
	uniqueDocs := deduplicateDocuments(buffer)

	log := logger.GetLogger()
	log.Debug().Msgf("Batch inserting %d unique messages into Weaviate...", len(uniqueDocs))

	// Assuming each document has org_id in its metadata
	if len(uniqueDocs) > 0 {
		orgID, ok := uniqueDocs[0].Metadata["org_id"].(string)
		if !ok {
			log.Error().Msg("Failed to retrieve org_id from document metadata.")
			return
		}

		namespace := orgID
		log.Debug().Msgf("Using namespace: %s for flushing buffer to Weaviate", namespace)

		_, err := am.VectorStore.AddDocuments(ctx, uniqueDocs, vectorstores.WithNameSpace(namespace))
		if err != nil {
			log.Error().Err(err).Msg("Failed to batch insert messages into Weaviate.")
			return
		}
	}

	log.Debug().Msg("Batch insertion to Weaviate completed successfully.")
}

// deduplicateDocuments removes duplicate documents from a slice based on PageContent.
func deduplicateDocuments(docs []schema.Document) []schema.Document {
	seen := make(map[string]bool)
	uniqueDocs := []schema.Document{}

	for _, doc := range docs {
		if _, exists := seen[doc.PageContent]; !exists {
			seen[doc.PageContent] = true
			uniqueDocs = append(uniqueDocs, doc)
		}
	}

	return uniqueDocs
}

// SyncMemory syncs the message buffer to the vector store
func (am *AgentManager) SyncMemory(ctx context.Context) error {
	log := logger.GetLogger()

	am.bufferMutex.Lock()
	defer am.bufferMutex.Unlock()

	if len(am.messageBuffer) == 0 {
		log.Debug().Msg("No messages to sync from buffer.")
		return nil
	}

	log.Debug().Msgf("Syncing %d messages from buffer to Weaviate...", len(am.messageBuffer))
	_, err := am.VectorStore.AddDocuments(ctx, am.messageBuffer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to sync messages to Weaviate.")
		return err
	}

	am.messageBuffer = nil
	log.Debug().Msg("Memory synced successfully to Weaviate.")
	return nil
}

// AddDocuments adds documents to the vector store
func (am *AgentManager) AddDocuments(ctx context.Context, docs []schema.Document, userID, orgID string) error {
	log := logger.GetLogger()

	// Add user_id to document metadata
	for i := range docs {
		docs[i].Metadata["user_id"] = userID
	}

	// Use org_id as namespace
	namespace := orgID
	log.Debug().Msgf("Using namespace: %s for adding documents", namespace)

	// Retry mechanism for adding documents
	maxRetries := 3
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err = am.VectorStore.AddDocuments(ctx, docs, vectorstores.WithNameSpace(namespace))
		if err != nil {
			log.Warn().Err(err).Msgf("Attempt %d to add documents to vector store failed.", attempt)
			if attempt < maxRetries {
				continue
			}
		} else {
			log.Info().Msgf("Documents added to vector store successfully on attempt %d.", attempt)
			return nil
		}
	}

	// Log final failure after retries
	log.Error().Err(err).Msg("Failed to add documents to vector store after multiple attempts.")
	return fmt.Errorf("failed to add documents to vector store: %w", err)
}

// AddDataset adds a dataset to the vector store
func (am *AgentManager) AddDataset(ctx context.Context, threadID, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Parse the dataset
	var documents []schema.Document
	if err := json.NewDecoder(file).Decode(&documents); err != nil {
		return fmt.Errorf("failed to parse JSON dataset: %w", err)
	}

	// Retrieve the thread-specific memory
	threadMemory := am.GetThreadMemory(threadID)
	if threadMemory == nil {
		return fmt.Errorf("failed to retrieve memory for thread ID: %s", threadID)
	}

	// Add each document to the memory
	for _, doc := range documents {
		err := threadMemory.ChatHistory.AddAIMessage(ctx, doc.PageContent)
		if err != nil {
			return fmt.Errorf("failed to add document to memory for thread ID %s: %w", threadID, err)
		}
	}

	return nil
}

// Chunk content into smaller parts based on word count
func ChunkContent(content string, maxWords int) []string {
	words := strings.Fields(content)
	var chunks []string

	for i := 0; i < len(words); i += maxWords {
		end := i + maxWords
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[i:end], " "))
	}

	return chunks
}

// QueryMemoryByUserAndOrgID queries the memory for a specific user and organization
func (am *AgentManager) QueryMemoryByUserAndOrgID(ctx context.Context, userID, orgID string) ([]schema.Document, error) {
	log := logger.GetLogger()
	log.Debug().Msgf("Querying memory for user_id: %s and org_id: %s", userID, orgID)

	// Build metadata filters using filters.Where
	whereBuilder := filters.Where().
		WithOperator(filters.And).
		WithOperands([]*filters.WhereBuilder{
			filters.Where().WithPath([]string{"user_id"}).WithOperator(filters.Equal).WithValueString(userID),
			filters.Where().WithPath([]string{"org_id"}).WithOperator(filters.Equal).WithValueString(orgID),
		})

	// Execute MetadataSearch with the constructed filter
	docs, err := am.VectorStore.MetadataSearch(ctx, 10, vectorstores.WithFilters(whereBuilder))
	if err != nil {
		log.Error().Err(err).Msg("Failed to query memory by user_id and org_id.")
		return nil, fmt.Errorf("failed to query memory for user_id %s and org_id %s: %w", userID, orgID, err)
	}

	log.Debug().Msgf("Retrieved %d documents for user_id: %s and org_id: %s", len(docs), userID, orgID)
	return docs, nil
}

// ConvertContentToString converts content to a string
func ConvertContentToString(content interface{}) (string, error) {
	switch v := content.(type) {
	case string:
		return v, nil
	case map[string]interface{}, []interface{}:
		// Serialize JSON content to a string
		jsonContent, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to serialize JSON content: %w", err)
		}
		return string(jsonContent), nil
	default:
		return "", fmt.Errorf("unsupported content type: %T", content)
	}
}

// Convert metadata values
func ConvertMetadata(metadata map[string]string) map[string]any {
	converted := make(map[string]any)
	for k, v := range metadata {
		converted[k] = v
	}
	return converted
}
