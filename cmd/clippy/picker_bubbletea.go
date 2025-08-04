package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neilberkman/clippy/pkg/recent"
	"github.com/neilberkman/mimedescription"
)

// pickerModel represents the state of our file picker
type pickerModel struct {
	files         []recent.FileInfo
	cursor        int
	selected      map[int]bool
	done          bool
	cancelled     bool
	pasteMode     bool // true if user pressed 'p' to copy & paste
	absoluteTime  bool
	terminalWidth int
}

// pickerItem represents a file item with its display state
type pickerItem struct {
	file     recent.FileInfo
	index    int
	selected bool
	focused  bool
}

// Initialize the model
func (m pickerModel) Init() tea.Cmd {
	// Request initial window size
	return tea.WindowSize()
}

// Update handles messages
func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			m.done = true
			return m, tea.Quit

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}
		}

		// Also handle string-based keys
		switch msg.String() {
		case "q":
			m.cancelled = true
			m.done = true
			return m, tea.Quit

		case "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}

		case " ", "space":
			// Toggle selection
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = true
			}

		case "enter":
			m.done = true
			return m, tea.Quit

		case "p":
			// Copy & paste mode
			m.pasteMode = true
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the picker
func (m pickerModel) View() string {
	if m.done {
		return ""
	}

	var builder strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	builder.WriteString(headerStyle.Render("Select files (Enter: current item, Space: multi-select, p: copy & paste)"))
	builder.WriteString("\n\n")

	// File list
	for i, file := range m.files {
		item := pickerItem{
			file:     file,
			index:    i,
			selected: m.selected[i],
			focused:  i == m.cursor,
		}
		builder.WriteString(m.renderItem(item))
		builder.WriteString("\n")
	}

	// Footer with file details
	if m.cursor < len(m.files) {
		builder.WriteString("\n")
		builder.WriteString(m.renderDetails(m.files[m.cursor]))
	}

	// Help text
	helpStyle := lipgloss.NewStyle().Faint(true)
	builder.WriteString("\n")
	builder.WriteString(helpStyle.Render("↑/↓ navigate • Enter: copy current • Space: toggle select • p: copy&paste • Esc: cancel"))

	return builder.String()
}

// renderItem renders a single file item
func (m pickerModel) renderItem(item pickerItem) string {
	// Styles
	normalStyle := lipgloss.NewStyle()
	focusedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	checkboxStyle := lipgloss.NewStyle().Width(3)
	ageStyle := lipgloss.NewStyle().Faint(true)
	extStyle := lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("243"))

	// Checkbox
	checkbox := "[ ]"
	if item.selected {
		checkbox = "[✓]"
	}

	// Format age
	var ageStr string
	if m.absoluteTime {
		ageStr = item.file.Modified.Format("Jan 2 15:04")
	} else {
		age := item.file.Age()
		if age < time.Minute {
			ageStr = fmt.Sprintf("%ds ago", int(age.Seconds()))
		} else if age < time.Hour {
			ageStr = fmt.Sprintf("%dm ago", int(age.Minutes()))
		} else if age < 24*time.Hour {
			ageStr = fmt.Sprintf("%dh ago", int(age.Hours()))
		} else {
			ageStr = fmt.Sprintf("%dd ago", int(age.Hours()/24))
		}
	}

	// Get file type display
	fileType := getFileTypeDisplay(item.file.MimeType)

	// Calculate available width for filename
	// Account for: checkbox(3) + spaces(2) + age(~10) + file type + padding
	availableWidth := 50 // default
	if m.terminalWidth > 0 {
		// Leave room for: "▶ " or "  " (2), checkbox (3), spaces (3), age (~10), file type, and some padding
		availableWidth = m.terminalWidth - 25 - len(ageStr) - len(fileType)
		if availableWidth < 20 {
			availableWidth = 20
		}
	}

	// Truncate filename using middle truncation
	displayName := truncateMiddle(item.file.Name, availableWidth)

	// Build the line
	line := fmt.Sprintf("%s %s [%s] (%s)",
		checkboxStyle.Render(checkbox),
		displayName,
		extStyle.Render(fileType),
		ageStyle.Render(ageStr),
	)

	// Apply styles
	if item.focused {
		if item.selected {
			return selectedStyle.Render("▶ ") + focusedStyle.Render(line[2:])
		}
		return focusedStyle.Render("▶ " + line[2:])
	}

	if item.selected {
		return selectedStyle.Render("  " + line[2:])
	}

	return normalStyle.Render("  " + line[2:])
}

// renderDetails renders file details for the currently focused item
func (m pickerModel) renderDetails(file recent.FileInfo) string {
	detailStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().Faint(true)
	valueStyle := lipgloss.NewStyle()

	// Format size
	var sizeStr string
	if file.Size < 1024 {
		sizeStr = fmt.Sprintf("%d B", file.Size)
	} else if file.Size < 1024*1024 {
		sizeStr = fmt.Sprintf("%.1f KB", float64(file.Size)/1024)
	} else if file.Size < 1024*1024*1024 {
		sizeStr = fmt.Sprintf("%.1f MB", float64(file.Size)/(1024*1024))
	} else {
		sizeStr = fmt.Sprintf("%.1f GB", float64(file.Size)/(1024*1024*1024))
	}

	details := fmt.Sprintf(
		"%s %s\n%s %s\n%s %s\n%s %s\n%s %s",
		labelStyle.Render("Name:"),
		valueStyle.Render(file.Name),
		labelStyle.Render("Type:"),
		valueStyle.Render(getFileTypeDisplay(file.MimeType)),
		labelStyle.Render("Size:"),
		valueStyle.Render(sizeStr),
		labelStyle.Render("Modified:"),
		valueStyle.Render(file.Modified.Format("Jan 2 15:04:05")),
		labelStyle.Render("Path:"),
		valueStyle.Render(truncateString(file.Path, 60)),
	)

	return detailStyle.Render(details)
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// truncateMiddle truncates a string in the middle, preserving start and end
func truncateMiddle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 5 {
		return truncateString(s, maxLen)
	}

	// Calculate how much to show from each end
	startLen := (maxLen - 3) / 2
	endLen := maxLen - 3 - startLen

	return s[:startLen] + "..." + s[len(s)-endLen:]
}

// getFileTypeDisplay returns a human-readable file type based on MIME type
func getFileTypeDisplay(mimeType string) string {
	if mimeType == "" {
		return ""
	}

	// Use the mimedescription library
	if desc, found := mimedescription.Get(mimeType); found {
		return desc
	}

	// Fallback to a simple extraction from MIME type
	parts := strings.Split(mimeType, "/")
	if len(parts) > 1 && parts[1] != "octet-stream" {
		// Just return the main type capitalized
		mainType := parts[0]
		if len(mainType) > 0 {
			return strings.ToUpper(mainType[:1]) + mainType[1:]
		}
	}
	return "File"
}

// showBubbleTeaPickerWithResult shows an interactive picker and returns the full result
func showBubbleTeaPickerWithResult(files []recent.FileInfo, absoluteTime bool) (*recent.PickerResult, error) {
	m := pickerModel{
		files:        files,
		cursor:       0,
		selected:     make(map[int]bool),
		absoluteTime: absoluteTime,
	}

	// Run the program inline (not fullscreen)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	// Get the final model
	finalPicker := finalModel.(pickerModel)

	// Check if cancelled
	if finalPicker.cancelled {
		return nil, fmt.Errorf("cancelled")
	}

	// Collect selected files
	var selectedFiles []*recent.FileInfo

	// If nothing is selected, use the current item
	if len(finalPicker.selected) == 0 && finalPicker.cursor < len(files) {
		fileCopy := files[finalPicker.cursor]
		selectedFiles = append(selectedFiles, &fileCopy)
	} else {
		// Return all selected files
		for i := range files {
			if finalPicker.selected[i] {
				fileCopy := files[i]
				selectedFiles = append(selectedFiles, &fileCopy)
			}
		}
	}

	return &recent.PickerResult{
		Files:     selectedFiles,
		PasteMode: finalPicker.pasteMode,
	}, nil
}
