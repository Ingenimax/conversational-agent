# ![Ingenimax](/img/logo-header.svg#gh-light-mode-only) ![Ingenimax](/img/logo-header-inverted.svg#gh-dark-mode-only)

# Conversational Agent

A robust conversational AI system built in Go that combines vector search capabilities with LLM-powered responses. The system supports organization-specific knowledge bases, conversation memory, and real-time streaming responses.

## Features

- **Dual-namespace Vector Search**: Searches both organization-specific and default knowledge bases
- **Conversation Memory**: Maintains thread-based conversation history
- **Streaming Support**: Real-time streaming of AI responses
- **Document Management**: Import and manage knowledge base documents
- **Organization Isolation**: Separate vector spaces for different organizations
- **Robust Error Handling**: Graceful error handling and retries
- **Efficient Memory Management**: Buffered message storage with automatic flushing

## Architecture

- **Agent Manager**: Core component managing LLM interactions and vector store operations
- **Vector Store**: Weaviate-based vector database for efficient similarity search
- **Memory System**: Thread-based conversation buffer with persistence
- **REST API**: Echo-based HTTP server with streaming support

## API Endpoints

- `POST /v1/agent/query/:org_id/:user_id/:thread_id` - Submit queries and receive responses
- `GET /v1/agent/memory/thread/:thread_id` - Retrieve conversation history
- `POST /v1/agent/memory/update` - Add documents to the knowledge base
- `POST /v1/agent/memory/import/:org_id/:user_id` - Bulk import documents

## Setup

1. Clone the repository
2. Create a `.env` file with the following variables:

3. Install dependencies:

```bash
go mod download
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

### Add Documents

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

## Dependencies

- [Echo](https://echo.labstack.com/) - Web framework
- [LangChain Go](https://github.com/tmc/langchaingo) - LLM framework
- [Weaviate](https://weaviate.io/) - Vector database
- [Zerolog](https://github.com/rs/zerolog) - Logging
- [Viper](https://github.com/spf13/viper) - Configuration management

## License

[Your chosen license]

## Contributing

[Your contribution guidelines]
