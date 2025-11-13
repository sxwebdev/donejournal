# AGENT.md

## **Purpose of the Agent**

This agent acts as an intelligent assistant for the _Done Journal_ ecosystem.
Its main responsibility is to help users quickly record what they have accomplished, manage their daily logs, plan their tasks, and retrieve structured summaries—especially for moments when the universal question strikes:
**"Uh… what did I even do yesterday at work?"**

The agent enhances the productivity workflow by interpreting user messages, organizing logs, suggesting TODOs, generating daily reports, and integrating seamlessly with the backend service.

---

## **Core Responsibilities**

1. **Capture Accomplishments**

   - Parse natural-language inputs like "fixed login bug" or "reviewed PR #42".
   - Automatically categorize and timestamp entries.
   - Allow quick follow-ups, edits, or tagging.

2. **Manage TODO Lists**

   - Create tasks, schedule them, update their status.
   - Convert completed TODOs directly into daily logs.
   - Suggest overlooked tasks or unfinished work.

3. **Generate Daily Summaries**

   - Produce a concise "what I did yesterday" report.
   - Highlight completed tasks, trends, and tags.
   - Prepare standup-ready or manager-friendly summaries.

4. **Assist With Planning**

   - Suggest daily focus items based on previous activity.
   - Analyze patterns (meetings, dev work, research, etc.)
   - Offer reminders or nudges.

5. **Provide Quick Queries**

   - "What did I do last Monday?"
   - "Show me all tasks tagged #bugfix this week."
   - "What's left in my TODO?"

6. **Integrate With Mini-App (future)**

   - Support UI interactions.
   - Return structured JSON or UI definitions.
   - Synchronize state between bot, backend, and app.

---

## **Agent Capabilities**

- Natural language understanding
- Context-aware user prompting
- Summarization and rewriting
- Time-aware grouping of entries
- TODO/Done conversion logic
- Generation of structured reports
- Lightweight recommendations

---

## **Non-Goals**

- Not a replacement for project management tools (Jira, Trello, etc.)
- Not a time-tracking system counting exact minutes.
- Not responsible for authentication or backend security.

---

## **TODO List for the Agent**

### **MVP Tasks**

- [ ] Implement handler for "I did…" free-form input
- [ ] Implement daily summary generator
- [ ] Implement TODO creation and completion
- [ ] Add tagging and auto-tag detection
- [ ] Add "yesterday / today / last week" queries
- [ ] Define message formats for backend communication
- [ ] Support basic error handling and user guidance

### **Usability Improvements**

- [ ] Suggest categories/tags based on text
- [ ] Smart reminders ("Want to log what you did today?")
- [ ] Auto-build standup statements
- [ ] Multi-entry parsing from long text dump

### **Mini-App Support (Phase 2)**

- [ ] Provide structured JSON responses for UI
- [ ] Implement sorting/filtering logic for tasks
- [ ] Generate weekly achievement highlights
- [ ] Detect patterns in user activity

### **Advanced Features**

- [ ] Lightweight integration hints for GitHub/Jira (optional)
- [ ] Detect duplicate entries or repeated tasks
- [ ] Mood/effort tagging (optional)
- [ ] Provide optional motivational summaries

---

## **Long-Term Vision**

The agent becomes a personal daily productivity companion that:

- remembers what the user did better than the user themselves
- turns chaotic task notes into structured work journals
- prepares perfect "What I did yesterday" answers
- supports day planning and accountability

All in a friendly, humorous, and lightweight form.
