package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"axe-desktop/internal/config"
	"axe-desktop/internal/storage"
	"axe-desktop/pkg/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/mcptoolset"
	"google.golang.org/genai"
)

type Service struct {
	config         *config.Config
	storage        *storage.Storage
	sessionService session.Service
	runners        map[string]*runner.Runner
	cancelFuncs    map[string]context.CancelFunc
	mu             sync.RWMutex
}

type MessageHandler func(role, content string)
type ToolCallHandler func(toolName string, args, result map[string]any, err error)
type DebugHandler func(line string)

func NewService(cfg *config.Config, store *storage.Storage) (*Service, error) {
	return &Service{
		config:         cfg,
		storage:        store,
		sessionService: session.InMemoryService(),
		runners:        make(map[string]*runner.Runner),
		cancelFuncs:    make(map[string]context.CancelFunc),
	}, nil
}

func (s *Service) getOrCreateRunner(sessionID string, provider *models.Provider) (*runner.Runner, error) {
	s.mu.RLock()
	r, exists := s.runners[sessionID]
	s.mu.RUnlock()

	if exists {
		return r, nil
	}

	ctx := context.Background()

	model, err := gemini.NewModel(ctx, provider.Model, &genai.ClientConfig{
		APIKey: provider.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	var agentToolsets []tool.Toolset
	for _, mcpSrv := range s.config.MCPServers {
		if !mcpSrv.Enabled {
			continue
		}

		var transport mcp.Transport
		if mcpSrv.Type == models.MCPServerHTTP {
			transport = &mcp.StreamableClientTransport{
				Endpoint: mcpSrv.URL,
			}
		} else {
			continue
		}

		exaToolset, err := mcptoolset.New(mcptoolset.Config{
			Transport: transport,
		})
		if err == nil {
			agentToolsets = append(agentToolsets, exaToolset)
		}
	}

	llmAgent, err := llmagent.New(llmagent.Config{
		Name:        "axe-agent",
		Model:       model,
		Description: "Axe Desktop Assistant",
		Instruction: "You are a helpful AI assistant. Search the web when needed.",
		Toolsets:    agentToolsets,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	r, err = runner.New(runner.Config{
		AppName:        "axe-desktop",
		Agent:          llmAgent,
		SessionService: s.sessionService,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	s.mu.Lock()
	s.runners[sessionID] = r
	s.mu.Unlock()

	return r, nil
}

func (s *Service) ensureSession(ctx context.Context, sessionID string) (string, error) {
	_, err := s.sessionService.Get(ctx, &session.GetRequest{
		AppName:   "axe-desktop",
		UserID:    "default",
		SessionID: sessionID,
	})
	if err == nil {
		return sessionID, nil
	}

	createResp, err := s.sessionService.Create(ctx, &session.CreateRequest{
		AppName:   "axe-desktop",
		UserID:    "default",
		SessionID: sessionID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return createResp.Session.ID(), nil
}

func (s *Service) SendMessage(ctx context.Context, sessionID string, content string,
	onMessage MessageHandler, onToolCall ToolCallHandler, onDebug DebugHandler) error {

	provider := s.config.GetActiveProvider()
	if provider == nil {
		return fmt.Errorf("no active provider configured")
	}
	if provider.APIKey == "" {
		return fmt.Errorf("no API key configured for provider %s", provider.Name)
	}
	if onDebug != nil {
		onDebug(fmt.Sprintf("provider=%s model=%s", provider.Name, provider.Model))
	}
	fmt.Printf("[Agent] provider=%s model=%s\n", provider.Name, provider.Model)
	fmt.Printf("[Agent] api_key_len=%d\n", len(provider.APIKey))

	r, err := s.getOrCreateRunner(sessionID, provider)
	if err != nil {
		return err
	}

	adkSessionID, err := s.ensureSession(ctx, sessionID)
	if err != nil {
		return err
	}

	userMsg := &models.Message{
		SessionID: sessionID,
		Role:      models.RoleUser,
		Content:   content,
	}
	if err := s.storage.CreateMessage(userMsg); err != nil {
		return err
	}

	assistantMsg := &models.Message{
		SessionID: sessionID,
		Role:      models.RoleAssistant,
		Content:   "",
	}
	if err := s.storage.CreateMessage(assistantMsg); err != nil {
		return err
	}

	streamCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.cancelFuncs[sessionID] = cancel
	s.mu.Unlock()

	userContent := genai.NewContentFromText(content, genai.RoleUser)

	go s.handleStreaming(streamCtx, r, adkSessionID, assistantMsg, userContent, onMessage, onToolCall, onDebug)

	return nil
}

func (s *Service) handleStreaming(ctx context.Context, r *runner.Runner, sessionID string,
	assistantMsg *models.Message, userContent *genai.Content,
	onMessage MessageHandler, onToolCall ToolCallHandler, onDebug DebugHandler) {

	var fullResponse strings.Builder
	var gotContent bool
	if onDebug != nil {
		onDebug(fmt.Sprintf("session=%s start", sessionID))
	}
	fmt.Printf("[Agent] session=%s start\n", sessionID)

	defer func() {
		s.mu.Lock()
		delete(s.cancelFuncs, assistantMsg.SessionID)
		s.mu.Unlock()

		assistantMsg.Content = fullResponse.String()
		s.storage.UpdateMessage(assistantMsg)
	}()

	stream := func(mode agent.StreamingMode) bool {
		if onDebug != nil {
			onDebug(fmt.Sprintf("streaming_mode=%v", mode))
		}
		fmt.Printf("[Agent] streaming_mode=%v\n", mode)
		events := r.Run(ctx, "default", sessionID, userContent, agent.RunConfig{
			StreamingMode: mode,
		})
		eventCount := 0

		for event, err := range events {
			eventCount++
			if err != nil {
				onMessage("system", fmt.Sprintf("Error: %v", err))
				if onDebug != nil {
					onDebug(fmt.Sprintf("error=%v", err))
				}
				fmt.Printf("[Agent] error=%v\n", err)
				return true
			}
			if event == nil {
				if onDebug != nil {
					onDebug("event=nil")
				}
				fmt.Println("[Agent] event=nil")
				continue
			}

			select {
			case <-ctx.Done():
				onMessage("assistant", fullResponse.String()+"\n[Cancelled]")
				return gotContent
			default:
				if event.ErrorCode != "" {
					onMessage("system", fmt.Sprintf("Error: %s - %s", event.ErrorCode, event.ErrorMessage))
					if onDebug != nil {
						onDebug(fmt.Sprintf("error=%s message=%s", event.ErrorCode, event.ErrorMessage))
					}
					fmt.Printf("[Agent] error=%s message=%s\n", event.ErrorCode, event.ErrorMessage)
					return true
				}

				if event.Content != nil {
					if onDebug != nil && !gotContent {
						onDebug("content=present")
					}
					if !gotContent {
						fmt.Println("[Agent] content=present")
					}
					for _, part := range event.Content.Parts {
						if part.Text != "" {
							fullResponse.WriteString(part.Text)
							onMessage("assistant", fullResponse.String())
							gotContent = true
						}

						if part.FunctionCall != nil {
							onToolCall(part.FunctionCall.Name, part.FunctionCall.Args, nil, nil)
							if onDebug != nil {
								onDebug(fmt.Sprintf("tool_call=%s", part.FunctionCall.Name))
							}
							fmt.Printf("[Agent] tool_call=%s\n", part.FunctionCall.Name)
						}

						if part.FunctionResponse != nil {
							onToolCall(part.FunctionResponse.Name, nil, part.FunctionResponse.Response, nil)
							if onDebug != nil {
								onDebug(fmt.Sprintf("tool_response=%s", part.FunctionResponse.Name))
							}
							fmt.Printf("[Agent] tool_response=%s\n", part.FunctionResponse.Name)
						}
					}
				}
			}
		}

		if eventCount == 0 {
			msg := "runner returned 0 events"
			if onDebug != nil {
				onDebug(msg)
			}
			fmt.Printf("[Agent] %s\n", msg)
		}

		return gotContent
	}

	if stream(agent.StreamingModeSSE) {
		return
	}

	if stream(agent.StreamingModeNone) {
		return
	}

	onMessage("system", "No response received. Check the model name and API key.")
	if onDebug != nil {
		onDebug("no_response")
	}
	fmt.Println("[Agent] no_response")
}

func (s *Service) RemoveRunner(sessionID string) {
	s.mu.Lock()
	delete(s.runners, sessionID)
	delete(s.cancelFuncs, sessionID)
	s.mu.Unlock()
}
