package triageui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

type triageMode int

const (
	modeBrowse triageMode = iota
	modeDecline
	modeComplete
)

// triageDoneMsg is sent after an async Items.Update finishes.
type triageDoneMsg struct {
	err  error
	id   string
	verb string
}

// Model is the Bubble Tea model for interactive triage.
type Model struct {
	ctx    context.Context
	svcs   domain.Services
	sess   domain.Session
	items  []domain.Item
	trayNames   map[string]string
	displayByID map[string]string

	list list.Model
	mode triageMode
	input textinput.Model

	width  int
	height int
	errLine    string
	statusLine string
}

// New builds a triage model. items is copied.
func New(ctx context.Context, svcs domain.Services, sess domain.Session, items []domain.Item, trayNames map[string]string, displayByID map[string]string) Model {
	if trayNames == nil {
		trayNames = map[string]string{}
	}
	if displayByID == nil {
		displayByID = map[string]string{}
	}
	cp := append([]domain.Item(nil), items...)

	del := list.NewDefaultDelegate()
	del.SetSpacing(1)

	lst := list.New([]list.Item{}, del, 0, 0)
	lst.SetFilteringEnabled(false)
	lst.SetShowStatusBar(false)
	lst.SetShowPagination(false)
	lst.Title = "Pending"

	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 2000
	ti.Width = 72

	m := Model{
		ctx:         ctx,
		svcs:        svcs,
		sess:        sess,
		items:       cp,
		trayNames:   trayNames,
		displayByID: displayByID,
		list:        lst,
		input:       ti,
	}
	m.rebuildList()
	return m
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutList()
		m.input.Width = max(20, msg.Width-6)
		return m, nil

	case triageDoneMsg:
		if msg.err != nil {
			m.errLine = msg.err.Error()
			m.statusLine = ""
			return m, nil
		}
		m.errLine = ""
		m.statusLine = msg.verb
		m.removeByID(msg.id)
		if len(m.items) == 0 {
			return m, tea.Quit
		}
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeDecline:
			return m.updateDeclineInput(msg)
		case modeComplete:
			return m.updateCompleteInput(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			return m.submitStatus("Accepted", func(_ string) domain.ItemPatch {
				st := "accepted"
				return domain.ItemPatch{Status: &st}
			})
		case "d":
			return m.beginDecline()
		case "c":
			return m.beginComplete()
		case "r":
			return m.submitStatus("Archived", func(_ string) domain.ItemPatch {
				st := "archived"
				return domain.ItemPatch{Status: &st}
			})
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) layoutList() {
	header := 5
	footer := 3
	detail := 8
	h := m.height - header - footer - detail
	if h < 6 {
		h = 6
	}
	if m.width < 20 {
		m.width = 80
	}
	m.list.SetWidth(m.width)
	m.list.SetHeight(h)
}

func (m *Model) rebuildList() {
	lis := make([]list.Item, len(m.items))
	for i := range m.items {
		lis[i] = triageItem{
			it:          m.items[i],
			trayNames:   m.trayNames,
			displayByID: m.displayByID,
			sess:        m.sess,
		}
	}
	m.list.SetItems(lis)
}

func (m *Model) selected() (domain.Item, bool) {
	sel := m.list.SelectedItem()
	if sel == nil {
		return domain.Item{}, false
	}
	ti, ok := sel.(triageItem)
	if !ok {
		return domain.Item{}, false
	}
	return ti.it, true
}

func (m *Model) removeByID(id string) {
	idx := m.list.Index()
	var out []domain.Item
	for _, it := range m.items {
		if it.ID != id {
			out = append(out, it)
		}
	}
	m.items = out
	m.rebuildList()
	if len(m.items) == 0 {
		return
	}
	newIdx := idx
	if newIdx >= len(m.items) {
		newIdx = len(m.items) - 1
	}
	if newIdx < 0 {
		newIdx = 0
	}
	m.list.Select(newIdx)
}

func (m *Model) submitStatus(verbLabel string, patchFn func(string) domain.ItemPatch) (tea.Model, tea.Cmd) {
	it, ok := m.selected()
	if !ok {
		m.errLine = "no item selected"
		return m, nil
	}
	id := strings.TrimSpace(it.ID)
	if id == "" {
		m.errLine = "invalid item"
		return m, nil
	}
	patch := patchFn(id)
	return m, m.runPatch(id, verbLabel, patch)
}

func (m *Model) runPatch(id, verb string, patch domain.ItemPatch) tea.Cmd {
	return func() tea.Msg {
		err := m.svcs.Items.Update(m.ctx, m.sess, id, patch)
		return triageDoneMsg{err: err, id: id, verb: verb}
	}
}

func (m *Model) beginDecline() (tea.Model, tea.Cmd) {
	if _, ok := m.selected(); !ok {
		m.errLine = "no item selected"
		return m, nil
	}
	m.mode = modeDecline
	m.errLine = ""
	m.input.Reset()
	m.input.Placeholder = "Decline reason (optional)"
	m.input.Focus()
	return m, textinput.Blink
}

func (m *Model) beginComplete() (tea.Model, tea.Cmd) {
	if _, ok := m.selected(); !ok {
		m.errLine = "no item selected"
		return m, nil
	}
	m.mode = modeComplete
	m.errLine = ""
	m.input.Reset()
	m.input.Placeholder = "Completion note (optional)"
	m.input.Focus()
	return m, textinput.Blink
}

func (m *Model) updateDeclineInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeBrowse
		m.input.Blur()
		return m, nil
	case "enter":
		it, ok := m.selected()
		if !ok {
			m.mode = modeBrowse
			return m, nil
		}
		id := strings.TrimSpace(it.ID)
		reason := strings.TrimSpace(m.input.Value())
		st := "declined"
		patch := domain.ItemPatch{Status: &st}
		if reason != "" {
			patch.DeclineReason = &reason
		}
		m.mode = modeBrowse
		m.input.Blur()
		return m, m.runPatch(id, "Declined", patch)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) updateCompleteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeBrowse
		m.input.Blur()
		return m, nil
	case "enter":
		it, ok := m.selected()
		if !ok {
			m.mode = modeBrowse
			return m, nil
		}
		id := strings.TrimSpace(it.ID)
		msgText := strings.TrimSpace(m.input.Value())
		st := "completed"
		patch := domain.ItemPatch{Status: &st}
		if msgText != "" {
			patch.CompletionMessage = &msgText
		}
		m.mode = modeBrowse
		m.input.Blur()
		return m, m.runPatch(id, "Completed", patch)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	switch m.mode {
	case modeDecline:
		return m.viewInput("Decline — Enter confirm · Esc cancel")
	case modeComplete:
		return m.viewInput("Complete — Enter confirm · Esc cancel")
	default:
		return m.viewBrowse()
	}
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
)

func (m *Model) viewInput(title string) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("tray triage"))
	b.WriteString("\n\n")
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Enter confirm · Esc cancel"))
	return b.String()
}

func (m *Model) viewBrowse() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("tray triage — %d pending", len(m.items))))
	b.WriteString("\n")
	if m.statusLine != "" {
		b.WriteString(okStyle.Render(m.statusLine))
		b.WriteString("\n")
	}
	if m.errLine != "" {
		b.WriteString(errStyle.Render(m.errLine))
		b.WriteString("\n")
	}
	b.WriteString(m.list.View())
	b.WriteString("\n")

	it, ok := m.selected()
	if ok {
		for _, line := range detailLines(it, m.trayNames, m.displayByID, m.sess) {
			b.WriteString(helpStyle.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("a accept · d decline (reason) · c complete (note) · r archive · ↑/↓ move · q quit"))
	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
