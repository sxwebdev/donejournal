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
```
