package main

import (
	tea "charm.land/bubbletea/v2"
	"log"
)

func main() {
	muik := tea.NewProgram(StartModel())
	if _, err := muik.Run(); err != nil {
		log.Fatal(err.Error())
	}
}
