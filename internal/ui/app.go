package ui

import (
	"timesheet/internal/db"

	tea "github.com/charmbracelet/bubbletea"
)

// Application modes
type AppMode int

const (
	TimesheetMode AppMode = iota
	FormMode
)

// AppModel is the top-level model that contains both timesheet and form models
type AppModel struct {
	Mode          AppMode
	TimesheetView TimesheetModel
	FormView      FormModel
}

// NewAppModel creates a new app model with timesheet as the default view
func NewAppModel() AppModel {
	return AppModel{
		Mode:          TimesheetMode,
		TimesheetView: InitialTimesheetModel(),
	}
}

func (m AppModel) Init() tea.Cmd {
	// Initialize the current mode
	if m.Mode == TimesheetMode {
		return m.TimesheetView.Init()
	}
	return m.FormView.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle global keys first
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Global quit handler
		if keyMsg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	// Handle mode-specific updates
	switch m.Mode {
	case TimesheetMode:
		// Special handling for switching to form mode
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "a" {
				m.Mode = FormMode
				// Initialize a fresh form model
				m.FormView = InitialFormModel()
				return m, m.FormView.Init()
			}
		}

		// Handle edit entry message
		if editMsg, ok := msg.(EditEntryMsg); ok {
			// Switch to form mode for editing
			m.Mode = FormMode

			// Initialize the form for editing
			date := editMsg.Date
			m.FormView = InitialFormModelWithDate(date)

			// Try to load existing data
			entry, err := db.GetTimesheetEntryByDate(date)
			if err == nil {
				// Entry found, populate form fields
				m.FormView.prefillFromEntry(entry)
				m.FormView.isEditing = true
			}

			return m, m.FormView.Init()
		}

		// Otherwise update timesheet view
		timesheetModel, cmd := m.TimesheetView.Update(msg)
		m.TimesheetView = timesheetModel.(TimesheetModel)
		return m, cmd

	case FormMode:
		// Check for special message to return to timesheet mode
		if _, ok := msg.(ReturnToTimesheetMsg); ok {
			m.Mode = TimesheetMode
			// Refresh the timesheet data
			return m, m.TimesheetView.RefreshCmd()
		}

		// Otherwise update form view
		formModel, cmd := m.FormView.Update(msg)
		m.FormView = formModel.(FormModel)
		return m, cmd
	}

	return m, cmd
}

func (m AppModel) View() string {
	switch m.Mode {
	case TimesheetMode:
		return m.TimesheetView.View()
	case FormMode:
		return m.FormView.View()
	}
	return "Unknown mode"
}

// Message to return to timesheet mode
type ReturnToTimesheetMsg struct{}

func ReturnToTimesheet() tea.Cmd {
	return func() tea.Msg {
		return ReturnToTimesheetMsg{}
	}
}
