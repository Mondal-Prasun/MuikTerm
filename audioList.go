package main

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type AudioList struct {
	list list.Model
}

func (al AudioList) Init() tea.Cmd {
	return nil
}

func (al AudioList) Update(msg tea.Msg) (AudioList, tea.Cmd) {
	var cmd tea.Cmd

	al.list, cmd = al.list.Update(msg)

	return al, cmd
}

func (al AudioList) View() tea.View {

	v := lipgloss.NewStyle().Margin(1, 2).Render(al.list.View())
	return tea.NewView(v)
}
