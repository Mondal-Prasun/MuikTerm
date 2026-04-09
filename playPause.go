package main

import (
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

type Player struct {
	height               int
	width                int
	currentAudio         Audio
	currentAudioIndex    int
	currentAudioPos      time.Duration
	currentAudioDuration time.Duration
	isPlaying            bool
	progressbar          progress.Model
	streamer             beep.StreamSeekCloser
	format               beep.Format
	ctrl                 *beep.Ctrl
}

func (p *Player) Init() tea.Cmd {

	p.progressbar = progress.New(progress.WithColors(lipgloss.Green, lipgloss.Magenta))

	return tea.Batch(func() tea.Msg {
		err := speaker.Init(beep.SampleRate(44100), beep.SampleRate(44100).N(time.Second/10))
		if err != nil {
			return PlayerError{
				err: err,
			}
		}
		return nil
	}, p.progressbar.Init())
}

func (p *Player) runAudio(audio Audio, playSeq bool, isDone chan<- bool) error {

	if p.streamer != nil {
		p.streamer.Close()
		speaker.Clear()
	}

	f, err := os.Open(audio.absoluePath)

	if err != nil {
		return err
	}

	splitted := strings.Split(audio.name, ".")

	switch splitted[len(splitted)-1] {
	case "flac":
		p.streamer, p.format, err = flac.Decode(f)
		if err != nil {
			return err
		}
	case "wav":
		p.streamer, p.format, err = wav.Decode(f)
		if err != nil {
			return err
		}
	case "mp3":
		p.streamer, p.format, err = mp3.Decode(f)
		if err != nil {
			return err
		}
	}

	p.currentAudioDuration = p.format.SampleRate.D(p.streamer.Len())

	p.ctrl = &beep.Ctrl{Streamer: p.streamer, Paused: false}

	if playSeq {
		speaker.Play(beep.Seq(p.ctrl, beep.Callback(func() {
			p.streamer.Close()
			f.Close()
			isDone <- true
		})))
	} else {
		speaker.Play(beep.Seq(p.ctrl, beep.Callback(func() {
			p.streamer.Close()
			f.Close()
		})))
	}
	return nil
}

func (p *Player) getAudioPos() tea.Cmd {

	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		if !p.isPlaying {
			return nil
		}

		speaker.Lock()
		pos := p.format.SampleRate.D(p.streamer.Position())
		speaker.Unlock()

		return AudioCurrentPlayPos{
			second: pos,
		}
	})
}

func (p *Player) ResumePauseAudio() func() tea.Msg {
	speaker.Lock()
	p.ctrl.Paused = !p.ctrl.Paused
	speaker.Unlock()

	p.isPlaying = !p.ctrl.Paused

	return func() tea.Msg {
		return Playing{
			isPlaying: !p.ctrl.Paused,
		}
	}
}

type Playing struct {
	isPlaying bool
}

type ResumePause struct {
}

type Next struct {
}

type AudioCurrentPlayPos struct {
	second time.Duration
}

type PlayAll struct {
	cAudioIndex int
	cAudio      Audio
}

type PlayerError struct {
	err error
}

type PlayAudio struct {
	audio Audio
}

func (p *Player) Update(msg tea.Msg) (Player, tea.Cmd) {

	switch msg := msg.(type) {
	case PlayAudio:
		p.currentAudio = msg.audio
		err := p.runAudio(msg.audio, false, nil)
		if err == nil {
			// log.Println("err is not nil")
			p.isPlaying = !p.ctrl.Paused
			return *p, func() tea.Msg {
				return Playing{
					isPlaying: true,
				}
			}
		} else {
			return *p, func() tea.Msg {
				return PlayerError{
					err: err,
				}
			}
		}
	case ResumePause:
		return *p, p.ResumePauseAudio()
	case PlayAll:
		p.currentAudio = msg.cAudio
		p.currentAudioIndex = msg.cAudioIndex
		done := make(chan bool)
		err := p.runAudio(msg.cAudio, true, done)

		if err != nil {
			// log.Println("err is not nil")
			return *p, func() tea.Msg {
				return PlayerError{err: err}
			}
		}
		p.isPlaying = !p.ctrl.Paused
		return *p, func() tea.Msg {
			<-done
			return Next{}
		}
	case Playing:
		if msg.isPlaying {
			return *p, p.getAudioPos()
		}

	case AudioCurrentPlayPos:
		p.currentAudioPos = msg.second.Round(time.Second)
		return *p, p.getAudioPos()

	}

	return *p, nil
}

var (
	footerStyle = func(bindings []string, height int) string {
		fS := strings.Join(bindings, " • ")
		fS = lipgloss.NewStyle().
			Height(height).
			Foreground(lipgloss.Color("241")).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			Align(lipgloss.Left, lipgloss.Center).
			// AlignHorizontal(lipgloss.Center).
			Render(fS)
		return fS
	}

	audioTitleStyle = func(height int, content string) string {
		return lipgloss.NewStyle().
			Height(height).
			Foreground(lipgloss.Cyan).
			Margin(1).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder(), true, true, true, true).
			Align(lipgloss.Left, lipgloss.Center).
			Render(content)
	}

	progressbarStyle = func(height int, content string) string {
		return lipgloss.NewStyle().
			// Height(height).
			Align(lipgloss.Left, lipgloss.Center).
			Render(content)
	}
)

func (p *Player) View() tea.View {
	footerHeight := 2
	audioTitleHeight := 2

	content := lipgloss.NewStyle().
		Height(p.height - (footerHeight + audioTitleHeight) - 6).
		Render("Test")

	footer := footerStyle([]string{
		"↩ Play",
		"space resume/Pause",
		"a Play All",
		"s Shuffle",
		"n Next",
		"p Prev",
	}, footerHeight)

	title := "Nothing"

	if p.currentAudio.name != "" {
		title = p.currentAudio.name
	}

	audioTitle := audioTitleStyle(audioTitleHeight, title)

	// p.progressbar.SetWidth(10)

	// showPercent := 0.0
	showPercent := p.currentAudioPos.Seconds() / p.currentAudioDuration.Seconds()

	audioProgressBar := p.progressbar.ViewAs(showPercent)
	s := lipgloss.JoinVertical(
		lipgloss.Center,
		content,
		audioProgressBar,
		audioTitle,
		footer,
	)

	return tea.NewView(s)
}
