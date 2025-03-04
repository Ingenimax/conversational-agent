# ![Ingenimax](/img/logo-header.png#gh-light-mode-only) ![Ingenimax](/img/logo-header-inverted.png#gh-dark-mode-only)

# Conversational Agent

A production-grade conversational AI agent built in Go that leverages Weaviate for vector search and OpenAI’s language models. Designed for enterprise applications, this multi-tenant agent provides tenant-specific knowledge retrieval, persistent conversation memory, and real-time response streaming.

Ideal for customer support automation, knowledge management, and enterprise AI assistants, it enables seamless, context-aware conversations tailored to specific organizational needs.

## Features

- **Multi-Namespace Vector Search**: Searches both organization-specific and general knowledge bases to provide accurate and context-aware responses.
- **Conversation Memory**: Maintains thread-based conversation history, ensuring continuity across user interactions and allowing persistent storage for long-term learning.
- **Streaming Support**: Supports real-time streaming of AI-generated responses for a more interactive and responsive experience.
- **Document Management**: Allows users to import and manage knowledge base documents with metadata, improving the AI agent’s contextual understanding.
- **Organization Isolation**: Ensures tenant-specific knowledge separation, preventing data leakage between organizations or users.
- **Buffered Message Storage**: Optimized memory management through automatic flushing to maintain performance and reduce system overhead.
- **Metadata Filtering**: Enables fine-grained document queries using filters for user and organization IDs, improving retrieval accuracy.

## Architecture

- **Agent Manager**: The central logic engine that handles LLM interactions, query processing, and vector store operations.
- **Vector Store**: A high-performance vector store (Weaviate) that enables efficient semantic search across organization-specific and general knowledge bases.
- **Memory System**: Implements thread-based conversation persistence, storing previous exchanges to maintain conversational context.
- **REST API**: A Go-based HTTP server that handles queries, document ingestion, and conversation retrieval, with support for response streaming.

### Diagram
![Diagram](/img/architecture_light.png#gh-light-mode-only) ![Diagram](/img/architecture_dark.png#gh-dark-mode-only)

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/agent/query/:org_id/:user_id/:thread_id` | Submit a query and retrieve an AI-generated response. |
| `GET`  | `/v1/agent/memory/thread/:thread_id` | Retrieve a conversation's memory for a specific thread. |
| `POST` | `/v1/agent/memory/update` | Add new knowledge base documents. |
| `POST` | `/v1/agent/memory/import/:org_id/:user_id` | Bulk import documents for an organization and user. |

## Installation

1. Clone the repository
    ```
    https://github.com/Ingenimax/conversational-agent.git
    cd conversational-agent
    ```

2. Create a `.env` file with the following variables:
   ```
   OPENAI_API_KEY=your_openai_api_key
   OPENAI_MODEL=gpt-4o-mini
   WEAVIATE_HOST=https://your_weaviate_host
   WEAVIATE_API_KEY=your_weaviate_api_key
   WEAVIATE_INDEX_NAME=your_index_name
   DEBUG=true
   ```

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Run the server:
   ```bash
   go run cmd/main.go
   ```

## Usage

### Query the Agent

```bash
curl -X POST "http://localhost:8080/v1/agent/query/:org_id/:user_id/:thread_id" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Your question here",
    "stream": false
  }'
```

For streaming responses, set `"stream": true` in the request body.

### Retrieve Conversation History

```bash
curl -X GET "http://localhost:8080/v1/agent/memory/thread/:thread_id"
```

### Add Document to Knowledge Base

```bash
curl -X POST "http://localhost:8080/v1/agent/memory/update" \
  -H "Content-Type: application/json" \
  -d '{
    "page_content": "Your document content",
    "metadata": {
      "source": "example",
      "author": "John Doe"
    }
  }'
```

### Import Dataset

```bash
curl -X POST "http://localhost:8080/v1/agent/memory/import/:org_id/:user_id" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/dataset.json"
  }'
```

## Key Components

- **Agent Manager**: Manages LLM interactions, vector store operations, and memory
- **Vector Store**: Weaviate storage for document embeddings with namespace support
- **Memory System**: Thread-based conversation buffer with persistence
- **Query Processing**: Combines organization-specific and default knowledge bases for comprehensive responses

## Dependencies

- [Echo](https://echo.labstack.com/) - Web framework
- [LangChain Go](https://github.com/tmc/langchaingo) - LLM framework
- [Weaviate](https://weaviate.io/) - Vector database
- [Zerolog](https://github.com/rs/zerolog) - Logging
- [Viper](https://github.com/spf13/viper) - Configuration management
- [OpenAI API](https://openai.com/) - LLM and embedding provider

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
