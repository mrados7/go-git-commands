package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"os/exec"
	"strings"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6FD0FB"))
	branchTypes  = []string{"FEAT", "FIX", "IMPR", "OPS"}
)

type model struct {
	branchType            string
	branchTypeCursor      int
	ticketIdInput         textinput.Model
	branchNameInput       textinput.Model
	step                  int
	placeholderBranchName string
}

func initialModel() model {
	m := model{
		branchType:       branchTypes[0],
		branchTypeCursor: 0,
	}

	// ---- ticketIdInput input -----
	var ticketIdInput textinput.Model

	ticketIdInput = textinput.New()
	ticketIdInput.SetValue("FE-XXXX")

	m.ticketIdInput = ticketIdInput

	// ---- branchName input -----
	var branchNameInput textinput.Model

	branchNameInput = textinput.New()
	branchNameInput.SetValue("short-message")

	m.branchNameInput = branchNameInput

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.step == 0 {
		m.ticketIdInput.Blur()
		m.branchNameInput.Blur()
	} else if m.step == 1 {
		m.ticketIdInput.Focus()
		m.branchNameInput.Blur()
	} else if m.step == 2 {
		m.ticketIdInput.Blur()
		m.branchNameInput.Focus()
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyRight:
			if m.step < 2 {
				m.step++
			}
		case tea.KeyLeft:
			if m.step > 0 {
				m.step--
			}
		case tea.KeyEnter:
			if m.step == 0 {
				m.branchType = branchTypes[m.branchTypeCursor]
			}
			if m.step < 2 {
				m.step++
			} else {
				checkoutCommand := exec.Command("git", "checkout", "-b", fmt.Sprintf("%s", getBranchName(m)))
				checkoutCommand.Stdout = os.Stdout
				checkoutCommand.Stderr = os.Stderr

				err := checkoutCommand.Run()
				if err != nil {
					fmt.Println(err)
					//return m, tea.Quit
				}

				return m, tea.Quit
			}

		case tea.KeyUp, tea.KeyDown:
			if m.step == 0 {
				if msg.Type == tea.KeyUp {
					if m.branchTypeCursor > 0 {
						m.branchTypeCursor--
					}
				} else {
					if m.branchTypeCursor < len(branchTypes)-1 {
						m.branchTypeCursor++
					}
				}
			}
		}
	}

	m.ticketIdInput, cmd = m.ticketIdInput.Update(msg)
	m.branchNameInput, cmd = m.branchNameInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("")
	b.WriteString(getBranchNameWithStyle(m))
	b.WriteRune('\n')
	b.WriteRune('\n')

	b.WriteString(getSelectedOptionTitle(m.step))
	b.WriteRune('\n')

	if m.step == 0 {
		b.WriteString(chooseBranchTypeView(m))
	} else if m.step == 1 {
		m.ticketIdInput.Focus()
		b.WriteString(m.ticketIdInput.View())
	} else if m.step == 2 {
		m.branchNameInput.Focus()
		b.WriteString(m.branchNameInput.View())
	}

	return b.String()
}

func getBranchName(m model) string {
	return fmt.Sprintf("%s/%s/%s", m.branchType, m.ticketIdInput.Value(), m.branchNameInput.Value())
}

func getBranchNameWithStyle(m model) string {
	var branchType string
	if m.step == 0 {
		branchType = focusedStyle.Render(resolveHighlightedBranchType(m.branchTypeCursor, m.branchType, m.step))
	} else {
		branchType = resolveHighlightedBranchType(m.branchTypeCursor, m.branchType, m.step)
	}

	var ticketId string
	if m.step == 1 {
		ticketId = focusedStyle.Render(m.ticketIdInput.Value())
	} else {
		ticketId = m.ticketIdInput.Value()
	}

	var branchName string
	if m.step == 2 {
		branchName = focusedStyle.Render(m.branchNameInput.Value())
	} else {
		branchName = m.branchNameInput.Value()
	}

	return fmt.Sprintf("%s/%s/%s", branchType, ticketId, branchName)
}

func chooseBranchTypeView(m model) string {
	var b strings.Builder

	for i := 0; i < len(branchTypes); i++ {
		if m.branchTypeCursor == i {
			b.WriteString(focusedStyle.Bold(true).Render("> "))
			b.WriteString(focusedStyle.Bold(true).Render(branchTypes[i]))
		} else {
			b.WriteString("> ")
			b.WriteString(branchTypes[i])
		}
		b.WriteString("\n")
	}

	return b.String()
}

func resolveHighlightedBranchType(branchTypeCursor int, branchType string, step int) string {
	if step != 0 {
		return branchType
	}
	return branchTypes[branchTypeCursor]
}

func getSelectedOptionTitle(selectedOption int) string {
	switch selectedOption {
	case 0:
		return "Select branch type:"
	case 1:
		return "Enter ticket id:"
	case 2:
		return "Enter short message:"
	default:
		return ""
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
