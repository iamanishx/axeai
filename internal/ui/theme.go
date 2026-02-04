package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type VercelTheme struct {
	defaultTheme fyne.Theme
}

const (
	ChatTextSizeName fyne.ThemeSizeName = "chatText"
	ChatMetaSizeName fyne.ThemeSizeName = "chatMeta"
)

func NewVercelTheme() fyne.Theme {
	return &VercelTheme{
		defaultTheme: theme.DarkTheme(),
	}
}

func (v *VercelTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{0, 0, 0, 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{17, 17, 17, 255}
	case theme.ColorNameOverlayBackground:
		return color.RGBA{17, 17, 17, 240}
	case theme.ColorNameButton:
		return color.RGBA{30, 30, 30, 255}
	case theme.ColorNameDisabledButton:
		return color.RGBA{51, 51, 51, 255}
	case theme.ColorNameForeground:
		return color.RGBA{255, 255, 255, 255}
	case theme.ColorNameDisabled:
		return color.RGBA{136, 136, 136, 255}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{102, 102, 102, 255}
	case theme.ColorNamePrimary:
		return color.RGBA{0, 112, 243, 255}
	case theme.ColorNameFocus:
		return color.RGBA{0, 112, 243, 255}
	case theme.ColorNameSelection:
		return color.RGBA{0, 112, 243, 100}
	case theme.ColorNameHover:
		return color.RGBA{28, 28, 28, 255}
	case theme.ColorNameShadow:
		return color.RGBA{0, 0, 0, 100}
	case theme.ColorNameScrollBar:
		return color.RGBA{51, 51, 51, 200}
	case theme.ColorNameSeparator:
		return color.RGBA{15, 15, 15, 255}

	default:
		return v.defaultTheme.Color(name, variant)
	}
}

func (v *VercelTheme) Font(style fyne.TextStyle) fyne.Resource {
	return v.defaultTheme.Font(style)
}

func (v *VercelTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return v.defaultTheme.Icon(name)
}

func (v *VercelTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case ChatTextSizeName:
		return v.defaultTheme.Size(theme.SizeNameText) * 1.15
	case ChatMetaSizeName:
		return v.defaultTheme.Size(theme.SizeNameText) * 1.0
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 18
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInputBorder:
		return 0
	case theme.SizeNameInputRadius:
		return 6
	case theme.SizeNamePadding:
		return 3
	case theme.SizeNameInnerPadding:
		return 2
	case theme.SizeNameScrollBar:
		return 4
	case theme.SizeNameScrollBarSmall:
		return 2
	case theme.SizeNameSeparatorThickness:
		return 0.1
	default:
		return v.defaultTheme.Size(name)
	}
}

var (
	VercelBlack     = color.RGBA{0, 0, 0, 255}
	VercelDarkGray  = color.RGBA{17, 17, 17, 255}
	VercelGray      = color.RGBA{51, 51, 51, 255}
	VercelLightGray = color.RGBA{102, 102, 102, 255}
	VercelMuted     = color.RGBA{136, 136, 136, 255}
	VercelWhite     = color.RGBA{255, 255, 255, 255}
	VercelBlue      = color.RGBA{0, 112, 243, 255}
	VercelBlueHover = color.RGBA{0, 91, 198, 255}
	VercelPurple    = color.RGBA{124, 58, 237, 255}
	VercelSuccess   = color.RGBA{0, 200, 83, 255}
	VercelError     = color.RGBA{255, 0, 0, 255}
	VercelWarning   = color.RGBA{255, 170, 0, 255}
)
