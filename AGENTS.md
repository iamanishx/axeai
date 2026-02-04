# Agents Plan (Fyne Desktop + ADK-Go + MCP)

This document outlines a comprehensive plan to build a Claude-desktop-like native app using Fyne + fyne-cross, ADK-Go, SQLite storage, and MCP tools (Exa now, filesystem later). It also summarizes feasibility of streaming and “thinking” displays with sources.

## Feasibility (based on docs)

- **Cross-platform native UI:** Fyne supports desktop apps and fyne-cross builds for multiple platforms. This is feasible.
- **ADK-Go integration:** ADK-Go is an open-source toolkit for building agents in Go. Feasible. See ADK-Go repo and Go docs. https://github.com/google/adk-go and https://google.github.io/adk-docs/get-started/go/ and https://pkg.go.dev/google.golang.org/adk
- **MCP tools:** ADK supports MCP tools (client or server patterns). Feasible for Exa MCP and later filesystem MCP. https://google.github.io/adk-docs/tools-custom/mcp-tools/ and https://exa.ai/docs/reference/exa-mcp
- **Streaming responses:** ADK supports streaming (including bidirectional streaming for live interactions). Feasible for text streaming in a desktop UI, but the exact Go streaming API should be validated against ADK-Go docs. https://google.github.io/adk-docs/streaming/
- **“Show thinking” / chain-of-thought:** Raw chain-of-thought is generally not exposed by model providers and should not be displayed. You can show a structured “reasoning summary” or “tool trace” instead. This is a product/UI choice and not guaranteed by ADK.

## Product Goals

- Native desktop app with Claude-desktop-style chat, session sidebar, and tool output panel.
- Multiple concurrent chat sessions with independent context/state.
- Local-first storage in SQLite for users, sessions, messages, and tool calls.
- Agent orchestration via ADK-Go with MCP tool integration.
- Streaming assistant responses with cancel, retry, and partial rendering.

## Architecture Overview

- **UI Layer (Fyne):** App shell, session list, chat timeline, composer, and tool output view.
- **App Core (Go):** Orchestrates sessions, agent lifecycle, tool registry, streaming pipelines.
- **Agent Layer (ADK-Go):** LLM agent config, tools, memory/session handling.
- **Storage (SQLite):** Durable storage for users, sessions, messages, tool calls, and settings.
- **Integrations:** MCP clients (Exa now, filesystem later), model providers configured via ADK.

## Data Model (SQLite)

Suggested tables (minimal viable schema):

- `users` (id, name, created_at)
- `sessions` (id, user_id, title, model, created_at, updated_at, archived_at)
- `messages` (id, session_id, role, content, created_at, token_count, metadata_json)
- `tool_calls` (id, session_id, message_id, tool_name, args_json, result_json, created_at)
- `attachments` (id, session_id, message_id, type, path, metadata_json)
- `settings` (key, value_json)

Indexes:

- `messages(session_id, created_at)`
- `tool_calls(session_id, created_at)`
- `sessions(user_id, updated_at)`

## Session & Agent Lifecycle

- Each session maps to a unique ADK session or runner context.
- Allow multiple sessions to run concurrently with cancellation support.
- Store system prompts and session config in `sessions` and replay into context on load.
- Keep a “summary” field per session to reduce token load.

## Streaming Plan

- Use ADK streaming where available (validate Go API endpoints in ADK-Go).
- Stream token deltas to Fyne UI and append to the last assistant bubble.
- UI controls: Stop, Regenerate, Continue.
- Persist partial messages with status: `in_progress`, `completed`, `cancelled`, `failed`.

Reference: ADK streaming overview and live streaming docs. https://google.github.io/adk-docs/streaming/

## “Thinking” Display Plan

- Do not display raw chain-of-thought.
- Offer a **Reasoning Summary** panel (short, structured summary of why the answer was produced) and a **Tool Trace** panel (requests/responses).
- This keeps the experience transparent without exposing restricted internal reasoning.

## MCP Integration (Exa + Future Filesystem)

- Use ADK MCP tools to connect to remote MCP servers.
- Exa MCP remote endpoint: https://mcp.exa.ai/mcp
- Tools available by default: `web_search_exa`, `get_code_context_exa`, `company_research_exa`.
- Support tool filtering to expose only needed tools per agent.

References: https://exa.ai/docs/reference/exa-mcp and ADK MCP tools guide https://google.github.io/adk-docs/tools-custom/mcp-tools/

## ADK-Go Integration Plan

- Build an ADK LLM agent configured with the desired model (Gemini, Claude, etc.)
- Attach MCP toolsets as needed (Exa now, filesystem later).
- Use ADK session/memory features to manage persistent context.

References: Go quickstart https://google.github.io/adk-docs/get-started/go/ and API docs https://pkg.go.dev/google.golang.org/adk

## UX Structure (Claude-Desktop-like)

- Left sidebar: user profile + sessions list + new session button.
- Main column: message timeline with timestamps, tool outputs, and error states.
- Right panel (optional): active tools, reasoning summary, or session settings.
- Input composer: multiline with send, stop, and model selector.

## Security & Privacy

- Store API keys in OS keychain (recommended) or encrypted local store.
- Gate tool usage with per-tool permissions.
- Ensure MCP filesystem server is scoped to user-approved directory only.

## Build & Packaging

- Use fyne-cross for builds on macOS, Windows, Linux.
- Bundle SQLite and migrations into the app.
- Provide a single config file for MCP endpoints and tool enablement.

## Phased Implementation Plan

1. **Foundation**
   - Create Fyne shell and layouts
   - Add SQLite schema + migrations
   - Implement sessions list and message rendering

2. **Agent MVP**
   - Integrate ADK-Go with a single model
   - Run simple prompt/response flow
   - Persist messages to SQLite

3. **Streaming**
   - Add streaming UI renderer
   - Handle cancel/regenerate
   - Persist partial messages and completion state

4. **MCP Tools**
   - Add Exa MCP toolset
   - Show tool trace panel
   - Add opt-in tool permissions

5. **Multi-Session Concurrency**
   - Run multiple sessions in parallel
   - Add session switching without losing streaming state

6. **Filesystem MCP (future)**
   - Add local filesystem MCP server support
   - Scope to user-approved directories

## Open Questions to Validate Early

- Confirm the ADK-Go streaming API surface for text streaming in desktop apps.
- Decide how much context to persist vs. summarize for long sessions.
- Confirm model provider limits and policies for “reasoning summary.”

## Documentation Links (primary)

- ADK-Go repo: https://github.com/google/adk-go
- ADK Go quickstart: https://google.github.io/adk-docs/get-started/go/
- ADK MCP tools: https://google.github.io/adk-docs/tools-custom/mcp-tools/
- ADK streaming overview: https://google.github.io/adk-docs/streaming/
- ADK Go API reference: https://pkg.go.dev/google.golang.org/adk
- Exa MCP: https://exa.ai/docs/reference/exa-mcp
