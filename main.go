package main

import (
	"fmt"
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
	textInput textinput.Model
	branch    string
	err       error
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

	result := strings.Split(branch, "/")

	// get first and second result into separate variables
	// check if result has more than 2 items
	if len(result) <= 2 {
		log.Fatal("Branch name should be in the format of <type>/<task-id>/short-message")
	}

	branchType := strings.ToUpper(result[0])
	ticketId := strings.ToUpper(result[1])

	ti := textinput.New()
	//ti.Width = 90
	//ti.ShowSuggestions = true
	ti.SetValue(fmt.Sprintf("[%s] [%s] ", branchType, ticketId))
	//ti.SetSuggestions([]string{fmt.Sprintf("[FIX] [%s] ", ticketId), fmt.Sprintf("[IMPR] [%s] ", ticketId)})
	ti.Focus()
	ti.CharLimit = 156
	//ti.Width = 20

	return model{
		textInput: ti,
		branch:    branch,
		err:       nil,
	}
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Execute git commit command
			commitCmd := exec.Command("git", "commit", "-m", m.textInput.Value())
			commitCmd.Stdout = os.Stdout
			commitCmd.Stderr = os.Stderr

			err := commitCmd.Run()
			if err != nil {
				return m, tea.Quit
			}

			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"Current branch: %s\n\nEnter commit message\n\n%s\n\n%s",
		m.branch,
		m.textInput.View(),
		"(esc to quit, enter to commit)",
	) + "\n"
}
