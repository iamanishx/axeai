package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type MessageBubble struct {
	container *fyne.Container
	content   *widget.Label
}

func NewMessageBubble(role, content string) *MessageBubble {
	mb := &MessageBubble{}

	roleText := "Assistant"
	roleColor := VercelMuted
	if role == "user" {
		roleText = "You"
		roleColor = VercelWhite
	} else if role == "system" {
		roleText = "System"
		roleColor = VercelMuted
	}

	roleLabel := canvas.NewText(roleText, roleColor)
	roleLabel.TextSize = 11
	roleLabel.TextStyle = fyne.TextStyle{Bold: true}

	mb.content = widget.NewLabel(content)
	mb.content.Wrapping = fyne.TextWrapWord
	mb.content.Alignment = fyne.TextAlignLeading

	bg := canvas.NewRectangle(VercelDarkGray)
	bg.CornerRadius = 8
	if role == "user" {
		bg.FillColor = VercelGray
	}
	bg.SetMinSize(fyne.NewSize(280, 0))

	contentBox := container.NewVBox(
		container.NewPadded(roleLabel),
		container.NewPadded(mb.content),
	)

	bubbleStack := container.NewStack(bg, contentBox)

	maxWidth := float32(680)
	minWidth := float32(280)
	if role == "assistant" {
		maxWidth = 684
		minWidth = 684
	}
	bubble := container.New(&MaxWidthLayout{MaxWidth: maxWidth, MinWidth: minWidth}, bubbleStack)

	alignBox := bubble
	if role == "user" {
		alignBox = container.NewHBox(layout.NewSpacer(), bubble)
	} else {
		alignBox = container.NewHBox(bubble, layout.NewSpacer())
	}

	gap := canvas.NewRectangle(color.Transparent)
	gap.SetMinSize(fyne.NewSize(0, 12))

	mb.container = container.NewVBox(container.NewPadded(alignBox), gap)

	return mb
}

func (mb *MessageBubble) GetContainer() fyne.CanvasObject {
	return mb.container
}

func (mb *MessageBubble) UpdateContent(content string) {
	mb.content.SetText(content)
}

type ChatView struct {
	scrollContainer *container.Scroll
	messages        *fyne.Container
	messageWidgets  []*MessageBubble
	currentRole     string
}

func NewChatView() *ChatView {
	content := container.NewVBox()
	cv := &ChatView{messages: content}

	contentWrapper := container.New(&MaxWidthLayout{MaxWidth: 760, MinWidth: 760}, content)
	centered := container.NewHBox(layout.NewSpacer(), contentWrapper, layout.NewSpacer())
	paddedContent := container.NewPadded(centered)

	cv.scrollContainer = container.NewScroll(paddedContent)
	cv.scrollContainer.SetMinSize(fyne.NewSize(760, 520))
	return cv
}

func (cv *ChatView) Container() fyne.CanvasObject {
	return cv.scrollContainer
}

func (cv *ChatView) AddMessage(role, content string) {
	if len(cv.messageWidgets) > 0 && cv.currentRole == role && role == "assistant" {
		cv.UpdateLastMessage(content)
		return
	}

	bubble := NewMessageBubble(role, content)
	cv.messages.Add(bubble.GetContainer())
	cv.messageWidgets = append(cv.messageWidgets, bubble)
	cv.currentRole = role
	cv.scrollContainer.ScrollToBottom()
}

func (cv *ChatView) Clear() {
	cv.messages.Objects = nil
	cv.messageWidgets = nil
	cv.currentRole = ""
	cv.messages.Refresh()
}

func (cv *ChatView) UpdateLastMessage(content string) {
	if len(cv.messageWidgets) == 0 {
		return
	}
	cv.messageWidgets[len(cv.messageWidgets)-1].UpdateContent(content)
}
