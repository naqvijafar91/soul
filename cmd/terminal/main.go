package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func main() {

	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		PaddingTop(2).
		PaddingLeft(4).
		Width(22)

	fmt.Println(style.Render("Hello, kitty."))
}
