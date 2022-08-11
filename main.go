package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FIXME: calc proper width and make it responsive
var docStyle = lipgloss.NewStyle().Margin(1, 2).Width(50)

func (i submission) Title() string       { return i.ProblemTitle }
func (i submission) Description() string { return i.Time + " Ago" }
func (i submission) FilterValue() string { return i.Title() }

type submissionsLoadCmd struct{ items []submission }

type model struct {
	list          list.Model
	spinner       spinner.Model
	isLoadingData bool
	note          string
	loadedNotes   map[string]string
}

func InitModel() model {
	return model{list: list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0), spinner: spinner.New(), isLoadingData: true, note: "", loadedNotes: make(map[string]string)}
}

func loadSubmissionsCmd() tea.Cmd {
	return func() tea.Msg {
		submissions, _ := GetAllSubmissions(time.Date(2022, time.May, 1, 0, 0, 0, 0, time.UTC))
		return submissionsLoadCmd{items: submissions}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(spinner.Tick, loadSubmissionsCmd())
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

			curQuestion, ok := m.list.Items()[m.list.Index()].(submission)
			note, noteOk := m.loadedNotes[curQuestion.Title_slug]
			if ok && noteOk {
				m.note = note
			} else {
				m.note = ""
			}
			return m, cmd
		case "enter":
			curQuestion, ok := m.list.Items()[m.list.Index()].(submission)
			if ok {
				note, ok := m.loadedNotes[curQuestion.Title_slug]
				if !ok {
					note = getNote(curQuestion.Title_slug)
					m.loadedNotes[curQuestion.Title_slug] = note
				}
				m.note = note
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case submissionsLoadCmd:
		var items []list.Item
		for _, sub := range msg.items {
			items = append(items, sub)
		}
		m.list.SetItems(items)
		m.isLoadingData = false
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var views []string
	if m.isLoadingData {
		views = append(views, docStyle.Render(fmt.Sprintf("\n\n   %s Loading...\n\n", m.spinner.View())))
	} else {
		views = append(views, docStyle.Render(m.list.View()))
	}
	note := lipgloss.NewStyle().
		Width(50).
		Height(m.list.Height()).
		Padding(2).Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderBackground(lipgloss.Color("63")).
		Render(m.note)
	// _ := lipgloss.PlaceHorizontal(80, lipgloss.Left, note)
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
	m.spinner.Spinner = spinner.Dot
	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m.list.Title = "Latest Submissions"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
