package main

import (
	"errors"
	"fmt"
	"os"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
)


type loadAudio struct {
	spinner spinner.Model
}

type audioLoaded struct {
	audioList *[]Audio
}

type audioLoadError struct {
	err error
}

func (l loadAudio) Update(msg tea.Msg) (loadAudio, tea.Cmd) {
	var cmd tea.Cmd

	l.spinner, cmd = l.spinner.Update(msg)

	return l, cmd
}

func (l loadAudio) Init() tea.Cmd {
	dirPath := "D:/Music"
	audioList := []Audio{}

	return tea.Batch(l.spinner.Tick, func() tea.Msg {

		err := readAudioPaths(&audioList, dirPath)

		if err != nil {
			return audioLoadError{
				err: err,
			}
		}

		return audioLoaded{
			audioList: &audioList,
		}
	})
}

func (l loadAudio) View() tea.View {

	var s string

	s = fmt.Sprintf("%v %v", l.spinner.View(), textStyle("Music loading \n"))
	s = lipgloss.NewStyle().Width(20).Render(s)

	return tea.NewView(s)

}

func readAudioPaths(audioList *[]Audio, dirPath string) error {

	fileEntries, err := os.ReadDir(dirPath)

	if err != nil {
		return err
	}

	// log.Printf("Folder lenght: %v \n", len(fileEntries))

	if len(fileEntries) == 0 {
		return errors.New("directory is empty")
	}

	for _, e := range fileEntries {
		info, err := e.Info()
		if err != nil {
			return err
		}
		path := fmt.Sprintf("%v/%v", dirPath, info.Name())

		if !e.IsDir() {
			// log.Printf("Path: %v\n", path)

			*audioList = append(*audioList, Audio{
				name:        info.Name(),
				absoluePath: path,
			})
		} else {
			readAudioPaths(audioList, path)
		}

	}

	return nil
}
