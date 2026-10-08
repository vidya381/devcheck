package reporter

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/vidya381/devcheck/internal/check"
)

var (
	passStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e"))
	warnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b"))
	failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444"))
	skipStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
)

// Render prints results to stdout. Skipped checks are only printed when
// showSkipped is set (the --verbose flag).
func Render(results []check.Result, showFix, showSkipped bool) {
	passed, warned, failed := 0, 0, 0

	for _, r := range results {
		switch r.Status {
		case check.StatusPass:
			fmt.Printf("%s  %s\n", passStyle.Render("✅"), r.Message)
			passed++
		case check.StatusWarn:
			fmt.Printf("%s  %s\n", warnStyle.Render("⚠️ "), r.Message)
			warned++
		case check.StatusFail:
			fmt.Printf("%s  %s\n", failStyle.Render("❌"), r.Message)
			if showFix && r.Fix != "" {
				fmt.Printf("   → %s\n", r.Fix)
			}
			failed++
		case check.StatusSkipped:
			if showSkipped {
				fmt.Printf("%s  %s\n", skipStyle.Render("–"), r.Message)
			}
		}
	}

	fmt.Println("────────────────────────────────────────")
	fmt.Println(summaryLine(passed, warned, failed))
	if failed > 0 && !showFix {
		fmt.Println("  Run `devcheck --fix` for suggestions")
	}
}

func summaryLine(passed, warned, failed int) string {
	return fmt.Sprintf("  %d passed  %d %s  %d failed", passed, warned, plural(warned, "warning"), failed)
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
