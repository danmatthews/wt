package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/danmatthews/wt/internal/model"
)

// Human-readable styling for the list views. lipgloss's default renderer
// detects the color profile of stdout, so styling is stripped automatically
// when stdout is not a terminal (piped/redirected) and when NO_COLOR is set
// (https://no-color.org). The --json path never reaches these renderers.
//
// Each role is a single combined style: lipgloss does not compose by nesting
// one Render inside another (it re-styles per rune and leaks literal escapes),
// so we style each segment once and concatenate the results.
var (
	styleProject = lipgloss.NewStyle().Bold(true).Underline(true)
	styleName    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	styleBadge   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleMuted   = lipgloss.NewStyle().Faint(true)
	styleDesc    = lipgloss.NewStyle().Italic(true)
	styleEPName  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleURL     = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("6"))
)

func renderProject(p *model.Project) {
	fmt.Println(styleProject.Render(p.ProjectPath))
	for _, w := range p.Worktrees {
		renderWorktree(w)
	}
}

func renderWorktree(w *model.Worktree) {
	// Heading: a bulleted, bold-cyan name with an optional "base" badge.
	head := "  " + styleName.Render("● "+w.Name)
	if w.Special {
		head += " " + styleBadge.Render("base")
	}
	fmt.Println(head)

	fmt.Printf("    %s\n", styleMuted.Render(w.Path))
	if w.Description != "" {
		fmt.Printf("    %s\n", styleDesc.Render(w.Description))
	}
	for _, ep := range w.EntryPoints {
		line := fmt.Sprintf("    %s %s  %s",
			styleMuted.Render("↳"), styleEPName.Render(ep.Name), styleURL.Render(ep.URL))
		if ep.Type != model.TypeURL {
			line += " " + styleMuted.Render("("+ep.Type+")")
		}
		if ep.Description != "" {
			line += "  " + styleMuted.Render("— "+ep.Description)
		}
		fmt.Println(line)
	}
}
