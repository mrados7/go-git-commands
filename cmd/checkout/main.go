package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrados7/go-git-commands/cmd/git"
	"os"
)

const (
	defaultWidth = 80
	listHeight   = 35
)

func (b jiraBoard) Title() string       { return b.Name }
func (b jiraBoard) Description() string { return b.D }
func (b jiraBoard) FilterValue() string { return b.Name }

func (p branchType) Title() string       { return p.T }
func (p branchType) Description() string { return p.D }
func (p branchType) FilterValue() string { return p.T }

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6FD0FB"))

	stepTitleStyle    = lipgloss.NewStyle().Background(lipgloss.Color("#DA8BFF")).Foreground(lipgloss.Color("#000000")).Padding(0, 1)
	listTitleBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Padding(1, 0)
	docStyle          = lipgloss.NewStyle().Margin(1, 2)

	inputStyle = lipgloss.NewStyle().Margin(1, 0)
)

type model struct {
	branchType            string
	ticketIdInput         textinput.Model
	branchNameInput       textinput.Model
	step                  int
	placeholderBranchName string
	// new model
	branchTypeList list.Model
	jiraBoardList  list.Model

	selectedBranchType  string
	selectedJiraBoard   string
	shouldEnterTicketId bool
	shouldEnterBranch   bool
}

func initialModel(branchTypes []list.Item, jiraBoards []list.Item) model {
	m := model{
		branchTypeList: list.New(branchTypes, list.NewDefaultDelegate(), defaultWidth, listHeight),
		jiraBoardList:  list.New(jiraBoards, list.NewDefaultDelegate(), defaultWidth, listHeight),
	}

	m.branchTypeList.Styles.Title = stepTitleStyle
	m.branchTypeList.Styles.TitleBar = listTitleBarStyle
	m.branchTypeList.Title = "Select branch type"
	m.branchTypeList.SetShowStatusBar(false)

	m.jiraBoardList.Styles.Title = stepTitleStyle
	m.jiraBoardList.Styles.TitleBar = listTitleBarStyle
	m.jiraBoardList.Title = "Select Jira board"
	m.jiraBoardList.SetShowStatusBar(false)

	// ---- ticketIdInput input -----
	var ticketIdInput textinput.Model

	ticketIdInput = textinput.New()
	ticketIdInput.Focus()

	m.ticketIdInput = ticketIdInput

	// ---- branchName input -----
	var branchNameInput textinput.Model

	branchNameInput = textinput.New()
	branchNameInput.Placeholder = "short-message"

	m.branchNameInput = branchNameInput

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.branchTypeList.SetSize(msg.Width-h, msg.Height-v)
		m.jiraBoardList.SetSize(msg.Width-h, msg.Height-v)
		return m, nil
	}
	switch {
	case m.selectedBranchType == "":
		return m.UpdateBranchTypeList(msg)
	case m.selectedJiraBoard == "":
		return m.UpdateJiraBoardList(msg)
	case m.shouldEnterTicketId:
		return m.UpdateTicketIdInput(msg)
	case m.shouldEnterBranch:
		return m.UpdateBranchNameInput(msg)
	}
	return m, cmd
}

func (m model) UpdateBranchNameInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, _ := docStyle.GetFrameSize()
		m.branchNameInput.Width = msg.Width - h
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			err := git.CheckoutNewBranch(getBranchName(m))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			m.shouldEnterBranch = false
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.branchNameInput, cmd = m.branchNameInput.Update(msg)
	return m, cmd
}

func (m model) UpdateTicketIdInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, _ := docStyle.GetFrameSize()
		m.ticketIdInput.Width = msg.Width - h
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.ticketIdInput.Blur()
			m.shouldEnterTicketId = false
			m.shouldEnterBranch = true
			m.branchNameInput.Focus()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.ticketIdInput, cmd = m.ticketIdInput.Update(msg)
	return m, cmd
}

func (m model) UpdateBranchTypeList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.branchTypeList.SetSize(msg.Width-h, msg.Height-v)
		return m, nil
	case tea.KeyMsg:
		if m.branchTypeList.FilterState() == list.Filtering {
			break
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter, tea.KeySpace:
			m.selectedBranchType = m.branchTypeList.SelectedItem().(branchType).Title()
		}
	}

	var cmd tea.Cmd
	m.branchTypeList, cmd = m.branchTypeList.Update(msg)
	return m, cmd
}

func (m model) UpdateJiraBoardList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.jiraBoardList.SetSize(msg.Width-h, msg.Height-v)
		return m, nil
	case tea.KeyMsg:
		if m.jiraBoardList.FilterState() == list.Filtering {
			break
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter, tea.KeySpace:
			m.selectedJiraBoard = m.jiraBoardList.SelectedItem().(jiraBoard).Title()
			m.ticketIdInput.SetValue(m.jiraBoardList.SelectedItem().(jiraBoard).Title() + "-")
			m.shouldEnterTicketId = true
		}
	}

	var cmd tea.Cmd
	m.jiraBoardList, cmd = m.jiraBoardList.Update(msg)
	return m, cmd
}

func (m model) View() string {
	s := ""

	branchName := lipgloss.NewStyle().MarginBottom(1).Foreground(lipgloss.Color("#fcbda1")).Render(getBranchName(m))

	switch {
	case m.selectedBranchType == "":
		s += lipgloss.JoinVertical(lipgloss.Top, branchName, m.branchTypeList.View())
	case m.selectedJiraBoard == "":
		s += lipgloss.JoinVertical(lipgloss.Top, branchName, m.jiraBoardList.View())
	case m.shouldEnterTicketId:
		title := stepTitleStyle.Render("Enter ticket id")
		s += lipgloss.JoinVertical(lipgloss.Top, branchName, title, inputStyle.Render(m.ticketIdInput.View()))
	case m.shouldEnterBranch:
		title := stepTitleStyle.Render("Enter branch name")
		s += lipgloss.JoinVertical(lipgloss.Top, branchName, title, inputStyle.Render(m.branchNameInput.View()))
	}

	return s
}

func getBranchName(m model) string {
	return fmt.Sprintf("%s/%s/%s", m.selectedBranchType, m.ticketIdInput.Value(), m.branchNameInput.Value())
}

func main() {
	isGitRepo := git.CheckIfGitRepo()
	if !isGitRepo {
		fail("Error: not a git repository")
	}
	branchTypes, jiraBoards, err := loadConfig()
	if err != nil {
		fail("Error: %s", err)
	}

	p := tea.NewProgram(initialModel(branchTypes, jiraBoards), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func fail(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
