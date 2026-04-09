package main

import (
	"log"
	"math/rand"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	LOAD_AUDIO_STATE = iota
	SHOW_LIST_STATE
	PLAY_PAUSE_STATE
)

type Audio struct {
	name        string
	absoluePath string
}

func (a Audio) Title() string {
	return a.name
}

func (a Audio) Description() string {
	return a.absoluePath
}

func (a Audio) FilterValue() string {
	return a.name
}

type muikModel struct {
	state        int
	windowHeight int
	windowWidth  int
	audioList    []Audio
	shuffleList  []Audio
	isShuffled   bool
	loadAudio    loadAudio
	audioUilist  AudioList
	player       Player
	currentView  string
	// viewport     viewport.Model
}

func StartModel() muikModel {

	loader := loadAudio{
		spinner: spinner.New(),
	}

	loader.spinner.Style = spinnerStyle
	loader.spinner.Spinner = spinner.MiniDot

	return muikModel{
		loadAudio: loader,
	}
}

func (m muikModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadAudio.Init(),
		m.player.Init(),
	)
}

func (m *muikModel) setSequenceList() {

	listUi := []list.Item{}
	for _, a := range m.audioList {
		v := list.Item(a)
		listUi = append(listUi, v)
	}
	m.audioUilist.list = list.New(listUi, list.NewDefaultDelegate(), m.windowWidth/2, m.windowHeight-5)
	m.audioUilist.list.Title = "Sequence Songs"
}

func (m *muikModel) setSuffleList() tea.Cmd {
	for _, n := range rand.Perm(len(m.audioList)) {
		m.shuffleList = append(m.shuffleList, m.audioList[n])
	}

	listUi := []list.Item{}

	for _, a := range m.shuffleList {
		listUi = append(listUi, list.Item(a))
	}

	m.audioUilist.list.Title = "Shuffled Songs"

	return m.audioUilist.list.SetItems(listUi)
}

func (m *muikModel) forNext(cmd tea.Cmd) tea.Cmd {

	// var cmd tea.Cmd
	if !m.isShuffled {
		if m.player.currentAudioIndex+1 < len(m.audioList)-1 {
			m.player, cmd = m.player.Update(PlayAll{
				cAudio:      m.audioList[m.player.currentAudioIndex+1],
				cAudioIndex: m.player.currentAudioIndex + 1,
			})
			return cmd
		}
	} else {
		if m.player.currentAudioIndex+1 < len(m.shuffleList)-1 {
			m.player, cmd = m.player.Update(PlayAll{
				cAudio:      m.shuffleList[m.player.currentAudioIndex+1],
				cAudioIndex: m.player.currentAudioIndex + 1,
			})
			m.audioUilist.list.Select(m.player.currentAudioIndex)
			return cmd
		}
	}
	return cmd
}

func (m *muikModel) forPrev(cmd tea.Cmd) tea.Cmd {

	// var cmd tea.Cmd
	if !m.isShuffled {
		if m.player.currentAudioIndex-1 > -1 {
			m.player, cmd = m.player.Update(PlayAll{
				cAudio:      m.audioList[m.player.currentAudioIndex-1],
				cAudioIndex: m.player.currentAudioIndex - 1,
			})
			m.audioUilist.list.Select(m.player.currentAudioIndex)
			return cmd
		}
	} else {
		if m.player.currentAudioIndex-1 > -1 {
			m.player, cmd = m.player.Update(PlayAll{
				cAudio:      m.shuffleList[m.player.currentAudioIndex-1],
				cAudioIndex: m.player.currentAudioIndex - 1,
			})
			m.audioUilist.list.Select(m.player.currentAudioIndex)
			return cmd
		}
	}
	return cmd
}

func (m muikModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.loadAudio, cmd = m.loadAudio.Update(msg)

	isFiltering := m.audioUilist.list.FilterState()

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if isFiltering != list.Filtering {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				a, ok := m.audioUilist.list.SelectedItem().(Audio)
				if ok {
					m.player, cmd = m.player.Update(PlayAudio{audio: a})
					return m, cmd
				}
			case "space":
				if m.player.streamer != nil {
					m.player, cmd = m.player.Update(ResumePause{})
					return m, cmd
				}
			case "a":
				m.isShuffled = false
				if len(m.audioList) > 0 {
					m.player, cmd = m.player.Update(PlayAll{
						cAudio:      m.audioList[0],
						cAudioIndex: 0,
					})
					return m, cmd
				}
			case "n":
				cmd = m.forNext(cmd)
				return m, cmd
			case "p":
				cmd = m.forPrev(cmd)
				return m, cmd
			case "s":
				m.isShuffled = true
				shuffleCmd := m.setSuffleList()
				var playerCmd tea.Cmd
				if len(m.shuffleList) > 0 {
					m.player, playerCmd = m.player.Update(PlayAll{
						cAudio:      m.shuffleList[0],
						cAudioIndex: 0,
					})
				}
				return m, tea.Batch(
					shuffleCmd,
					playerCmd,
				)
			}
		}
	case Next:
		cmd = m.forNext(cmd)
		return m, cmd

	case PlayerError:
		log.Fatal(msg.err.Error())

	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.windowWidth = msg.Width

	case audioLoaded:
		m.audioList = *msg.audioList
		m.setSequenceList()

	case audioLoadError:
		log.Fatalf("Cannot load audio: %v", msg.err.Error())

	}

	if m.audioUilist.list.Items() != nil {
		m.audioUilist, cmd = m.audioUilist.Update(msg)
	}

	var playerCmd tea.Cmd
	m.player, playerCmd = m.player.Update(msg)

	return m, tea.Batch(
		cmd,
		playerCmd,
	)
}

var (
	borderStyle = func(height int, width int, content string) string {
		return lipgloss.NewStyle().
			Width(width).
			Height(height).
			Border(lipgloss.RoundedBorder()).
			// Margin(1).
			Render(content)
	}

	loadAudioStyle = func(height int, width int, content string) string {
		return lipgloss.NewStyle().
			Height(height).
			Width(width).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Render(content)
	}

	leftContentStyle = func(height int, width int, content string) string {
		return lipgloss.NewStyle().
			Height(height).
			Width(width).
			Border(lipgloss.NormalBorder(),
				false, true, false, false).
			Render(content)
	}

	rightContentStyle = func(height int, width int, content string) string {
		return lipgloss.NewStyle().
			Height(height).
			Width(width).
			Align(lipgloss.Center).
			Render(content)
	}
)

func (m muikModel) View() tea.View {
	mainHeight := m.windowHeight - 1
	mainWidth := m.windowWidth - 2

	mainContent := ""

	if len(m.audioList) == 0 {
		m.state = LOAD_AUDIO_STATE
		mainContent = loadAudioStyle(mainHeight, mainWidth, m.loadAudio.View().Content)
	} else {
		contentHeight := mainHeight - 1
		contentWidth := mainWidth / 2
		m.state = SHOW_LIST_STATE
		l := leftContentStyle(contentHeight, contentWidth, m.audioUilist.View().Content)

		m.player.height = contentHeight
		m.player.width = contentWidth

		r := rightContentStyle(contentHeight, contentWidth, m.player.View().Content)
		mainContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			l,
			r,
		)
	}

	renderer := borderStyle(mainHeight, mainWidth, mainContent)
	v := tea.NewView(renderer)
	v.WindowTitle = "MuikTerm"
	v.AltScreen = true
	return v
}
