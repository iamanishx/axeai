# Axe Desktop

A Claude Desktop-inspired native AI chat application built with Go, Fyne, ADK-Go, and SQLite. Features a Vercel-inspired dark theme with streaming responses and multi-session support.

## Features

- **Native Desktop App**: Cross-platform application using Fyne UI framework
- **Minimal Vercel Design**: Clean black (#000000) background with sleek, minimal UI - no fat buttons or bulky cards
- **Multi-Session Chat**: Create and manage multiple concurrent chat sessions
- **Streaming Responses**: Real-time AI response streaming with visual feedback
- **Local-First Storage**: SQLite database for persistence
- **ADK-Go Integration**: Google's Agent Development Kit for AI orchestration
- **Tool Integration**: MCP tools support (Exa web search ready, filesystem tools planned)
- **Session Management**: Full CRUD operations for sessions, messages, and tool calls

## Architecture

```
axe-desktop/
├── cmd/axe-desktop/          # Main application entry point
├── internal/
│   ├── agent/                # ADK-Go agent service with streaming
│   ├── config/               # Configuration management
│   ├── storage/              # SQLite storage layer with migrations
│   └── ui/                   # Fyne UI components
│       ├── main.go           # Main UI coordinator
│       ├── sidebar.go        # Session list sidebar
│       ├── chat.go           # Message display
│       ├── composer.go       # Input composer
│       ├── toolpanel.go      # Tool trace panel
│       └── theme.go          # Vercel color theme
├── pkg/models/               # Data models
└── AGENTS.md                 # Architecture documentation
```

## Data Model (SQLite)

- **users**: User profiles
- **sessions**: Chat sessions with metadata
- **messages**: Chat messages with streaming status support
- **tool_calls**: Tool invocation records
- **attachments**: File attachments
- **settings**: Application configuration

## Getting Started

### Prerequisites

- Go 1.24+ (uses iter.Seq2 for streaming)
- SQLite development libraries
- API keys for:
  - Google Gemini (or other supported model)
  - Exa MCP (optional, for web search)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd axe-desktop

# Install dependencies
go mod tidy

# Build the application
go build ./cmd/axe-desktop

# Run the application
./axe-desktop
```

### Configuration

Set environment variables:

```bash
export AXE_API_KEY="your-gemini-api-key"
export EXA_API_KEY="your-exa-api-key"  # Optional, for web search
export AXE_MODEL="gemini-2.0-flash"    # Default model
```

Or create a config file at `~/.axe-desktop/config.json`:

```json
{
  "model_name": "gemini-2.0-flash",
  "api_key": "your-api-key",
  "exa_enabled": true,
  "exa_api_key": "your-exa-api-key"
}
```

## Usage

1. **Create a Session**: Click the "+" button in the sidebar or use File > New Session
2. **Send Messages**: Type in the composer and hit Enter or click Send
3. **View Sessions**: Click on any session in the sidebar to switch
4. **Tool Traces**: View tool calls and reasoning in the right panel
5. **Streaming**: Watch AI responses appear in real-time

## UI Layout

```
┌─────────────────────────────────────────────────────────────┐
│  Sidebar    │  Chat Area                    │  Tool Panel │
│             │                               │             │
│  Sessions   │  ┌─────────────────────────┐  │  Tool Calls │
│  List       │  │ Messages                │  │  Reasoning  │
│             │  │                         │  │             │
│  [+] New    │  │ User: Hello!            │  │             │
│             │  │ Assistant: Hi there!    │  │             │
│             │  │                         │  │             │
│             │  └─────────────────────────┘  │             │
│             │                               │             │
│             │  [Type your message...] [Send]             │
│             │                               │             │
└─────────────────────────────────────────────────────────────┘
```

## Development

### Project Structure

The codebase follows clean architecture principles:

- **Modular Design**: Clear separation between UI, storage, and agent layers
- **Interface-Based**: Easy to swap implementations
- **Event-Driven**: Streaming responses via channels
- **Type-Safe**: Strong typing throughout

### Key Components

1. **Agent Service** (`internal/agent/service.go`)
   - Manages ADK-Go runners
   - Handles streaming events
   - Persists messages and tool calls

2. **Storage Layer** (`internal/storage/storage.go`)
   - SQLite with WAL mode
   - Automatic migrations
   - Foreign key constraints

3. **UI Layer** (`internal/ui/`)
   - Vercel-inspired dark theme
   - Responsive layout with split panes
   - Real-time message updates

### Adding Features

**New MCP Tools:**
1. Add tool configuration to agent service
2. Register tools in runner creation
3. Update tool panel to display results

**New UI Components:**
1. Create component file in `internal/ui/`
2. Apply Vercel theme colors
3. Register in main UI coordinator

## Building for Production

### Using fyne-cross

```bash
# Install fyne-cross
go install github.com/fyne-io/fyne-cross@latest

# Build for multiple platforms
fyne-cross windows -output axe-desktop.exe
cd /home/manish/projects/axe-desktop && fyne-cross linux -output axe-desktop
cd /home/manish/projects/axe-desktop && fyne-cross darwin -output axe-desktop
```

### Manual Build

```bash
# Windows
go build -o axe-desktop.exe ./cmd/axe-desktop

# macOS
go build -o axe-desktop ./cmd/axe-desktop

# Linux
go build -o axe-desktop ./cmd/axe-desktop
```

## Roadmap

### Phase 1: Foundation ✓
- [x] Fyne UI with Vercel theme
- [x] SQLite storage with migrations
- [x] Session management
- [x] Message rendering

### Phase 2: Agent Integration ✓
- [x] ADK-Go integration
- [x] Streaming responses
- [x] Message persistence
- [x] Session context

### Phase 3: MCP Tools
- [ ] Exa MCP integration
- [ ] Web search capabilities
- [ ] Tool permission system
- [ ] Tool trace panel

### Phase 4: Multi-Session
- [ ] Concurrent session support
- [ ] Session switching without losing state
- [ ] Session archive/restore

### Phase 5: Filesystem MCP
- [ ] File browser integration
- [ ] Code editing capabilities
- [ ] Project-aware context

## Technology Stack

- **UI Framework**: [Fyne](https://fyne.io/) v2.7.2
- **AI Framework**: [ADK-Go](https://github.com/google/adk-go) v0.4.0
- **AI Models**: Google Gemini (via genai SDK)
- **Storage**: SQLite with [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- **MCP**: [Exa MCP Server](https://exa.ai/docs/reference/exa-mcp)
- **Build Tool**: [fyne-cross](https://github.com/fyne-io/fyne-cross)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./...`
5. Build and test locally
6. Submit a pull request

## License

MIT License - See LICENSE file for details

## Acknowledgments

- [Google ADK-Go](https://github.com/google/adk-go) - Agent framework
- [Fyne](https://fyne.io/) - Cross-platform GUI framework
- [Exa](https://exa.ai/) - Web search MCP server
- [Vercel](https://vercel.com/) - Design inspiration

## Support

For issues and feature requests, please use the GitHub issue tracker.

---

**Note**: This application requires valid API keys for AI services. Store keys securely and never commit them to version control.
