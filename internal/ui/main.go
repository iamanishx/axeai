package ui

import (
	"axe-desktop/internal/agent"
	"axe-desktop/internal/config"
	"axe-desktop/internal/storage"
	"axe-desktop/pkg/models"
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type MainUI struct {
	window       fyne.Window
	storage      *storage.Storage
	config       *config.Config
	agentService *agent.Service

	sidebar  *Sidebar
	chatView *ChatView
	composer *Composer

	currentSessionID string
}

func New(window fyne.Window, store *storage.Storage, cfg *config.Config, agentSvc *agent.Service) *MainUI {
	return &MainUI{
		window:       window,
		storage:      store,
		config:       cfg,
		agentService: agentSvc,
	}
}

func (ui *MainUI) Initialize() {
	ui.sidebar = NewSidebar(ui.storage, ui.onSessionSelected, ui.onNewSession, ui.onDeleteSession)
	ui.chatView = NewChatView()
	ui.composer = NewComposer(ui.onSendMessage)

	centralColumn := container.NewBorder(
		nil,
		ui.composer.Container(),
		nil, nil,
		ui.chatView.Container(),
	)

	content := container.NewHSplit(ui.sidebar.Container(), centralColumn)
	content.SetOffset(0.15)

	ui.window.SetContent(content)
	ui.window.SetMainMenu(ui.createMenu())

	ui.sidebar.LoadSessions("default")
}

func (ui *MainUI) createMenu() *fyne.MainMenu {
	return fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New Session", ui.onNewSession),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				ui.window.Close()
			}),
		),
		fyne.NewMenu("Settings",
			fyne.NewMenuItem("Providers & MCP", ui.showSettingsDialog),
		),
	)
}

func (ui *MainUI) onSessionSelected(sessionID string) {
	ui.currentSessionID = sessionID
	ui.chatView.Clear()

	messages, err := ui.storage.ListMessages(sessionID, 100, 0)
	if err != nil {
		return
	}

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		ui.chatView.AddMessage(string(msg.Role), msg.Content)
	}

	ui.chatView.ClearStatus()
}

func (ui *MainUI) onNewSession() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Session name")

	primaryBtn := widget.NewButton("Create", nil)
	primaryBtn.Importance = widget.HighImportance

	secondaryBtn := widget.NewButton("Cancel", nil)
	secondaryBtn.Importance = widget.LowImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("New Session", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nameEntry,
		container.NewHBox(layout.NewSpacer(), secondaryBtn, primaryBtn),
	)

	d := dialog.NewCustomWithoutButtons("", content, ui.window)
	d.Resize(fyne.NewSize(400, 180))

	secondaryBtn.OnTapped = func() { d.Hide() }

	primaryBtn.OnTapped = func() {
		title := nameEntry.Text
		if title == "" {
			title = "New Chat"
		}

		provider := ui.config.GetActiveProvider()

		session := &models.Session{
			UserID:       "default",
			Title:        title,
			Model:        provider.Model,
			ProviderID:   provider.ID,
			SystemPrompt: "You are a helpful AI assistant.",
		}

		if err := ui.storage.CreateSession(session); err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		ui.sidebar.AddSession(*session)
		ui.sidebar.sessionList.Select(0)
		d.Hide()
	}

	d.Show()
}

func (ui *MainUI) onDeleteSession(sessionID string) {
	if sessionID == "" {
		return
	}

	confirm := dialog.NewConfirm(
		"Delete Session",
		"Delete this session and all its messages?",
		func(ok bool) {
			if !ok {
				return
			}
			if err := ui.storage.DeleteSession(sessionID); err != nil {
				dialog.ShowError(err, ui.window)
				return
			}
			if ui.currentSessionID == sessionID {
				ui.currentSessionID = ""
				ui.chatView.Clear()
				ui.agentService.RemoveRunner(sessionID)
			}
			ui.sidebar.LoadSessions("default")
		},
		ui.window,
	)
	confirm.Show()
}

func (ui *MainUI) showSettingsDialog() {
	provider := ui.config.GetActiveProvider()

	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetPlaceHolder("Enter API Key")
	if provider != nil && provider.APIKey != "" {
		apiKeyEntry.SetText(provider.APIKey)
	}

	modelEntry := widget.NewEntry()
	modelEntry.SetPlaceHolder("Model (e.g. gemini-1.5-flash)")
	if provider != nil {
		modelEntry.SetText(provider.Model)
	}

	saveBtn := widget.NewButton("Save", func() {
		if provider != nil {
			provider.APIKey = apiKeyEntry.Text
			provider.Model = modelEntry.Text
			if err := ui.config.Save(); err != nil {
				dialog.ShowError(err, ui.window)
				return
			}
			ui.agentService.RemoveRunner(ui.currentSessionID)
			dialog.ShowInformation("Settings Saved", "Provider settings updated.", ui.window)
		}
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", nil)

	content := container.NewVBox(
		widget.NewLabelWithStyle("API Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Active Provider: "+provider.Name),
		widget.NewLabel("API Key"),
		apiKeyEntry,
		widget.NewLabel("Model"),
		modelEntry,
		container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn),
	)

	d := dialog.NewCustomWithoutButtons("", content, ui.window)
	d.Resize(fyne.NewSize(500, 300))

	cancelBtn.OnTapped = func() { d.Hide() }

	d.Show()
}

func (ui *MainUI) onSendMessage(content string) {
	if ui.currentSessionID == "" {
		provider := ui.config.GetActiveProvider()
		session := &models.Session{
			UserID:       "default",
			Title:        content[:min(50, len(content))] + "...",
			Model:        provider.Model,
			ProviderID:   provider.ID,
			SystemPrompt: "You are a helpful AI assistant.",
		}

		if err := ui.storage.CreateSession(session); err != nil {
			return
		}

		ui.sidebar.AddSession(*session)
		ui.currentSessionID = session.ID
	}

	ui.chatView.AddMessage("user", content)
	ui.chatView.AddMessage("assistant", "")
	ui.chatView.SetStatus("Thinking...")
	ui.composer.SetEnabled(false)

	ctx := context.Background()
	err := ui.agentService.SendMessage(ctx, ui.currentSessionID, content,
		func(role, content string) {
			if role == "assistant" {
				fyne.Do(func() {
					if content != "" {
						ui.chatView.ClearStatus()
					}
					ui.chatView.UpdateLastMessage(content)
				})
				return
			}
			if role == "system" {
				fyne.Do(func() {
					ui.chatView.ClearStatus()
					ui.chatView.RemoveLastAssistantIfEmpty()
					ui.chatView.AddMessage("system", content)
				})
			}
		},
		func(toolName string, args, result map[string]any, err error) {
			fyne.Do(func() {
				if toolName != "" {
					if result != nil {
						ui.chatView.AddNote("Tool done: " + toolName)
						return
					}
					ui.chatView.AddNote("Tool: " + toolName)
				}
			})
		},
		nil,
	)

	if err != nil {
		ui.chatView.AddMessage("system", fmt.Sprintf("Error: %v", err))
	}

	ui.composer.SetEnabled(true)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
