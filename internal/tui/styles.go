package tui

import "github.com/charmbracelet/lipgloss"

var messageStyle = lipgloss.NewStyle().
	Width(93).
	Padding(0, 2, 0, 2).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#767676"))

var leftBoxContainer = lipgloss.NewStyle().
	Width(40).
	Height(8).
	Padding(0, 2, 0, 2).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#767676"))

var rightBoxContainer = lipgloss.NewStyle().
	Width(51).
	Height(8).
	Padding(0, 2, 0, 2).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#767676"))

var ttlStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#767676"))

var errStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#D70040")).
	Background(lipgloss.Color("235"))

var itemStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#767676"))

var dataStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("204")).
	Background(lipgloss.Color("235"))

var focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffff"))
