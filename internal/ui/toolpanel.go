package ui

import (
	"axe-desktop/pkg/models"
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ToolPanel displays tool calls and reasoning
type ToolPanel struct {
	tabs      *container.AppTabs
	toolList  *widget.List
	toolCalls []models.ToolCall
	debugLog  *widget.Entry
}

// NewToolPanel creates a new tool panel
func NewToolPanel() *ToolPanel {
	tp := &ToolPanel{
		toolCalls: []models.ToolCall{},
	}

	tp.debugLog = widget.NewMultiLineEntry()
	tp.debugLog.Wrapping = fyne.TextWrapWord
	tp.debugLog.Disable()

	tp.toolList = widget.NewList(
		func() int { return len(tp.toolCalls) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Tool Call")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(tp.toolCalls) {
				return
			}
			call := tp.toolCalls[id]
			label := item.(*widget.Label)
			label.SetText(fmt.Sprintf("%s: %s", call.ToolName, formatArgs(call.Args)))
		},
	)

	tp.toolList.OnSelected = func(id widget.ListItemID) {
		if id < len(tp.toolCalls) {
			// Show details in a dialog
			// TODO: Implement detail view
		}
	}

	tp.tabs = container.NewAppTabs(
		container.NewTabItem("Tool Calls", tp.toolList),
		container.NewTabItem("Reasoning", widget.NewLabel("Reasoning summary will appear here")),
		container.NewTabItem("Debug", tp.debugLog),
	)

	return tp
}

// Container returns the tool panel container
func (tp *ToolPanel) Container() fyne.CanvasObject {
	return tp.tabs
}

func (tp *ToolPanel) AppendDebug(line string) {
	if tp.debugLog == nil {
		return
	}

	fyne.Do(func() {
		if tp.debugLog.Text == "" {
			tp.debugLog.SetText(line)
			return
		}
		tp.debugLog.SetText(tp.debugLog.Text + "\n" + line)
	})
}

// UpdateToolCalls updates the displayed tool calls
func (tp *ToolPanel) UpdateToolCalls(calls []models.ToolCall) {
	tp.toolCalls = calls
	tp.toolList.Refresh()
}

func formatArgs(args map[string]any) string {
	data, _ := json.Marshal(args)
	return string(data)
}
