package selector

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Tiryoh/rgw/internal/worktree"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	wt    worktree.Worktree
	dirty bool
}

func (i item) Title() string {
	branch := i.wt.Branch
	if i.wt.Detached {
		branch = "(detached)"
	}
	if i.dirty {
		return branch + " *"
	}
	return branch
}

func (i item) Description() string {
	parts := []string{i.wt.Path}
	if i.wt.HEAD != "" {
		head := i.wt.HEAD
		if len(head) > 8 {
			head = head[:8]
		}
		parts = append(parts, head)
	}
	return strings.Join(parts, "  ")
}

func (i item) FilterValue() string {
	return i.wt.Branch + " " + i.wt.Path
}

type model struct {
	list     list.Model
	selected *worktree.Worktree
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if sel, ok := m.list.SelectedItem().(item); ok {
				wt := sel.wt
				m.selected = &wt
			}
			m.quitting = true
			return m, tea.Quit
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	return docStyle.Render(m.list.View())
}

// Select presents an interactive list for worktree selection.
// Returns the selected Worktree or an error if cancelled.
func Select(worktrees []worktree.Worktree) (*worktree.Worktree, error) {
	items := make([]list.Item, len(worktrees))
	for i, wt := range worktrees {
		dirty, _ := worktree.IsDirty(wt.Path)
		items[i] = item{wt: wt, dirty: dirty}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a worktree"
	l.SetShowStatusBar(false)

	m := model{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("selector error: %w", err)
	}

	final := result.(model)
	if final.selected == nil {
		return nil, fmt.Errorf("selection cancelled")
	}
	return final.selected, nil
}
