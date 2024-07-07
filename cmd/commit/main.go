package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrados7/go-git-commands/cmd/git"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	focusedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#DA8BFF")).Bold(true)
	blurredStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle      = focusedStyle.Copy()
	noStyle          = lipgloss.NewStyle()
	stagedFilesStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10ffcb"))
	focusedButton    = focusedStyle.Copy().Render("[ Commit ]")
	blurredButton    = fmt.Sprintf("[ %s ]", blurredStyle.Render("Commit"))
)

type model struct {
	focusInputIndex int
	inputs          []textinput.Model
	branch          string
	err             error
	stagedFiles     []string
}

func main() {
	isGitRepo := git.CheckIfGitRepo()
	if !isGitRepo {
		log.Fatal("Not a git repository")
	}
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func initialModel() model {
	branch, err := git.GetCurrentGitBranch()
	if err != nil {
		log.Fatal(err)
	}

	stagedFiles := git.GetStagedFiles()

	if len(stagedFiles) == 0 {
		log.Fatal("No staged files found")
	}

	result := strings.Split(branch, "/")

	var branchType string
	var ticketId string

	if len(result) >= 2 {
		branchType = strings.ToUpper(result[0])
		ticketId = strings.ToUpper(result[1])
	}

	m := model{
		inputs:          make([]textinput.Model, 2),
		branch:          branch,
		err:             nil,
		focusInputIndex: 0,
		stagedFiles:     stagedFiles,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 72

		switch i {
		case 0:
			if branchType != "" {
				t.SetValue(fmt.Sprintf("[%s] ", branchType))
			}
			if ticketId != "" {
				t.SetValue(t.Value() + fmt.Sprintf("[%s] ", ticketId))
			}
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.ShowSuggestions = true
			commitFlags := []string{
				"--message",
				"--all",
				"--patch",
				"--reuse-message",
				"--amend",
				"--signoff",
				"--no-verify",
				"--allow-empty",
				"--no-edit",
			}
			t.SetSuggestions(commitFlags)
			t.Placeholder = "Commit flags"
		}

		m.inputs[i] = t
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		// Set focus to next input
		case tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusInputIndex == len(m.inputs) {
				// Execute git commit command with flags
				commitCmd := exec.Command("git", "commit", "-m", fmt.Sprintf("%s %s", m.inputs[0].Value(), m.inputs[1].Value()))
				commitCmd.Stdout = os.Stdout
				commitCmd.Stderr = os.Stderr

				err := commitCmd.Run()
				if err != nil {
					return m, tea.Quit
				}

				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusInputIndex--
			} else {
				m.focusInputIndex++
			}

			if m.focusInputIndex > len(m.inputs) {
				m.focusInputIndex = 0
			} else if m.focusInputIndex < 0 {
				m.focusInputIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusInputIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	s := ""

	stagedFilesView := ""
	stagedFilesView += lipgloss.JoinVertical(lipgloss.Top, "Staged files:")
	stagedFilesView += "\n"
	for index, elem := range m.stagedFiles {
		stagedFilesView += lipgloss.JoinVertical(lipgloss.Top, stagedFilesStyle.Render(fmt.Sprintf("• %s", elem)))
		if index != len(m.stagedFiles)-1 {
			stagedFilesView += "\n"
		}
	}

	currentBranch := lipgloss.NewStyle().Foreground(lipgloss.Color("#fcbda1")).SetString(fmt.Sprintf("Branch: %s", m.branch))

	s += lipgloss.JoinVertical(lipgloss.Top, currentBranch.Render())
	s += "\n\n"

	inputLabels := []string{"Commit message:", "Commit flags (optional):"}

	for i := range m.inputs {
		s += lipgloss.JoinVertical(1, inputLabels[i])
		s += "\n"
		s += lipgloss.JoinVertical(0.5, m.inputs[i].View())
		if i < len(m.inputs)-1 {
			s += "\n"
		}
	}

	button := &blurredButton
	if m.focusInputIndex == len(m.inputs) {
		button = &focusedButton
	}
	s += "\n\n"
	s += lipgloss.JoinVertical(1, *button)

	stagedFilesStyle := lipgloss.NewStyle().MarginLeft(8).Border(lipgloss.NormalBorder()).Padding(0, 1)

	f := lipgloss.JoinHorizontal(lipgloss.Left, s, stagedFilesStyle.Render(stagedFilesView))

	return f
}
