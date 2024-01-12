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

type (
	errMsg error
)

type model struct {
	inputLabel         string
	commitMessageInput textinput.Model
	flagsInput         textinput.Model
	branch             string
	err                error
	askingForFlags     bool
	stagedFiles        []string
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
	stagedFiles, errStaged := getStagedFiles()

	if errStaged != nil {
		log.Fatal(errStaged)
	}

	result := strings.Split(branch, "/")

	// get first and second result into separate variables
	// check if result has more than 2 items
	if len(result) <= 2 {
		log.Fatal("Branch name should be in the format of <type>/<task-id>/short-message")
	}

	branchType := strings.ToUpper(result[0])
	ticketId := strings.ToUpper(result[1])

	ti := textinput.New()
	ti.SetValue(fmt.Sprintf("[%s] [%s] ", branchType, ticketId))
	ti.Focus()
	ti.CharLimit = 156

	return model{
		inputLabel:         "Enter commit message:",
		commitMessageInput: ti,
		branch:             branch,
		err:                nil,
		stagedFiles:        stagedFiles,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.askingForFlags {
				// Execute git commit command with flags
				commitCmd := exec.Command("git", "commit", m.flagsInput.Value(), "-m", m.commitMessageInput.Value())
				commitCmd.Stdout = os.Stdout
				commitCmd.Stderr = os.Stderr

				err := commitCmd.Run()
				if err != nil {
					return m, tea.Quit
				}

				return m, tea.Quit
			} else {
				m.inputLabel = "Commit flags"
				// Ask for commit flags
				flagsInput := textinput.New()
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
				flagsInput.SetSuggestions(commitFlags)
				flagsInput.ShowSuggestions = true
				flagsInput.Focus()
				m.flagsInput = flagsInput
				m.askingForFlags = true
				return m, nil
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	if m.askingForFlags {
		m.flagsInput, cmd = m.flagsInput.Update(msg)
	} else {
		m.commitMessageInput, cmd = m.commitMessageInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	var inputView string
	if m.askingForFlags {
		inputView = m.flagsInput.View()
	} else {
		inputView = m.commitMessageInput.View()
	}

	currentCommitMessage := "Current commit message: " + m.commitMessageInput.Value() + " " + m.flagsInput.Value()

	stagedFiles := strings.Join(m.stagedFiles, "\n")

	currentBranch := lipgloss.NewStyle().Foreground(lipgloss.Color("#fbd87f")).SetString(fmt.Sprintf("Current branch: %s", m.branch))
	inputLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#b5f8fe")).SetString(m.inputLabel).Bold(true)
	inputStyle := lipgloss.NewStyle().SetString(inputView)

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n%s\n\n%s",
		currentBranch.Render(),
		currentCommitMessage,
		inputLabel.Render(),
		inputStyle.Render(),
		stagedFiles,
	) + "\n"
}

// cmds
func getCurrentGitBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error getting branch name: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func getStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error getting staged files: %v", err)
	}

	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}
