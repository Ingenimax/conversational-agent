package agents

import (
	"fmt"
	"sync"

	"github.com/blog/conversational-agent/internal/logger"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
)

type AgentManager struct {
	LLM                 *openai.LLM
	VectorStore         weaviate.Store
	AgentMemory         map[string]*memory.ConversationBuffer
	ConversationalChain *chains.ConversationalRetrievalQA
	WeaviateIndex       string
	LLMChain            *chains.LLMChain
	messageBuffer       []schema.Document
	bufferMutex         sync.Mutex
	maxBufferMessages   int
	memoryMutex         sync.Mutex
}

func NewAgentManager(
	openAIApiKey, openAIModel, weaviateHost, weaviateApiKey, weaviateIndex string,
	maxBufferMessages int,
) (*AgentManager, error) {
	log := logger.GetLogger()

	log.Info().Msgf("Initializing OpenAI LLM with Model: %s", openAIModel)
	llm, err := InitializeLLM(openAIApiKey, openAIModel)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAI LLM: %w", err)
	}
	log.Info().Msg("OpenAI LLM initialized successfully")

	log.Info().Msg("Initializing OpenAI Embedder...")
	embedder, err := InitializeEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAI embedder: %w", err)
	}
	log.Info().Msg("OpenAI Embedder initialized successfully")

	log.Info().Msg("Initializing Weaviate vector store...")
	vectorStore, err := InitializeVectorStore(weaviateHost, weaviateApiKey, weaviateIndex, embedder)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Weaviate vector store: %w", err)
	}

	log.Info().Msg("Weaviate vector store initialized successfully")

	log.Info().Msg("Initializing ConversationBuffer memory...")
	agentMemory := InitializeMemory()
	log.Info().Msg("ConversationBuffer memory initialized successfully")

	log.Info().Msg("Initializing Chain...")
	chain, err := InitializeChain(llm, agentMemory)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chain: %w", err)
	}
	log.Info().Msg("Chain initialized successfully")

	return &AgentManager{
		LLM:               llm,
		VectorStore:       vectorStore,
		AgentMemory:       make(map[string]*memory.ConversationBuffer),
		LLMChain:          chain,
		WeaviateIndex:     weaviateIndex,
		messageBuffer:     []schema.Document{},
		maxBufferMessages: maxBufferMessages,
	}, nil
}
