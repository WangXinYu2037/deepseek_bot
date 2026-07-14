package utils

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var renderer *glamour.TermRenderer

func init() {
	var err error
	renderer, err = glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		panic(err)
	}
}

func RenderMarkdown(text string) string {
	if text == "" {
		return ""
	}

	result, err := renderer.Render(text)
	if err != nil {
		return text
	}

	return lipgloss.NewStyle().MarginTop(1).MarginBottom(1).Render(result)
}
