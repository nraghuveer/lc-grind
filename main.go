package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nraghuveer/lc-grind/lc_api"
	lc "github.com/nraghuveer/lc-grind/lc_api"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type progressLoadCmd struct { items []*lc.ProgressQuestion }
type progressMsg float32

type model struct {
	list          list.Model
	progress      float32
	progressChan  chan float32
	progressBar   progress.Model
	isLoadingData bool
	note          string
	loadedNotes   map[string]string
}

func InitModel() model {
	return model{list: list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0), isLoadingData: true, note: "", loadedNotes: make(map[string]string), progressBar: progress.New(progress.WithDefaultGradient()), progress: 0.0, progressChan: make(chan float32)}
}

func fetchProgress(m *model) tea.Cmd {
	return func() tea.Msg {
		progress := &lc_api.Progress{}
		err := progress.Init()
		if err != nil {
			return err
		}
		for progress.HasNext() {
			progress.FetchNext()
			m.progressChan <- progress.CompletedPercentage()
		}
		iter := progress.CreateIterator()
		items := make([]*lc.ProgressQuestion, 0)
		for iter.HasNext() {
			value, _ := iter.Next()
			items = append(items, value)
		}
		return progressLoadCmd{items: items}
	}
}

func waitForProgressUpdate(c <-chan float32) tea.Cmd {
	return func() tea.Msg {
		return progressMsg(<-c)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchProgress(&m), waitForProgressUpdate(m.progressChan))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k", "down", "j":
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)

			curQuestion, ok := m.list.Items()[m.list.Index()].(*lc.ProgressQuestion)
			note, noteOk := m.loadedNotes[curQuestion.QuestionTitle]
			if ok && noteOk {
				m.note = note
			} else {
				m.note = ""
			}
			return m, cmd
		case "enter":
			curQuestion, ok := m.list.Items()[m.list.Index()].(*lc.ProgressQuestion)
			if ok {
				note, ok := m.loadedNotes[curQuestion.QuestionTitle]
				if !ok {
					note = lc.GetNote(curQuestion.QuestionTitle)
					m.loadedNotes[curQuestion.QuestionTitle] = note
				}
				m.note = note
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case progressLoadCmd:
		var items []list.Item
		for _, sub := range msg.items {
			items = append(items, sub)
		}
		m.list.SetItems(items)
		m.isLoadingData = false
	case progressMsg:
		m.progress = float32(msg)
		cmd := m.progressBar.SetPercent(float64(m.progress))
		return m, tea.Batch(cmd, waitForProgressUpdate(m.progressChan))
	case progress.FrameMsg:
		progressModel, cmd := m.progressBar.Update(msg)
		m.progressBar = progressModel.(progress.Model)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var views []string
	if m.isLoadingData {
		views = append(views, docStyle.Render(m.progressBar.View()))
	} else {
		views = append(views, docStyle.Render(m.list.View()))
	}
	note := lipgloss.NewStyle().
		Width(50).
		Height(m.list.Height()).
		Padding(2).Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderBackground(lipgloss.Color("63")).
		Render(m.note)
	views = append(views, note)
	return lipgloss.JoinHorizontal(lipgloss.Center, views...)
}

func main() {
	db, err := GetDB()
	if err != nil {
		log.Fatalln("Failed to create db instance", err.Error())
	}
	defer db.Close()
	m := InitModel()
	m.list.Title = "Latest Submissions"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
