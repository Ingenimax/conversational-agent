package agents

import (
	"context"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
)

// InitializeLLM sets up the OpenAI LLM.
func InitializeLLM(openAIApiKey, openAIModel string) (*openai.LLM, error) {
	return openai.New(
		openai.WithToken(openAIApiKey),
		openai.WithModel(openAIModel),
	)
}

// InitializeMemory sets up conversation buffer memory.
func InitializeMemory() *memory.ConversationBuffer {
	return memory.NewConversationBuffer(
		memory.WithMemoryKey("history"),
		memory.WithReturnMessages(false), // Return buffer as a string
		memory.WithHumanPrefix("You"),
		memory.WithAIPrefix("StarOps"),
	)
}

func InitializeChain(llm llms.Model, memory *memory.ConversationBuffer) (*chains.LLMChain, error) {
	// Define the prompt template with a single 'context' field
	prompt := prompts.NewPromptTemplate(`
{{.context}}

AI Response:`,
		[]string{"context"}, // Explicitly define 'context' as the input key
	)
	// Create the LLMChain with the prompt and memory
	chain := chains.NewLLMChain(llm, prompt)
	chain.Memory = memory

	return chain, nil
}

// InitializeEmbedder sets up the OpenAI embedder.
func InitializeEmbedder(llm *openai.LLM) (*embeddings.EmbedderImpl, error) {
	// Wrap LLM's CreateEmbedding method in EmbedderClientFunc
	client := embeddings.EmbedderClientFunc(func(ctx context.Context, texts []string) ([][]float32, error) {
		return llm.CreateEmbedding(ctx, texts)
	})

	return embeddings.NewEmbedder(client)
}

// InitializeVectorStore sets up the Weaviate vector store.
func InitializeVectorStore(
	weaviateHost, weaviateApiKey, weaviateIndex string,
	embedder *embeddings.EmbedderImpl,
) (weaviate.Store, error) {
	return weaviate.New(
		weaviate.WithHost(weaviateHost),
		weaviate.WithScheme("https"),
		weaviate.WithAPIKey(weaviateApiKey),
		weaviate.WithIndexName(weaviateIndex),
		weaviate.WithTextKey("text"),
		weaviate.WithEmbedder(embedder),
	)
}
