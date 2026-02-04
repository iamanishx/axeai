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

	c.sendBtn = widget.NewButtonWithIcon("", theme.MailSendIcon(), func() {
		content := c.entry.Text
		if content != "" {
			c.onSubmit(content)
			c.entry.SetText("")
		}
	})
	c.sendBtn.Importance = widget.HighImportance

	return c
}

func (c *Composer) Container() fyne.CanvasObject {
	bg := canvas.NewRectangle(VercelDarkGray)
	bg.CornerRadius = 8
	bg.SetMinSize(fyne.NewSize(760, 60))

	inputSurface := container.NewBorder(nil, nil, nil, c.sendBtn, c.entry)

	bar := container.NewStack(bg, container.NewPadded(inputSurface))
	maxWidth := container.New(&MaxWidthLayout{MaxWidth: 760, MinWidth: 760}, bar)

	return container.NewHBox(layout.NewSpacer(), maxWidth, layout.NewSpacer())
}

func (c *Composer) SetEnabled(enabled bool) {
	if enabled {
		c.entry.Enable()
		c.sendBtn.Enable()
	} else {
		c.entry.Disable()
		c.sendBtn.Disable()
	}
}
