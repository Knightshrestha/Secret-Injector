package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/Knightshrestha/Secret-Injector/core"
	"github.com/Knightshrestha/Secret-Injector/database/generated"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize and select projects",
	Long:  `Fetch available projects and allow user to select one or multiple projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Fetch projects
		projects, err := core.FetchProjects()
		if err != nil {
			log.Fatalf("Something went wrong, fetching projects: %s", err)
		}

		if len(projects) == 0 {
			fmt.Println("No projects available")
			return
		}

		// Run interactive selector
		selectedProjects := selectProjects(projects)

		// Display results
		if len(selectedProjects) == 0 {
			fmt.Println("No projects selected")
			return
		}

		fmt.Println("\n✓ Selected projects:")
		for _, project := range selectedProjects {
			fmt.Printf("  • %s (ID: %s)\n", project.Name, project.ID)
		}

		var projectIDs []string
		for _, project := range selectedProjects {
			projectIDs = append(projectIDs, project.ID)
		}

		allSecrets, err := core.FetchSecrets(projectIDs)
		if err != nil {
			log.Fatalf("Something went wrong, fetching secrets: %s", err)
		}
		
		if len(allSecrets) == 0 {
			fmt.Println("No secrets to display")
			return
		}

		fmt.Println("\n✓ Selected Secrets:")
		for _, secret := range allSecrets {
			fmt.Printf("  • %s: %s (ID: %s)\n", secret.Key, secret.Value, secret.ID)
		}

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// selectProjects runs the interactive multi-select
func selectProjects(projects []generated.ProjectList) []generated.ProjectList {
	m := model{
		projects: projects,
		selected: make(map[int]bool),
		cursor:   0,
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return []generated.ProjectList{}
	}

	finalModel := result.(model)
	var selected []generated.ProjectList
	for idx := range finalModel.selected {
		selected = append(selected, projects[idx])
	}

	return selected
}

// Bubbletea Model
type model struct {
	projects []generated.ProjectList
	selected map[int]bool
	cursor   int
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}

		case " ":
			// Toggle selection
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = true
			}

		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString("Select projects (space to toggle, enter to confirm):\n\n")

	for i, project := range m.projects {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if m.selected[i] {
			checked = "✓"
		}

		b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, checked, project.Name))
	}

	b.WriteString("\nControls: ↑/↓ navigate • space select • enter confirm • q quit\n")

	return b.String()
}
