# DoneJournal — Предложения по улучшению

## Context

DoneJournal — productivity-приложение с Telegram-ботом для логирования ежедневных достижений. Стек: Go + Connect-RPC + SQLite (бэк), React 19 + TanStack + shadcn/Base UI (фронт), Groq/Llama для AI-парсинга сообщений. Текущие фичи: inbox, todos, notes, calendar с drag-and-drop, workspaces, real-time подписки, Telegram-бот с NLP и распознаванием голоса (Whisper).

---

## 1. AI-фичи (главный приоритет)

### 1.1 AI-обработка Inbox в веб-UI

**Effort**: Medium | **Impact**: High

Сейчас AI-парсинг (`mcp.ParseMessage`) работает только через Telegram-бот. Нужно вынести в веб.

- **Backend**: Новый RPC `ProcessInboxItem(id)` в `InboxService` (или отдельный `AIService`). Вызывает `mcpService.ParseMessage()`, показывает preview, потом создаёт todos/notes
- **Frontend**: Кнопка "волшебная палочка" на каждой `inbox-card.tsx` + кнопка "Process All" для пакетной обработки
- **UX**: Preview-диалог с результатом парсинга (title, date, workspace, type) перед подтверждением
- **Ключевые файлы**: `internal/processor/processor.go` (извлечь логику в reusable сервис), `inbox.proto`, `inbox-card.tsx`

### 1.2 Smart Quick Add с естественным языком

**Effort**: Medium | **Impact**: High

Превратить `inbox-quick-add.tsx` в умную строку ввода (как бот, но в вебе):

- Пользователь пишет: "закончить отчёт завтра для проекта Alpha"
- Новый RPC `ParseText` → inline preview: "Todo: закончить отчёт | Дата: 14 мар | Workspace: Alpha"
- Подтверждение или редактирование перед сохранением
- Fallback в inbox-item при ошибке парсинга

### 1.3 AI-саммари дня/недели

**Effort**: Medium | **Impact**: Medium

Новый RPC `SummarizeActivity(dateFrom, dateTo)` — генерирует текстовое резюме выполненных задач через Groq. Полезно для стендапов, ретроспектив, дневников. Кнопка "Generate Summary" на дашборде или календаре.

### 1.4 AI-подсказки и автопланирование

**Effort**: Large | **Impact**: Medium

На основе паттернов завершённых задач предлагать повторяющиеся задачи. Если "standup" выполняется каждый Пн-Пт — предложить автоматизацию. Более сложная фича, но инфраструктура уже есть.

---

## 2. UI/UX улучшения

### 2.1 Dashboard / Главная страница

**Effort**: Medium | **Impact**: High

Сейчас `/` редиректит на `/inbox`. Заменить на дашборд:

- Задачи на сегодня (pending + in_progress)
- Статистика: выполнено за неделю, streak
- Счётчик inbox-элементов
- Quick add bar
- Последние заметки
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
- Обновить LLM-промпт в `mcp.go` для извлечения приоритета из текста

### 3.2 Повторяющиеся задачи

**Effort**: Large | **Impact**: High

- Поле `recurrence_rule` + `recurrence_parent_id` в таблице todos
- River job для генерации следующего occurrence при завершении
- Picker в `todo-form.tsx` (daily, weekly, monthly, custom)
- LLM-промпт: "каждый день", "еженедельно", "каждый понедельник"

### 3.3 Глобальный поиск

**Effort**: Medium | **Impact**: High

У notes уже есть `search`. Расширить на todos и inbox. Единая страница поиска или интеграция в Command Palette.

### 3.4 Теги / Метки

**Effort**: Medium | **Impact**: Medium

Таблицы `tags`, `todo_tags`, `note_tags`. Цветные теги, фильтрация. LLM извлекает хештеги из сообщений.

### 3.5 Пакетное перенесение просроченных задач

**Effort**: Small | **Impact**: Medium

Кнопка "Перенести все на сегодня" в `overdue-banner.tsx`. Новый `BatchUpdateTodos` RPC или цикл `updateTodo`.

### 3.6 Markdown preview в описаниях тасков

**Effort**: Small | **Impact**: Medium

Рендерить `todo.description` как markdown в `todo-item.tsx` (переиспользовать паттерн из notes).

---

## 4. Приоритетная таблица

| #   | Фича                        | Effort | Impact | Категория |
| --- | --------------------------- | ------ | ------ | --------- |
| 1   | AI-обработка Inbox в веб-UI | Medium | High   | AI        |
| 2   | Smart Quick Add с NLP       | Medium | High   | AI        |
| 3   | Dashboard / Главная         | Medium | High   | UI        |
| 4   | Приоритеты задач            | Small  | High   | Impl      |
| 5   | Command Palette (Cmd+K)     | Medium | High   | UI        |
| 6   | Глобальный поиск            | Medium | High   | Impl      |
| 7   | Badge на Inbox              | Small  | Medium | UI        |
| 8   | Inline создание тасков      | Small  | Medium | UI        |
| 9   | AI-саммари дня/недели       | Medium | Medium | AI        |
| 10  | Недельный вид календаря     | Medium | Medium | UI        |
| 11  | Empty States                | Small  | Medium | UI        |
| 12  | Повторяющиеся задачи        | Large  | High   | Impl      |
| 13  | Теги / Метки                | Medium | Medium | Impl      |
| 14  | Markdown preview            | Small  | Medium | Impl      |
| 15  | Overdue reschedule          | Small  | Medium | Impl      |
| 16  | Mobile-оптимизация          | Medium | Medium | UI        |
| 17  | RT Toast-уведомления        | Small  | Small  | UI        |
| 18  | AI-автопланирование         | Large  | Medium | AI        |

---

## Главный вывод

**Самое ценное улучшение** — вынести AI-парсинг из Telegram-бота в веб-UI (пункты 1.1 и 1.2). Инфраструктура (`mcp.go`, `processor.go`, Groq API) уже полностью готова на бэкенде — нужно только создать RPC-обёртку и фронтенд-компоненты. Это превращает DoneJournal из "удобного таск-трекера" в "AI-powered productivity tool".
