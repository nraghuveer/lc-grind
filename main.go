package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	utils "github.com/nraghuveer/lc-grind/app"
	"github.com/nraghuveer/lc-grind/lc_api"
	lc "github.com/nraghuveer/lc-grind/lc_api"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)
var progressListStyle = lipgloss.NewStyle().Inherit(docStyle).Border(lipgloss.RoundedBorder()).BorderForeground((lipgloss.Color("238")))
var noteHeadingStyle = lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4"))
var noteStyle = lipgloss.NewStyle().Inherit(docStyle).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238"))

type progressLoadCmd struct{ items []*lc.ProgressQuestion }
type progressMsg float32

type model struct {
	width         int
	height        int
	list          list.Model
	progress      float32
	progressChan  chan float32
	progressBar   progress.Model
	isLoadingData bool
	note          string
	loadedNotes   map[string]string
	Tabs []string
	TabContent []string
	activeTab int
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
			note, noteOk := m.loadedNotes[curQuestion.ParseTitleSlug()]
			if ok && noteOk {
				m.note = note
			} else {
				m.note = ""
			}
			return m, cmd
		case "ctrl+enter":
			curQuestion, ok := m.list.Items()[m.list.Index()].(*lc.ProgressQuestion)
			if ok {
				if err := utils.OpenUrlInBrowser(curQuestion.URL); err != nil {
					log.Printf("Failed to open the url in browser - %s", err)
				}
			}
		case "enter":
			curQuestion, ok := m.list.Items()[m.list.Index()].(*lc.ProgressQuestion)
			if ok {
				// FIXME: we need title slug not the title
				note, ok := m.loadedNotes[curQuestion.QuestionTitle]
				if !ok {
					note, _ = lc.GetNote(curQuestion.ParseTitleSlug())
					m.loadedNotes[curQuestion.ParseTitleSlug()] = note
				}
				m.note = note
			}
			case "ctrl+h", "right":
				m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
				return m, nil
			case "ctrl+l", "left":
				m.activeTab = max(m.activeTab-1, 0)
				return m, nil
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

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

	lh, lw := docStyle.GetFrameSize()
	m.list.SetSize((m.width*2/3)-lw, m.height-lh)
	progressListStyle = progressListStyle.Width((m.width * 2 / 3) - lw).Height(m.height - lh)
	noteStyle = noteStyle.Width((m.width * 1 / 3) - lw).Height(m.height - lh)

	if m.isLoadingData {
		views = append(views, progressListStyle.Render(m.progressBar.View()))
	} else {
		views = append(views, progressListStyle.Render(m.list.View()))
	}
	noteHeading := noteHeadingStyle.Padding(0,1,0,1).Margin(0,0,3,1).Align(lipgloss.Center).Render("Notes")
	note := noteStyle.Render(noteHeading + "\n\n" + m.note)

	views = append(views, note)
	return lipgloss.JoinHorizontal(lipgloss.Center, views...)
}

func main() {
	db, err := GetDB()
	if err != nil {
		log.Fatalln("Faliled to create db instance", err.Error())
	}
	defer db.Close()
	m := InitModel()
	m.list.Title = "Progress"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
