package tui

import (
	"fmt"
	"strings"
)

func renderBox(box string, data ...any) string {
	s := ""
	lines := strings.Split(box, "\n")

	for i, line := range lines {
		splitted := strings.Split(line, ":")
		s += itemStyle.Render(fmt.Sprintf("%s: ", splitted[0]))
		s += dataStyle.Render(fmt.Sprintf(fmt.Sprintf("%s", splitted[1]), data[i]))
		s += "\n"
	}

	return s
}
