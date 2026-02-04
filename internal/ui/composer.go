package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Composer struct {
	onSubmit func(string)
	entry    *widget.Entry
	sendBtn  *widget.Button
}

func NewComposer(onSubmit func(string)) *Composer {
	c := &Composer{onSubmit: onSubmit}

	c.entry = widget.NewMultiLineEntry()
	c.entry.SetPlaceHolder("Message...")
	c.entry.Wrapping = fyne.TextWrapWord
	c.entry.SetMinRowsVisible(2)
	c.entry.OnSubmitted = func(_ string) {
		c.send()
	}

	c.sendBtn = widget.NewButtonWithIcon("", theme.MailSendIcon(), func() {
		c.send()
	})
	c.sendBtn.Importance = widget.HighImportance
	c.sendBtn.Disable()

	c.entry.OnChanged = func(content string) {
		if content == "" {
			c.sendBtn.Disable()
			return
		}
		c.sendBtn.Enable()
	}

	return c
}

func (c *Composer) Container() fyne.CanvasObject {
	bg := canvas.NewRectangle(VercelDarkGray)
	bg.CornerRadius = 10
	bg.StrokeColor = VercelGray
	bg.StrokeWidth = 1
	bg.SetMinSize(fyne.NewSize(860, 64))

	buttonWrap := container.NewGridWrap(fyne.NewSize(48, 48), c.sendBtn)
	inputSurface := container.NewBorder(nil, nil, nil, buttonWrap, c.entry)

	bar := container.NewStack(bg, container.NewPadded(inputSurface))
	maxWidth := container.New(&MaxWidthLayout{MaxWidth: 860, MinWidth: 860}, bar)

	return container.NewHBox(layout.NewSpacer(), maxWidth, layout.NewSpacer())
}

func (c *Composer) SetEnabled(enabled bool) {
	if enabled {
		c.entry.Enable()
		if c.entry.Text != "" {
			c.sendBtn.Enable()
		}
	} else {
		c.entry.Disable()
		c.sendBtn.Disable()
	}
}

func (c *Composer) send() {
	content := c.entry.Text
	if content == "" {
		return
	}
	c.onSubmit(content)
	c.entry.SetText("")
}
