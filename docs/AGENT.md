# AI Agent Architecture

## Обзор

DoneJournal Agent — AI-ассистент для управления задачами и заметками через Telegram-бота. Использует паттерн **Tool-Use / Function Calling** с поддержкой многоитерационного цикла, историей диалога и асинхронной обработкой через очередь задач.

**Ключевые технологии:**

- **LLM Provider**: Groq API (OpenAI-совместимый формат)
- **Conversation Memory**: BadgerDB (key-value store)
- **Job Queue**: River (SQLite-backed)
- **Telegram Bot**: Telego

---

## Архитектура

### Общая схема обработки сообщения

```mermaid
flowchart TD
    User([Telegram User]) -->|сообщение / голос / кнопка| Bot[Telego Bot]
    Bot -->|update channel| TM[TManager.handleUpdate]

    TM -->|voice| VW[Voice Worker]
    TM -->|text| PW[Processor Worker]
    TM -->|command| CMD[Processor.HandleCommand]
    TM -->|callback| CB[Processor.HandleCallbackQuery]

    VW -->|STT transcribe| STT[STT Service]
    STT -->|text| PW

    PW -->|typing indicator| Bot
    PW -->|ProcessNewRequest| Proc[Processor]

    Proc -->|inbox keyword?| Inbox[(Inbox DB)]
    Proc -->|delegate| Agent[Agent.Process]

    Agent -->|load history| BDB[(BadgerDB)]
    Agent -->|agent loop| Groq[Groq API]
    Groq -->|tool_calls| Exec[Executor]
    Exec -->|CRUD| Services[(SQLite DB)]
    Exec -->|result| Agent
    Agent -->|save history| BDB

    Agent -->|response| SMW[Send Message Worker]
    SMW -->|SendMessage| Bot
    Bot -->|ответ| User

    style Agent fill:#e1f5fe
    style Groq fill:#fff3e0
    style Exec fill:#f3e5f5
```

### Agent Loop (цикл tool-use)

```mermaid
flowchart TD
    Start([Process вызван]) --> Load[Загрузить историю из BadgerDB]
    Load --> LoadWS[Загрузить воркспейсы пользователя]
    LoadWS --> Build[Собрать messages:<br/>system + history + user]
    Build --> Loop{Итерация < 10?}

    Loop -->|да| LLM[Запрос к Groq API<br/>messages + tools]
    LLM --> Check{Есть tool_calls?}

    Check -->|нет| Final[Финальный ответ]
    Check -->|да| ExecTools[Выполнить каждый tool call<br/>через Executor]
    ExecTools --> Append[Добавить assistant + tool results<br/>в messages]
    Append --> Loop

    Loop -->|нет, лимит| Fallback[Ответ: 'Не удалось обработать']

    Final --> Save[Сохранить историю в BadgerDB]
    Save --> Return([Вернуть текст])
    Fallback --> Save

    style LLM fill:#fff3e0
    style ExecTools fill:#f3e5f5
    style Final fill:#e8f5e9
```

### Очередь задач (River)

```mermaid
flowchart LR
    subgraph Workers["River Workers (max 100)"]
        VoiceW[Voice Worker<br/>timeout: 5 min]
        ProcW[Processor Worker<br/>timeout: 120s]
        SendW[Send Message Worker<br/>timeout: 30s]
    end

    TG([Telegram Update]) --> VoiceW
    TG --> ProcW
    VoiceW -->|transcribed text| ProcW
    ProcW -->|response| SendW
    SendW -->|SendMessage| TG2([Telegram API])

    style VoiceW fill:#e3f2fd
    style ProcW fill:#fff3e0
    style SendW fill:#e8f5e9
```

---

## Структура файлов

```text
internal/agent/
├── agent.go           # Agent struct, Process(), system prompt, ClearConversation()
├── conversation.go    # ConversationStore (BadgerDB), Load/Save/Clear
├── executor.go        # Executor — диспетчер tool calls → сервисы
├── tools.go           # Определения всех tools (JSON Schema)
└── provider/
    ├── types.go       # Provider interface, ChatMessage, ToolCall, ToolDefinition
    └── groq/
        └── groq.go    # Groq API client (OpenAI-compatible)
```

---

## Tools (инструменты агента)

| Tool              | Описание                         | Ключевые параметры                                     |
| ----------------- | -------------------------------- | ------------------------------------------------------ |
| `create_todo`     | Создать задачу                   | `title`\*, `status`, `planned_date`, `workspace`       |
| `create_note`     | Создать заметку                  | `title`\*, `body`, `workspace`                         |
| `find_todos`      | Найти задачи по фильтрам         | `status[]`, `date_from`, `date_to`, `workspace`        |
| `find_notes`      | Найти заметки                    | `search`, `workspace`                                  |
| `complete_todo`   | Отметить задачу выполненной      | `id`\*                                                 |
| `update_todo`     | Изменить задачу                  | `id`\*, `title`, `status`, `planned_date`, `workspace` |
| `delete_todo`     | Удалить задачу                   | `id`\*                                                 |
| `update_note`     | Изменить заметку                 | `id`\*, `title`, `body`, `workspace`                   |
| `delete_note`     | Удалить заметку                  | `id`\*                                                 |
| `list_workspaces` | Список воркспейсов               | —                                                      |
| `save_to_inbox`   | Сохранить в inbox                | `text`\*                                               |
| `get_todo_stats`  | Статистика по задачам            | `date_from`, `date_to`, `workspace`                    |
| `list_inbox`      | Список inbox items               | —                                                      |
| `convert_inbox`   | Конвертировать inbox → todo/note | `inbox_id`_, `convert_to`_                             |

\* — обязательный параметр

---

## Как добавить новый tool

### Шаг 1: Определение tool в `tools.go`

Добавить `provider.ToolDefinition` в функцию `toolDefinitions()`:

```go
{
    Name:        "my_new_tool",
    Description: "Описание для LLM — что делает этот инструмент и когда его использовать",
    Parameters: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "param1": map[string]any{
                "type":        "string",
                "description": "Описание параметра",
            },
            "param2": map[string]any{
                "type":        "integer",
                "description": "Описание параметра",
            },
        },
        "required": []string{"param1"},
    },
},
```

**Рекомендации:**

- `Description` — пишите подробно, LLM использует его для выбора инструмента
- `required` — указывайте обязательные параметры
- Типы: `string`, `integer`, `boolean`, `array`, `object`
- Для enum: `"enum": []string{"value1", "value2"}`

### Шаг 2: Реализация в `executor.go`

1. Добавить case в `Execute()`:

```go
case "my_new_tool":
    return e.myNewTool(ctx, userID, args)
```

2. Создать метод:

```go
func (e *Executor) myNewTool(ctx context.Context, userID int64, args map[string]any) (string, error) {
    // 1. Извлечь параметры
    param1, _ := args["param1"].(string)
    if param1 == "" {
        return "", fmt.Errorf("param1 is required")
    }

    // 2. Вызвать сервис
    result, err := e.services.MyService().DoSomething(ctx, param1)
    if err != nil {
        return "", fmt.Errorf("failed to do something: %w", err)
    }

    // 3. Вернуть JSON
    response := map[string]any{
        "id":     result.ID,
        "status": "ok",
    }
    data, _ := json.Marshal(response)
    return string(data), nil
}
```

### Шаг 3: Обновить system prompt (при необходимости)

Если tool требует специального поведения, добавить правило в `buildSystemPrompt()` в `agent.go`:

```go
- "keyword"/"ключевое слово" → my_new_tool
```

### Шаг 4: Проверить

```bash
go build ./...
```

Отправить сообщение боту, которое должно триггерить новый tool. Проверить логи:

```text
agent tool calls  iteration=1  tool_count=1
```

---

## Conversation Memory

### Как работает

```mermaid
flowchart LR
    subgraph BadgerDB
        K1["conv:12345"] -->|JSON| V1["[msg1, msg2, ..., msgN]"]
        K2["conv:67890"] -->|JSON| V2["[msg1, msg2, ...]"]
    end

    Agent -->|Load| BadgerDB
    Agent -->|Save last 50 msgs| BadgerDB
    Clear["/clear command"] -->|Delete key| BadgerDB
```

- **Ключ**: `conv:{userID}` (например, `conv:46472`)
- **Значение**: JSON-массив `ChatMessage` (role, content, tool_calls, tool_call_id)
- **Лимит**: последние 50 сообщений
- **System prompt** не сохраняется — генерируется при каждом запросе

### Структура сообщения

```go
type ChatMessage struct {
    Role       string     // "user", "assistant", "tool"
    Content    string     // Текст сообщения
    ToolCalls  []ToolCall // Вызовы инструментов (только для assistant)
    ToolCallID string     // ID вызова (только для tool results)
}
```

---

## Provider (LLM интеграция)

### Интерфейс

```go
type Provider interface {
    ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
```

### Как добавить нового провайдера

1. Создать пакет `internal/agent/provider/newprovider/`
2. Реализовать `provider.Provider` интерфейс
3. Подключить в `cmd/donejournal/start.go`

### Текущий провайдер: Groq

| Параметр     | Значение                                          |
| ------------ | ------------------------------------------------- |
| API URL      | `https://api.groq.com/openai/v1/chat/completions` |
| Формат       | OpenAI-compatible                                 |
| Temperature  | 0.1                                               |
| Max Tokens   | 4096                                              |
| HTTP Timeout | 120s                                              |

---

## Обработка ошибок

```mermaid
flowchart TD
    UserMsg([Сообщение пользователя]) --> Agent[Agent.Process]

    Agent -->|ошибка LLM| InboxFallback[Сохранить в Inbox]
    InboxFallback --> Reply1["'Saved to inbox 📥'"]

    Agent -->|ошибка tool| ToolError["JSON: {error: '...'}"]
    ToolError --> NextIter[LLM получает ошибку<br/>и пробует другой подход]

    Agent -->|max iterations| Fallback["'Не удалось обработать'"]

    Agent -->|успех| Response[Ответ пользователю]
    Response --> SendMsg[SendMessage<br/>Markdown → plain text fallback]

    style InboxFallback fill:#fff3e0
    style ToolError fill:#ffebee
    style Fallback fill:#ffebee
    style Response fill:#e8f5e9
```

- **LLM недоступен**: сообщение сохраняется в Inbox как fallback
- **Tool execution error**: ошибка передаётся LLM в формате JSON, он может попробовать снова
- **Max iterations**: возвращается сообщение об ошибке
- **Telegram parse error**: автоматический fallback с Markdown на plain text

---

## Константы и лимиты

| Константа                   | Значение   | Где                      |
| --------------------------- | ---------- | ------------------------ |
| `maxToolIterations`         | 10         | `agent.go`               |
| `maxConversationMessages`   | 50         | `conversation.go`        |
| `processorWorker.Timeout`   | 120s       | `worker_processor.go`    |
| `voiceWorker.Timeout`       | 5 min      | `worker_voice.go`        |
| `sendMessageWorker.Timeout` | 30s        | `worker_send_message.go` |
| `River MaxWorkers`          | 100        | `tmanager.go`            |
| `Groq Temperature`          | 0.1        | `groq.go`                |
| `Groq MaxTokens`            | 4096       | `groq.go`                |
| `Bot updates channel`       | 500 buffer | `bot.go`                 |

---

## Telegram Bot команды

| Команда           | Действие                       |
| ----------------- | ------------------------------ |
| `/start`, `/menu` | Главное меню с inline-кнопками |
| `/clear`          | Очистить историю диалога       |
| Текст             | Обработка через Agent          |
| Голосовое         | STT → Agent                    |

### Inline-кнопки (callback queries)

```text
Главное меню
├── 📝 Показать TODO
│   ├── Сегодня
│   ├── Завтра
│   └── « Назад
└── ✅ Показать выполненное
    ├── Сегодня
    ├── Вчера
    ├── Позавчера
    └── « Назад
```
