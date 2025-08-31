
# AI Contextual Prompter

**AI Contextual Prompter** is a modular Go library for building context-aware LLM applications. It provides:
- Pluggable vector database support (in-memory, PostgreSQL/pgvector, etc.)
- Pluggable LLM provider support (OpenAI, Claude, Gemini, etc.)
- A unified interface for adding, searching, and using context with LLMs

## Features

- Add and store context with embeddings
- Retrieve most relevant context for a query
- Query LLMs with context-aware prompts
- Easily swap vector DB or LLM provider


## Architecture

```
┌────────────┐      ┌───────────────┐      ┌─────────────┐
│  Your App  │ <--> │ContextPrompter│ <--> │  Vector DB  │
└────────────┘      └───────────────┘      └──────┬──────┘
                        ▲                         │
                        │                         │
                        │                  ┌──────▼─────┐
                        │                  │    LLM     │
                        └─────────────────►│ (Embeddings│
                                           │   & Prompt)│
                                           └────────────┘
```
*LLM is used both for generating embeddings (for context storage/search) and for answering prompts with context.*

---

## LLMs

The library supports pluggable LLM providers via a common interface. You can add your own or use the built-ins.

### OpenAI

**Modes:**

- **Normal (Chat/Completion) Mode:**
  - Uses models like `gpt-3.5-turbo` or `gpt-4`.
  - Prompts are sent directly to the chat/completions API.
  - Good for most use-cases and simple context injection.

- **Assistant Mode:**
  - Uses OpenAI's Assistant API (with an Assistant ID).
  - Supports advanced workflows, threads, and persistent conversations.
  - Context is added as messages to a thread, and the assistant manages state.

**Embedding:**

- The OpenAI LLM is also used to generate vector embeddings for your context using the `/embeddings` API (e.g., `text-embedding-ada-002`).
- This allows the vector DB to store and search context semantically.

---

## Vector DBs

The library supports pluggable vector database backends:

### PostgreSQL (pgvector)

- Stores embeddings in a Postgres table with the [pgvector](https://github.com/pgvector/pgvector) extension.
- Supports fast similarity search, persistence, and scaling.
- Recommended for production and large datasets.

### In-Memory (local)

- Stores embeddings in a Go map in memory.
- Fast and simple, but **not persistent**—all data is lost on restart.
- Great for prototyping, testing, or small-scale use.

---

## Quick Start

### 1. Install

Clone the repo and run:

```sh
go mod tidy
```

### 2. Example: Using OpenAI + Postgres Vector DB

```go
package main

import (
    "context"
    "log"
    "github.com/shreetheja/ai-contextual-prompter/context-prompter"
    "github.com/shreetheja/ai-contextual-prompter/llm-providers/openai"
    "github.com/shreetheja/ai-contextual-prompter/vector-db/pgsql-vec"
)

func main() {
    // Fill these configs from env or file
    openaiKey := "YOUR_OPENAI_KEY"
    openaiOrg := "YOUR_OPENAI_ORG"
    var openaiAsst *string = nil // or pointer to assistant id

    pgCfg := pgsqlvec.Config{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "password",
        Database: "vectordb",
        Table:    "embeddings",
        Col:      "vec",
        IdColName: "id",
    }

    llm := openai.NewClient(openaiKey, openaiOrg, openaiAsst)
    vecDB, err := pgsqlvec.NewEntity(pgCfg)
    if err != nil {
        log.Fatalf("failed to init pgsql vector db: %v", err)
    }

    prompter := context_prompter.Prompter{
        LLM:      llm,
        VectorDB: vecDB,
        MaxContext: 5,
    }

    ctx := context.Background()
    // Add context
    err = prompter.AddContext(ctx, "The Eiffel Tower is in Paris.", map[string]interface{}{"text": "The Eiffel Tower is in Paris."})
    if err != nil {
        log.Fatal(err)
    }
    // Query with context
    resp, err := prompter.Query(ctx, "Where is the Eiffel Tower?", 3)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("LLM Response:", resp)
}
```

### 3. In-Memory Vector DB Example

```go
import (
    "github.com/shreetheja/ai-contextual-prompter/vector-db/local"
)
vecDB := local.NewInMemoryVectorDB()
```

## Extending

- Implement the `VectorDB` or `LLM` interface for new backends/providers.
- Use the factory methods for easy swapping.

## Contributing

Pull requests and issues are welcome! Please open an issue to discuss major changes.

## License

MIT