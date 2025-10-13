package selector

import (
	"github.com/atolix/clyst/theme"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func NewStyleDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.NormalTitle = d.Styles.NormalTitle.PaddingLeft(2)
	d.Styles.NormalDesc = d.Styles.NormalDesc.PaddingLeft(2)

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(theme.Primary).
		MarginLeft(0).
		PaddingLeft(2).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.Border).
		Bold(true)

	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(theme.Primary).
		MarginLeft(0).
		PaddingLeft(2).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.Border)

	return d
}
