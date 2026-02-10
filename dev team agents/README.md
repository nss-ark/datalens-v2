# DataLens 2.0 — Multi-Agent Development System

## How It Works

You operate a **hub-and-spoke** model across separate chats:

```
  ┌──────────────────────────┐
  │   ORCHESTRATOR CHAT      │  ← You start here each session
  │   (orchestrator.md)      │
  │                          │
  │   Reads TASK_TRACKER.md  │
  │   Produces task specs    │
  │   Reviews agent outputs  │
  └──────────┬───────────────┘
             │ produces task specs
             ▼
  ┌──────────────────────────┐
  │   YOU (Human Router)     │  ← You copy task specs into new chats
  └──┬────────┬────────┬─────┘
     │        │        │
     ▼        ▼        ▼
  ┌──────┐ ┌──────┐ ┌──────┐
  │ Chat │ │ Chat │ │ Chat │   ← Sub-agent chats (one per task)
  │  BE  │ │ AI/ML│ │ Test │
  └──────┘ └──────┘ └──────┘
```

## Step-by-Step Execution Flow

### Session Start

1. **Open a new chat** and paste the contents of `agents/orchestrator.md`
2. Tell them: *"Pick up the next sprint. Read `TASK_TRACKER.md` and decompose the next unblocked items into task specs."*
3. The orchestrator will produce numbered **Task Specification** documents

### Running Tasks

4. For each task spec the orchestrator produces:
   - **Open a new chat**
   - Paste the appropriate sub-agent prompt (`backend-agent.md`, `ai-ml-agent.md`, or `test-agent.md`)
   - Then paste the task spec from the orchestrator
   - Let the agent execute
   - Copy the agent's **completion summary** back to the orchestrator chat

5. **Parallel tasks** (no dependencies between them) can run as separate chats simultaneously

### Session End

6. Return to the orchestrator chat with all results
7. Tell it: *"Here are the results from this batch. Update TASK_TRACKER and plan the next batch."*

## When to Open the Orchestrator Chat

- **Start of each work session** — to get the next batch of tasks
- **After completing a batch** — to report results and get the next batch
- **When you hit a blocker** — to re-plan or adjust priorities

## File Index

| File | Purpose | When to use |
|------|---------|-------------|
| `orchestrator.md` | System prompt for orchestrator agent | Start of each planning session |
| `backend-agent.md` | System prompt for Go backend tasks | When running backend task specs |
| `ai-ml-agent.md` | System prompt for AI/ML tasks | When running detection/AI tasks |
| `test-agent.md` | System prompt for testing tasks | When running test task specs |

## Tips

- **Don't merge task specs** — one chat, one task. Keeps context focused.
- **The orchestrator doesn't write code** — it writes specs. Sub-agents write code.
- **Always verify compilation** — Every sub-agent chat should end with `go build ./...` passing.
- **Visual review checkpoints** — When the orchestrator flags "READY FOR VISUAL REVIEW", spin up the app locally and start a review chat.
