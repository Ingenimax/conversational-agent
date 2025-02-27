package agents

import (
	"bytes"
	"context"
	"fmt"

	"github.com/blog/conversational-agent/internal/logger"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/vectorstores"
)

// Query performs a similarity search and LLM chain call.
func (am *AgentManager) Query(
	ctx context.Context,
	userID, orgID, threadID, input string,
	chunkCallback func([]byte),
) (string, error) {
	log := logger.GetLogger()

	// Retrieve memory and prepare for search
	threadMemory := am.GetThreadMemory(threadID)
	log.Debug().Msg("Performing similarity search in vector store...")

	// Use org_id as namespace
	namespace := orgID
	log.Debug().Msgf("Using namespace: %s for similarity search", namespace)

	// Perform similarity search in org namespace
	log.Debug().Msgf("Performing similarity search in namespace: %s", namespace)
	orgDocs, err := am.VectorStore.SimilaritySearch(ctx, input, 5, vectorstores.WithNameSpace(namespace))
	if err != nil && err.Error() != "empty response" {
		log.Error().Err(err).Msg("Failed to perform org-specific similarity search.")
		return "", fmt.Errorf("org similarity search failed: %w", err)
	}

	// Search in default namespace
	log.Debug().Msgf("Performing similarity search in default namespace")
	defaultDocs, err := am.VectorStore.SimilaritySearch(ctx, input, 5)
	if err != nil && err.Error() != "empty response" {
		log.Error().Err(err).Msg("Failed to perform default similarity search.")
		return "", fmt.Errorf("default similarity search failed: %w", err)
	}

	// Combine results
	log.Debug().Msgf("Combining results")
	similarDocs := append(orgDocs, defaultDocs...)
	if len(similarDocs) == 0 {
		log.Warn().Msg("No relevant documents found in either namespace. Proceeding with empty context.")
		similarDocs = nil
	}

	// Log retrieved documents
	log.Debug().Msgf("Retrieved documents for thread %s: %+v", threadID, similarDocs)

	// Combine documents for context
	var docContext bytes.Buffer
	if len(similarDocs) > 0 {
		for _, doc := range similarDocs {
			docContext.WriteString(fmt.Sprintf("Document: %s\nMetadata: %+v\n", doc.PageContent, doc.Metadata))
		}
	} else {
		docContext.WriteString("No relevant documents found.\n")
	}

	// Prepare LLM input context
	history, _ := threadMemory.ChatHistory.Messages(ctx)
	chainInputs := map[string]any{
		"context": fmt.Sprintf(
			"History:\n%s\n\nRelevant Documents:\n%s\n\nUser Input:\n%s",
			formatMessages(history), docContext.String(), input,
		),
	}
	log.Debug().Msgf("LLM context:\n%s", chainInputs["context"])

	// Call LLM chain
	chainOutputs, err := chains.Call(ctx, am.LLMChain, chainInputs, chains.WithStreamingFunc(
		func(ctx context.Context, chunk []byte) error {
			if chunkCallback != nil {
				chunkCallback(chunk)
			}
			return nil
		},
	))
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute LLMChain.")
		return "", fmt.Errorf("failed to execute LLMChain: %w", err)
	}

	// Parse response and store in memory
	fullResponse := chainOutputs["text"].(string)
	log.Debug().Msgf("Response: %s", fullResponse)
	threadMemory.SaveContext(ctx,
		map[string]any{"input": input},
		map[string]any{"response": fullResponse},
	)

	// Pass userID and orgID to addToBuffer
	am.addToBuffer(threadID, input, fullResponse, userID, orgID)

	return fullResponse, nil
}
