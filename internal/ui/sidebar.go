package ui

import (
	"axe-desktop/internal/storage"
	"axe-desktop/pkg/models"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Sidebar struct {
	storage      *storage.Storage
	sessionList  *widget.List
	sessions     []models.Session
	onSelect     func(sessionID string)
	onNewSession func()
	container    *fyne.Container
}

func NewSidebar(store *storage.Storage, onSelect func(sessionID string), onNew func()) *Sidebar {
	s := &Sidebar{
		storage:      store,
		onSelect:     onSelect,
		onNewSession: onNew,
	}
	s.build()
	return s
}

func (s *Sidebar) build() {
	title := widget.NewLabel("Chats")
	title.TextStyle = fyne.TextStyle{Bold: true}

	newBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		s.onNewSession()
	})
	newBtn.Importance = widget.MediumImportance

	header := container.NewBorder(nil, nil, nil, newBtn, title)

	separator := canvas.NewRectangle(VercelGray)
	separator.SetMinSize(fyne.NewSize(0, 1))

	s.sessionList = widget.NewList(
		func() int { return len(s.sessions) },
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.DocumentIcon())
			label := widget.NewLabel("Session")
			label.Truncation = fyne.TextTruncateEllipsis
			return container.NewHBox(icon, label)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(s.sessions) {
				return
			}
			session := s.sessions[id]
			box := item.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)
			label.SetText(session.Title)
		},
	)

	s.sessionList.OnSelected = func(id widget.ListItemID) {
		if id < len(s.sessions) {
			s.onSelect(s.sessions[id].ID)
		}
	}

	s.container = container.NewBorder(
		container.NewVBox(header, separator),
		nil, nil, nil,
		s.sessionList,
	)
}

func (s *Sidebar) Container() fyne.CanvasObject {
	return s.container
}

func (s *Sidebar) LoadSessions(userID string) {
	sessions, err := s.storage.ListSessions(userID)
	if err != nil {
		fmt.Printf("Failed to load sessions: %v\n", err)
		return
	}
	s.sessions = sessions
	s.sessionList.Refresh()
}

func (s *Sidebar) AddSession(session models.Session) {
	s.sessions = append([]models.Session{session}, s.sessions...)
	s.sessionList.Refresh()
}
