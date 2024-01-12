package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	focusedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6FD0FB")).Bold(true)
	blurredStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle      = focusedStyle.Copy()
	noStyle          = lipgloss.NewStyle()
	stagedFilesStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10ffcb"))
	focusedButton    = focusedStyle.Copy().Render("[ Commit ]")
	blurredButton    = fmt.Sprintf("[ %s ]", blurredStyle.Render("Commit"))
)

type (
	errMsg error
)

type model struct {
	focusInputIndex int
	inputs          []textinput.Model
	branch          string
	err             error
	stagedFiles     []string
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func initialModel() model {
	branch, err := getCurrentGitBranch()
	if err != nil {
		log.Fatal(err)
	}

	stagedFiles := getStagedFiles()

	if len(stagedFiles) == 0 {
		log.Fatal("No staged files found")
	}

	result := strings.Split(branch, "/")

	// get first and second result into separate variables
	// check if result has more than 2 items
	if len(result) <= 2 {
		log.Fatal("Branch name should be in the format of <type>/<task-id>/short-message")
	}

	branchType := strings.ToUpper(result[0])
	ticketId := strings.ToUpper(result[1])

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
			t.SetValue(fmt.Sprintf("[%s] [%s] ", branchType, ticketId))
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

func getCurrentGitBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error getting branch name: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
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
				commitCmd := exec.Command("git", "commit", m.inputs[0].Value(), "-m", m.inputs[1].Value())
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
	var b strings.Builder

	currentBranch := lipgloss.NewStyle().Foreground(lipgloss.Color("#fcbda1")).SetString(fmt.Sprintf("Branch: %s", m.branch))

	b.WriteString(currentBranch.Render())
	b.WriteRune('\n')
	b.WriteRune('\n')

	currentCommitCommand := lipgloss.NewStyle().Foreground(lipgloss.Color("#a1e0fc")).SetString(fmt.Sprintf("git commit -m \"%s\" %s", m.inputs[0].Value(), m.inputs[1].Value()))

	b.WriteString(currentCommitCommand.Render())
	b.WriteRune('\n')
	b.WriteRune('\n')

	inputLabels := []string{"Commit message ", "Commit flags "}

	for i := range m.inputs {
		if i == m.focusInputIndex {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6FD0FB")).Bold(true).SetString(inputLabels[i]).Render())
		} else {
			b.WriteString(inputLabels[i])
		}
		b.WriteString(m.inputs[i].View())
		b.WriteRune('\n')
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusInputIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString("Staged changes: ")
	b.WriteRune('\n')
	b.WriteString(stagedFilesStyle.Render(strings.Join(m.stagedFiles, "\n")))

	b.WriteRune('\n')

	return b.String()
}

func getStagedFiles() []string {
	cmd := exec.Command("git", "diff", "--name-only", "--cached")
	out, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	if len(out) == 0 {
		return []string{}
	}

	return strings.Split(string(out), "\n")
}
