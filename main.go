package main

import (
	tea "charm.land/bubbletea/v2"
	"log"
)

func main() {

	model := StartModel()
	modelPtr := &model

	muik := tea.NewProgram(modelPtr)

	modelPtr.setProgram(muik)

	if _, err := muik.Run(); err != nil {
		log.Fatal(err.Error())
	}

	if modelPtr.cavaCmd != nil && modelPtr.cavaCmd.Process != nil {
		err := modelPtr.cavaCmd.Process.Kill()

		if err != nil {
			log.Fatal(err.Error())
		}

	}

}
