package ui

import "fyne.io/fyne/v2"

type MaxWidthLayout struct {
	MaxWidth float32
	MinWidth float32
}

func (l *MaxWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}

	obj := objects[0]
	min := obj.MinSize()
	width := size.Width
	if l.MaxWidth > 0 && width > l.MaxWidth {
		width = l.MaxWidth
	}
	if l.MinWidth > 0 && width < l.MinWidth {
		width = l.MinWidth
	}
	height := size.Height
	if height < min.Height {
		height = min.Height
	}
	obj.Resize(fyne.NewSize(width, height))
	obj.Move(fyne.NewPos(0, 0))
}

func (l *MaxWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	min := objects[0].MinSize()
	if l.MinWidth > 0 && min.Width < l.MinWidth {
		min.Width = l.MinWidth
	}
	if l.MaxWidth > 0 && min.Width > l.MaxWidth {
		min.Width = l.MaxWidth
	}
	return min
}
