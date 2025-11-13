````markdown
---
name: "Done Journal Development Guide"
description: "Development guide for Done Journal: a Telegram-based daily accomplishments logger with a Go + SQLite backend and React MiniApp frontend."
category: "Fullstack App"
author: "Done Journal"
authorUrl: "https://github.com/your-org-or-user"
tags:
  [
    "go",
    "sqlite",
    "connectrpc",
    "telegram-bot",
    "react",
    "vite",
    "miniapp",
    "connectrpc-web",
    "tanstack-router",
    "tanstack-react-query",
    "shadcn-ui",
  ]
lastUpdated: "2025-11-13"
---

# Done Journal Development Guide

## Project Overview

Done Journal is a small productivity tool built around one simple question:

> “А чо я делал вчера на работе?” / “What did I even do yesterday at work?”

The system helps users:

- log what they **actually** did during the day,
- manage simple TODOs,
- get quick summaries for daily standups (yesterday/today),
- use a Telegram bot and MiniApp as the main interface.

This guide is for **coding agents / IDE assistants** working inside this repository:

- how the system is structured,
- how to build and run it,
- how to test it,
- what tech stack and conventions it uses.

Humans should refer to `README.md` for the high-level overview.

---

## Architecture Overview

High-level components:

1. **Telegram Bot**

   - Handles user interaction in Telegram (commands, inline buttons, messages).
   - Sends structured requests to the backend via HTTP/ConnectRPC.
   - Main flows:
     - Add “done” entry (something user already did).
     - Add TODO (future task).
     - Show “yesterday” / “today” / “day X” summary.

2. **Backend (Go + SQLite + ConnectRPC)**

   - Single service (at least in v1) exposing a ConnectRPC API.
   - Responsibilities:
     - Store “done” entries.
     - Store TODO items.
     - Generate summaries per user + date.
     - Handle simple business rules (ownership, basic validation).
   - Uses SQLite as the primary storage.

3. **MiniApp Frontend (React + Vite)**

   - Telegram MiniApp served as a web UI.
   - Responsibilities:
     - Provide interactive view for:
       - planned TODOs,
       - completed tasks,
       - daily/weekly summary.
     - Allow editing/planning tasks per day.
   - Communicates with backend through `connectrpc-web` and `@connectrpc/connect-query`.

4. **Database (SQLite)**
   - Stores:
     - users (mapped to Telegram user IDs),
     - done entries,
     - TODO items,
     - possibly simple metadata.

Basic flow example (user logging work):

1. User sends a message to the bot: “fixed login bug, reviewed PR #42”.
2. Bot sends a request to backend: `POST /rpc.donejournal.LogService/CreateDoneEntry`.
3. Backend writes the entry into SQLite.
4. The next day, user asks “what did I do yesterday?” → bot requests summary from backend.
5. MiniApp can show the same data visually.

---

## Tech Stack

### Backend

- **Language:** Go
- **Storage:** SQLite
- **API:** ConnectRPC (gRPC over HTTP/JSON)
- **Key responsibilities:**
  - Logging completed work (“done entries”)
  - Managing TODO items
  - Generating per-user per-day summaries

### Frontend (Telegram MiniApp)

- **Bundler:** Vite
- **Framework:** React
- **RPC client:** `connectrpc-web`, `@connectrpc/connect-query`
- **Routing:** `@tanstack/router`
- **Forms:** `@tanstack/react-form`
- **Validation:** `zod`
- **Data fetching / cache:** `@tanstack/react-query`
- **UI components:** `shadcn/ui`

### Bot Integration

- Telegram Bot API (exact library may vary, e.g. `go-telegram-bot-api` or similar).
- Bot acts as a thin client over the backend API.

---

## Repository Structure (expected)

> ⚠️ If the actual structure differs, keep this as a reference for where new code should go.

```text
.
├── backend/
│   ├── cmd/
│   │   └── server/           # main backend entrypoint
│   ├── internal/
│   │   ├── app/              # application wiring, services
│   │   ├── domain/           # domain models (log entries, todos, users)
│   │   ├── persistence/      # SQLite repositories
│   │   ├── transport/        # ConnectRPC / HTTP handlers
│   │   └── config/           # config loading
│   ├── proto/                # .proto definitions for ConnectRPC
│   ├── migrations/           # SQL migrations for SQLite
│   ├── go.mod
│   └── go.sum
├── bot/
│   └── ...                   # Telegram bot implementation
├── web/
│   ├── src/
│   │   ├── app/              # App shell, routes
│   │   ├── features/         # Feature modules (done, todo, summary)
│   │   ├── components/       # Shared UI components
│   │   ├── lib/              # helpers/hooks (connect-query, forms)
│   │   └── types/            # shared front-end types/zod schemas
│   ├── index.html
│   ├── vite.config.ts
│   ├── package.json
│   └── pnpm-lock.yaml | package-lock.json | yarn.lock
├── deployments/
│   └── ...                   # docker-compose, k8s manifests, etc.
├── AGENTS.md                 # this file
└── README.md                 # human-facing project README
```
````

---

## Development Environment Setup

### Prerequisites

- **Go** (version matching `go.mod`)
- **Node.js + npm / pnpm / yarn** (for the MiniApp)
- **SQLite** (CLI optional but useful)
- **Protocol Buffers / Connect tooling** (if proto files are used directly)
- **Telegram Bot Token** (for running the bot)

---

### Backend Setup

From the project root:

```bash
cd backend

# Download Go dependencies
go mod tidy
```

If there are SQL migrations:

```bash
# Example: using a simple migration tool or custom script
go run ./cmd/migrate
```

To run the backend server (dev mode):

```bash
go run ./cmd/server
```

Typical backend responsibilities:

- serve ConnectRPC endpoints on e.g. `http://localhost:8080`
- connect to local `donejournal.db` (SQLite file)
- log requests and errors to stdout

---

### Frontend (MiniApp) Setup

```bash
cd web

# Install dependencies
npm install
# or
pnpm install
# or
yarn
```

Run development server:

```bash
npm run dev
# or pnpm dev / yarn dev
```

The MiniApp will be available at something like `http://localhost:5173` and integrated into Telegram via MiniApp URL in bot settings (production/staging).

---

### Telegram Bot Setup

1. Create a bot via [@BotFather](https://t.me/BotFather).
2. Get the **bot token**.
3. Configure environment variables (example):

```bash
# .env (not committed)
TELEGRAM_BOT_TOKEN=123456:abcdef-...
BACKEND_BASE_URL=http://localhost:8080
```

4. Run the bot:

```bash
cd bot
go run ./cmd/bot
# or whichever entrypoint is used
```

The bot should:

- respond to basic commands (`/start`, `/help`, maybe `/yesterday`, `/today`)
- send requests to the backend for actual data.

---

## Running the Full Stack (Dev)

Typical dev sequence:

```bash
# Terminal 1 – backend
cd backend
go run ./cmd/server

# Terminal 2 – web (MiniApp)
cd web
npm run dev

# Terminal 3 – bot
cd bot
go run ./cmd/bot
```

Once all are running, you can:

- Talk to the bot in Telegram.
- Use the MiniApp for rich UI flows.
- Watch data being persisted in SQLite.

---

## Testing

### Backend Testing (Go)

```bash
cd backend

# Run all tests
go test ./...

# Optional: with race detector
go test -race ./...

# Optional: lint (if configured)
golangci-lint run
```

Tests should typically cover:

- domain logic (done entries, TODOs, summaries),
- repository implementations (with SQLite, possibly using tmp DB),
- transport layer (ConnectRPC handlers).

---

### Frontend Testing / Quality

```bash
cd web

# Typecheck (if TypeScript)
npm run typecheck

# Lint (if configured)
npm run lint

# Unit tests (if configured)
npm test
```

Focus points:

- zod validation working as expected,
- forms using `@tanstack/react-form`,
- RPC hooks using `@connectrpc/connect-query`,
- UI behaviour consistent with backend contracts.

---

## Conventions & Notes for Coding Agents

- Prefer small, focused changes rather than huge rewrites.
- Keep **backend, bot, and frontend contracts in sync**:

  - If you change a proto → update Go server → update frontend hooks.

- Always validate data:

  - Backend: type-safe & basic validation.
  - Frontend: zod schemas + React forms.

- Use existing patterns:

  - For HTTP/RPC handlers, mimic existing handlers.
  - For UI, reuse Shadcn components and shared form patterns.

- Don’t introduce additional tech/libraries without necessity.

The main goal of Done Journal is to make answering
**“what did I do yesterday at work?”**
a one-tap operation — keep the codebase simple enough to support that.
