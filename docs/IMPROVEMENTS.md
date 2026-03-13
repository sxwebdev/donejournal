# DoneJournal — Предложения по улучшению

## Context

DoneJournal — productivity-приложение с Telegram-ботом для управления задачами и заметками. Стек: Go + Connect-RPC + SQLite (бэк), React 19 + TanStack + shadcn/Base UI (фронт).

**AI Agent**: полноценный tool-use агент (`internal/agent/`) с 14 инструментами, conversation memory (BadgerDB, 50 сообщений), multi-iteration loop (до 10 итераций), provider Groq API (Llama, temperature 0.1). Агент умеет создавать/искать/обновлять/удалять todos и notes, управлять workspaces, работать с inbox, получать статистику. Поддерживает русский и английский, относительные даты, голосовые сообщения (Whisper STT).

**Текущие фичи**: inbox, todos (статусы, даты, workspaces), notes (markdown), calendar с drag-and-drop, workspaces, real-time подписки (server streaming), Telegram-бот с AI-агентом и голосовым вводом.

---

## 1. AI-фичи (главный приоритет)

### 1.1 AI-чат в веб-UI

**Effort**: Medium-Large | **Impact**: Very High

Главная возможность — перенести AI-агента из Telegram в веб-интерфейс. Агент (`internal/agent/`) уже полностью реализован с tool-use циклом, conversation memory и 14 инструментами. Нужно только создать веб-обёртку.

- **Backend**: Новый `ChatService` proto с RPC `SendMessage(text) → ChatResponse(text, created_entities[])`. Внутри вызывает `agent.Process()`. Streaming вариант `StreamMessage` для потокового вывода ответа
- **Frontend**: Чат-панель (slide-over или отдельная страница `/chat`):
  - Поле ввода + история сообщений
  - Отображение действий агента (создал таск, нашёл заметки) как structured cards
  - Typing indicator пока агент работает (tool iterations)
  - Кнопка очистки истории (`ClearConversation`)
- **Conversation memory**: уже реализована в BadgerDB — нужно решить: общая история с Telegram или раздельная (разные ключи `conv:web:{userId}` vs `conv:tg:{userId}`)
- **Ключевые файлы**: `internal/agent/agent.go` (Process), `internal/agent/conversation.go`, новый `api/schema/donejournal/chat/v1/chat.proto`

### 1.2 Smart Quick Add с естественным языком

**Effort**: Medium | **Impact**: High

Превратить inbox quick-add в умную строку, использующую агента:

- Пользователь пишет: "закончить отчёт завтра для проекта Alpha"
- Вызов `agent.Process()` с одноразовым контекстом (без сохранения в conversation history)
- Inline preview результата: "Todo: закончить отчёт | Дата: 14 мар | Workspace: Alpha"
- Подтверждение или редактирование перед финальным сохранением
- Fallback в inbox-item при ошибке
- **Альтернатива**: облегчённый вариант — отдельный RPC `ParseText` без полного agent loop, только один вызов LLM с tool definitions для парсинга

### 1.3 AI-саммари и отчёты

**Effort**: Medium | **Impact**: High

Новый tool `generate_summary` для агента + RPC для веба:

- **Daily digest**: "Что я сделал сегодня?" — агент использует `get_todo_stats` + `find_todos(status=completed)` и генерирует текстовое резюме
- **Weekly report**: саммари за неделю для стендапов — markdown с группировкой по дням/воркспейсам
- **Standup generator**: формат "вчера сделал / сегодня планирую / блокеры"
- Кнопка "Generate Report" на дашборде и в календаре
- Агент уже умеет вызывать `get_todo_stats` и `find_todos` — нужен только правильный промпт

### 1.4 AI-планирование дня

**Effort**: Medium | **Impact**: High

Новый tool `plan_day` — агент анализирует:

- Незавершённые задачи (pending + in_progress)
- Просроченные задачи
- Паттерны из предыдущих дней (что обычно делается в этот день недели)
- Предлагает план на день с приоритизацией
- Пользователь может одобрить план → агент создаёт/перенесёт задачи

### 1.5 Умные напоминания и suggestions

**Effort**: Large | **Impact**: Medium

- **Overdue detection**: агент замечает просроченные задачи и спрашивает: "У тебя 3 просроченных задачи. Перенести на сегодня или отменить?"
- **Pattern recognition**: "Ты обычно делаешь standup каждый будний день. Создать на сегодня?"
- **Completion suggestions**: если задача в статусе in_progress >3 дней — "Задача X в работе уже 3 дня. Обновить статус?"
- Реализация: River cron job → вызывает агента с системным промптом "проанализируй задачи пользователя" → отправляет suggestion через Telegram или веб-уведомление

### 1.6 Мульти-провайдер LLM

**Effort**: Small-Medium | **Impact**: Medium

Интерфейс `provider.Provider` уже абстрагирован. Добавить:

- **OpenAI** (GPT-4o) — для пользователей с API ключом
- **Anthropic** (Claude) — через тот же OpenAI-compatible формат или нативный SDK
- **Ollama** (локальные модели) — для self-hosted без внешних API
- Выбор провайдера в конфиге или per-user настройка
- **Файлы**: `internal/agent/provider/openai/`, `internal/agent/provider/anthropic/`, `internal/agent/provider/ollama/`

### 1.7 Контекстные действия через AI

**Effort**: Medium | **Impact**: Medium

Добавить AI-действия на существующие UI-элементы:

- **На списке задач**: "AI: Приоритизировать задачи" — агент анализирует все pending задачи и предлагает порядок
- **На заметке**: "AI: Извлечь задачи из заметки" — парсит markdown заметки и создаёт todos
- **На календаре**: "AI: Оптимизировать расписание" — перераспределяет задачи по дням
- **На workspace**: "AI: Саммари проекта" — генерирует отчёт по всем задачам и заметкам workspace

### 1.8 Улучшения текущего агента

**Effort**: Small-Medium | **Impact**: Medium

- **Token counting**: проверка лимита контекста перед отправкой в Groq (сейчас нет)
- **Streaming responses**: SSE/streaming для агента вместо ожидания полного ответа
- **Tool confirmation UI**: для delete-операций показывать confirmation в Telegram inline-кнопками перед выполнением
- **Rate limiting**: защита от спама запросов к Groq API
- **Conversation summary**: при достижении лимита 50 сообщений — суммаризировать старые через LLM вместо обрезки
- **Приоритет в create_todo**: добавить параметр `priority` в tool definition когда будет реализован 3.1

---

## 2. UI/UX улучшения

### 2.1 Dashboard / Главная страница

**Effort**: Medium | **Impact**: High

Сейчас `/` редиректит на `/inbox`. Заменить на дашборд:

- Задачи на сегодня (pending + in_progress)
- Статистика: выполнено за неделю, streak
- Счётчик inbox-элементов
- Quick add bar (с AI-парсингом из 1.2)
- Последние заметки
- AI-виджет: "Спросить агента" (быстрый доступ к чату из 1.1)
- **Файл**: `routes/_authenticated/index.tsx`

### 2.2 Inline создание тасков

**Effort**: Small | **Impact**: Medium

Быстрая строка ввода вверху `todo-list.tsx` — текст + Enter = таск на сегодня. Диалог остаётся для полного редактирования (как в Todoist/Linear).

### 2.3 Command Palette (Cmd+K)

**Effort**: Medium | **Impact**: High

Глобальная палитра команд:

- Поиск по всем сущностям (todos, notes, inbox)
- Быстрая навигация (pages, workspaces)
- Действия (new todo, new note)
- AI-режим: ввод с `/` переключает на natural language → отправка агенту
- Горячие клавиши: `N` — новый таск, `I` — inbox, `/` — поиск

### 2.4 Недельный вид календаря

**Effort**: Medium | **Impact**: Medium

Текущий календарь — только месяц. Добавить переключатель Week/Month. Недельный вид показывает больше деталей на день. RPC `getCalendarEntries` уже поддерживает произвольные диапазоны дат.

### 2.5 Улучшенные Empty States

**Effort**: Small | **Impact**: Medium

Пустые состояния с иллюстрациями и подсказками:

- Inbox: "Inbox пуст! Записывайте мысли здесь или отправляйте через Telegram"
- Todos: "Нет задач на сегодня. Добавить?"
- Notes: "Пока нет заметок"

### 2.6 Mobile-оптимизация

**Effort**: Small-Medium | **Impact**: Medium

- Bottom navigation bar вместо sidebar на мобильных
- Swipe-жесты на тасках (вправо — выполнить, влево — удалить)
- Увеличенные touch-targets на ячейках календаря

### 2.7 Badge на Inbox в сайдбаре

**Effort**: Small | **Impact**: Medium

Показывать количество необработанных inbox-элементов рядом с "Inbox" в `app-sidebar.tsx`.

### 2.8 Toast-уведомления при real-time обновлениях

**Effort**: Small | **Impact**: Small

Сейчас подписки молча обновляют данные. Добавить toast через sonner: "Новый таск создан через Telegram: X".

---

## 3. Улучшения текущей реализации

### 3.1 Приоритеты задач

**Effort**: Small | **Impact**: High

Новое поле `priority` (NONE/LOW/MEDIUM/HIGH/URGENT):

- Миграция: `priority TEXT NOT NULL DEFAULT 'none'` в `todos`
- Proto enum + поля в `Todo`, `CreateTodoRequest`, `UpdateTodoRequest`
- Селектор приоритета в `todo-form.tsx`, индикатор в `todo-item.tsx`
- Новый параметр `priority` в tool `create_todo` и `update_todo` в `tools.go`
- Обновить system prompt в `agent.go` для распознавания приоритета

### 3.2 Повторяющиеся задачи

**Effort**: Large | **Impact**: High

- Поле `recurrence_rule` + `recurrence_parent_id` в таблице todos
- River job для генерации следующего occurrence при завершении
- Picker в `todo-form.tsx` (daily, weekly, monthly, custom)
- Новый tool `create_recurring_todo` в агенте
- System prompt: "каждый день", "еженедельно", "каждый понедельник"

### 3.3 Глобальный поиск

**Effort**: Medium | **Impact**: High

У notes уже есть `search`. Расширить на todos и inbox. Единая страница поиска или интеграция в Command Palette. Агент уже умеет искать через `find_todos` и `find_notes`.

### 3.4 Теги / Метки

**Effort**: Medium | **Impact**: Medium

Таблицы `tags`, `todo_tags`, `note_tags`. Цветные теги, фильтрация. Новые tools в агенте: `add_tag`, `find_by_tag`. LLM извлекает хештеги из сообщений.

### 3.5 Пакетное перенесение просроченных задач

**Effort**: Small | **Impact**: Medium

Кнопка "Перенести все на сегодня" в `overdue-banner.tsx`. Новый `BatchUpdateTodos` RPC. Также добавить tool `reschedule_overdue` в агента.

### 3.6 Markdown preview в описаниях тасков

**Effort**: Small | **Impact**: Medium

Рендерить `todo.description` как markdown в `todo-item.tsx` (переиспользовать паттерн из notes).

---

## 4. Улучшения карточек Todo (`todo-item.tsx`)

Текущая карточка: статус-иконка, title, description (2 строки), planned date, status badge, menu. Не отображается workspace (хотя `workspaceId` есть в модели), `createdAt`, `completedAt`. Нужна подготовка к будущим тегам и приоритетам.

### 4.1 Workspace badge

**Effort**: Small | **Impact**: High

- Маленький badge с именем workspace рядом с датой
- Resolve `workspaceId` → имя через кэш `listWorkspaces` query
- Стиль: `text-xs bg-muted px-1.5 py-0.5 rounded`
- Если workspace не задан — не показывать

### 4.2 Metadata row (рефакторинг нижней строки)

**Effort**: Small | **Impact**: Medium

Объединить дату, workspace и будущие метаданные в единую расширяемую строку:

```text
📅 14 мар · 🏷 Work · #frontend #urgent
```

Формат: `flex items-center gap-1.5 text-xs text-muted-foreground`. Это фундамент для будущих тегов и приоритетов — новые элементы просто добавляются в строку.

### 4.3 Relative dates + иконки

**Effort**: Small | **Impact**: Small-Medium

- "Сегодня", "Завтра", "Вчера" вместо полных дат
- Иконка `Calendar` перед датой
- Overdue: `AlertTriangle` + красный текст (частично есть)

### 4.4 Priority indicator (left border)

**Effort**: Small | **Impact**: High | **Зависимость**: после 3.1 Приоритеты задач

- Цветная полоска `border-l-4` слева от карточки
- Urgent: `border-red-500`, High: `border-orange-500`, Medium: `border-yellow-500`, Low/None: без индикатора
- Самый компактный вариант — не добавляет элементов в layout

### 4.5 Теги как chips

**Effort**: Medium | **Impact**: Medium | **Зависимость**: после 3.4 Теги/Метки

- Цветные chips в metadata row: `text-[10px] px-1 py-0 rounded-sm`
- Максимум 3 видимых + "+N" badge
- Клик по тегу → фильтрация списка

### 4.6 Completion info

**Effort**: Small | **Impact**: Small

- Для завершённых: "Выполнено 14 мар в 15:30" вместо planned date
- Приглушённая карточка: `opacity-60` или `bg-muted/30`
- Tooltip на иконке: "Completed 2 hours ago"

### 4.7 Quick actions на hover

**Effort**: Medium | **Impact**: Medium

- При hover показать иконки: Calendar (перенести), Check (завершить), Trash (удалить)
- Убирает необходимость открывать dropdown для частых действий
- На мобильном — swipe gestures

### Порядок реализации карточек

| #   | Улучшение       | Effort | Зависимости                     |
| --- | --------------- | ------ | ------------------------------- |
| 1   | Workspace badge | Small  | Нет — можно сделать сейчас      |
| 2   | Metadata row    | Small  | Нет — основа для всего          |
| 3   | Relative dates  | Small  | После metadata row              |
| 4   | Priority border | Small  | После реализации priority (3.1) |
| 5   | Теги chips      | Medium | После реализации тегов (3.4)    |
| 6   | Completion info | Small  | Нет                             |
| 7   | Quick actions   | Medium | Нет                             |

---

## 5. Приоритетная таблица

| #   | Фича                              | Effort       | Impact    | Категория |
| --- | --------------------------------- | ------------ | --------- | --------- |
| 1   | AI-чат в веб-UI                   | Medium-Large | Very High | AI        |
| 2   | Smart Quick Add с NLP             | Medium       | High      | AI        |
| 3   | AI-саммари и отчёты               | Medium       | High      | AI        |
| 4   | AI-планирование дня               | Medium       | High      | AI        |
| 5   | Dashboard / Главная               | Medium       | High      | UI        |
| 6   | Приоритеты задач                  | Small        | High      | Impl      |
| 7   | Command Palette (Cmd+K)           | Medium       | High      | UI        |
| 8   | Глобальный поиск                  | Medium       | High      | Impl      |
| 9   | Контекстные AI-действия           | Medium       | Medium    | AI        |
| 10  | Улучшения агента (streaming, etc) | Small-Medium | Medium    | AI        |
| 11  | Badge на Inbox                    | Small        | Medium    | UI        |
| 12  | Inline создание тасков            | Small        | Medium    | UI        |
| 13  | Мульти-провайдер LLM              | Small-Medium | Medium    | AI        |
| 14  | Умные напоминания                 | Large        | Medium    | AI        |
| 15  | Недельный вид календаря           | Medium       | Medium    | UI        |
| 16  | Empty States                      | Small        | Medium    | UI        |
| 17  | Повторяющиеся задачи              | Large        | High      | Impl      |
| 18  | Теги / Метки                      | Medium       | Medium    | Impl      |
| 19  | Markdown preview                  | Small        | Medium    | Impl      |
| 20  | Overdue reschedule                | Small        | Medium    | Impl      |
| 21  | Mobile-оптимизация                | Medium       | Medium    | UI        |
| 22  | RT Toast-уведомления              | Small        | Small     | UI        |

---

## Главный вывод

Агент уже реализован как полноценный tool-use AI с conversation memory, 14 инструментами и multi-iteration loop. **Самое ценное улучшение** — вынести этого агента в веб-UI (пункт 1.1). Backend полностью готов (`agent.Process()`) — нужны только Connect-RPC обёртка и React-чат. Это превращает DoneJournal из таск-трекера с ботом в **AI-first productivity platform**, где агент доступен везде: в Telegram, в вебе, и контекстно на каждом экране.

Второй приоритет — AI-саммари (1.3) и планирование дня (1.4): агент уже имеет все нужные tools (`get_todo_stats`, `find_todos`, `create_todo`), нужна только правильная промпт-инженерия и UI для отображения результатов.
