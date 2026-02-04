package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MessageBubble struct {
	container *fyne.Container
	content   *widget.RichText
	segment   *widget.TextSegment
}

type StatusLine struct {
	container *fyne.Container
	text      *canvas.Text
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
	roleLabel.TextSize = theme.Size(ChatMetaSizeName)
	roleLabel.TextStyle = fyne.TextStyle{Bold: true}

	mb.segment = &widget.TextSegment{
		Text: content,
		Style: widget.RichTextStyle{
			Alignment: fyne.TextAlignLeading,
			SizeName:  ChatTextSizeName,
		},
	}
	mb.content = widget.NewRichText(mb.segment)
	mb.content.Wrapping = fyne.TextWrapWord

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
		maxWidth = 774
		minWidth = 774
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
	mb.segment.Text = content
	mb.content.Refresh()
}

func (mb *MessageBubble) Text() string {
	return mb.segment.Text
}

func NewStatusLine(content string) *StatusLine {
	text := canvas.NewText(content, VercelMuted)
	text.TextSize = theme.Size(ChatMetaSizeName)

	line := container.NewHBox(text)
	return &StatusLine{
		container: line,
		text:      text,
	}
}

func (sl *StatusLine) Container() fyne.CanvasObject {
	return sl.container
}

func (sl *StatusLine) SetText(content string) {
	sl.text.Text = content
	sl.text.Refresh()
}

type ChatView struct {
	scrollContainer *container.Scroll
	messages        *fyne.Container
	messageWidgets  []*MessageBubble
	currentRole     string
	statusLine      *StatusLine
	lastAssistant   fyne.CanvasObject
	adjustingScroll bool
	lastScroll      fyne.Position
}

func NewChatView() *ChatView {
	content := container.NewVBox()
	cv := &ChatView{messages: content}

	contentWrapper := container.New(&MaxWidthLayout{MaxWidth: 860, MinWidth: 860}, content)
	centered := container.NewHBox(layout.NewSpacer(), contentWrapper, layout.NewSpacer())
	paddedContent := container.NewPadded(centered)

	cv.scrollContainer = container.NewScroll(paddedContent)
	cv.scrollContainer.SetMinSize(fyne.NewSize(860, 520))
	cv.scrollContainer.OnScrolled = func(pos fyne.Position) {
		if cv.adjustingScroll {
			cv.lastScroll = pos
			return
		}
		if cv.lastScroll == (fyne.Position{}) {
			cv.lastScroll = pos
			return
		}
		delta := pos.Y - cv.lastScroll.Y
		if delta == 0 {
			return
		}
		boost := delta * 1.2
		newOffset := fyne.NewPos(pos.X, pos.Y+boost)
		maxY := float32(0)
		if cv.scrollContainer.Content != nil {
			contentHeight := cv.scrollContainer.Content.Size().Height
			viewHeight := cv.scrollContainer.Size().Height
			if contentHeight > viewHeight {
				maxY = contentHeight - viewHeight
			}
		}
		if newOffset.Y < 0 {
			newOffset.Y = 0
		}
		if newOffset.Y > maxY {
			newOffset.Y = maxY
		}
		cv.adjustingScroll = true
		cv.scrollContainer.ScrollToOffset(newOffset)
		cv.adjustingScroll = false
		cv.lastScroll = newOffset
	}
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
	if role == "assistant" {
		cv.lastAssistant = bubble.GetContainer()
	}
	cv.scrollContainer.ScrollToBottom()
}

func (cv *ChatView) Clear() {
	cv.messages.Objects = nil
	cv.messageWidgets = nil
	cv.currentRole = ""
	cv.statusLine = nil
	cv.lastAssistant = nil
	cv.lastScroll = fyne.Position{}
	cv.messages.Refresh()
}

func (cv *ChatView) UpdateLastMessage(content string) {
	if len(cv.messageWidgets) == 0 {
		return
	}
	cv.messageWidgets[len(cv.messageWidgets)-1].UpdateContent(content)
	cv.scrollContainer.ScrollToBottom()
}

func (cv *ChatView) RemoveLastAssistantIfEmpty() {
	if len(cv.messageWidgets) == 0 {
		return
	}
	last := cv.messageWidgets[len(cv.messageWidgets)-1]
	if cv.currentRole != "assistant" || last.Text() != "" {
		return
	}

	for i, obj := range cv.messages.Objects {
		if obj == last.GetContainer() {
			cv.messages.Objects = append(cv.messages.Objects[:i], cv.messages.Objects[i+1:]...)
			break
		}
	}
	cv.messageWidgets = cv.messageWidgets[:len(cv.messageWidgets)-1]
	if len(cv.messageWidgets) == 0 {
		cv.currentRole = ""
		cv.lastAssistant = nil
	} else {
		cv.currentRole = "assistant"
		cv.lastAssistant = cv.messageWidgets[len(cv.messageWidgets)-1].GetContainer()
	}
	cv.messages.Refresh()
}

func (cv *ChatView) SetStatus(content string) {
	if cv.statusLine == nil {
		cv.statusLine = NewStatusLine(content)
		cv.insertBeforeAssistant(cv.statusLine.Container())
		cv.scrollContainer.ScrollToBottom()
		return
	}
	cv.statusLine.SetText(content)
}

func (cv *ChatView) ClearStatus() {
	if cv.statusLine == nil {
		return
	}
	for i, obj := range cv.messages.Objects {
		if obj == cv.statusLine.Container() {
			cv.messages.Objects = append(cv.messages.Objects[:i], cv.messages.Objects[i+1:]...)
			break
		}
	}
	cv.statusLine = nil
	cv.messages.Refresh()
}

func (cv *ChatView) AddNote(content string) {
	line := NewStatusLine(content)
	cv.insertBeforeAssistant(line.Container())
	cv.scrollContainer.ScrollToBottom()
}

func (cv *ChatView) insertBeforeAssistant(obj fyne.CanvasObject) {
	if cv.lastAssistant == nil {
		cv.messages.Add(obj)
		return
	}

	idx := -1
	for i, existing := range cv.messages.Objects {
		if existing == cv.lastAssistant {
			idx = i
			break
		}
	}
	if idx == -1 {
		cv.messages.Add(obj)
		return
	}

	objects := append([]fyne.CanvasObject{}, cv.messages.Objects...)
	objects = append(objects[:idx], append([]fyne.CanvasObject{obj}, objects[idx:]...)...)
	cv.messages.Objects = objects
	cv.messages.Refresh()
}
