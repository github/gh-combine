package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ANSI color codes
const (
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
)

// Table styling characters
const (
	tableTopLeft     = "╭"
	tableTopRight    = "╮"
	tableBottomLeft  = "╰"
	tableBottomRight = "╯"
	tableHorizLine   = "─"
	tableVertLine    = "│"
	tableTopJoin     = "┬"
	tableMidJoin     = "┼"
	tableBottomJoin  = "┴"
	tableMidLeft     = "├"
	tableMidRight    = "┤"
)

// Labels used in output
const (
	labelMC  = "MC"  // Merge Conflict
	labelDNM = "DNM" // Did Not Match criteria

	maxRepoNameLength = 40 // Hard cap for very long repo names
)

// displayTableStats displays stats in a table format
func displayTableStats(stats *StatsCollector) {
	// Calculate column widths
	colWidths := calculateColumnWidths(stats)

	// Generate the table header
	top, sep, bot := generateTableBorders(colWidths)
	head := generateTableHeader(colWidths)

	// Print table header
	fmt.Println(top)
	fmt.Println(head)
	fmt.Println(sep)

	// Print each repo row
	for _, repoStat := range stats.PerRepoStats {
		fmt.Println(formatRepoRow(repoStat, colWidths))
	}
	fmt.Println(bot)

	// Print summary table
	displaySummaryTable(stats)

	// Print PR links
	displayPRLinks(stats.CombinedPRLinks)

	fmt.Println()
}

// calculateColumnWidths determines the appropriate width for each column
func calculateColumnWidths(stats *StatsCollector) []int {
	// Find max repo name length
	maxRepoLen := len("Repository")
	for _, repoStat := range stats.PerRepoStats {
		if l := len(repoStat.RepoName); l > maxRepoLen {
			maxRepoLen = l
		}
	}
	if maxRepoLen > maxRepoNameLength {
		maxRepoLen = maxRepoNameLength
	}

	return []int{maxRepoLen, 14, 20, 12}
}

// generateTableBorders creates the top, separator, and bottom borders of the table
func generateTableBorders(colWidths []int) (top, sep, bot string) {
	top = tableTopLeft
	sep = tableMidLeft
	bot = tableBottomLeft

	for i, w := range colWidths {
		paddedLine := strings.Repeat(tableHorizLine, w+2) // +2 for padding spaces
		top += paddedLine
		sep += paddedLine
		bot += paddedLine

		if i < len(colWidths)-1 {
			top += tableTopJoin
			sep += tableMidJoin
			bot += tableBottomJoin
		} else {
			top += tableTopRight
			sep += tableMidRight
			bot += tableBottomRight
		}
	}

	return top, sep, bot
}

// generateTableHeader creates the header row for the table
func generateTableHeader(colWidths []int) string {
	headRepo := fmt.Sprintf("%-*s", colWidths[0], "Repository")
	headCombined := fmt.Sprintf("%*s", colWidths[1], "PRs Combined")
	headSkipped := fmt.Sprintf("%-*s", colWidths[2], "Skipped")
	headStatus := fmt.Sprintf("%-*s", colWidths[3], "Status")

	return fmt.Sprintf(
		"%s %-*s %s %s %s %s %s %s %s",
		tableVertLine,
		colWidths[0], headRepo,
		tableVertLine,
		headCombined,
		tableVertLine,
		headSkipped,
		tableVertLine,
		headStatus,
		tableVertLine,
	)
}

// formatRepoRow formats a single repository row for the table
func formatRepoRow(repoStat *RepoStats, colWidths []int) string {
	// Format status text
	status, statusColor := getStatusInfo(repoStat)
	statusColored := colorize(status, statusColor)
	statusPadding := colWidths[3] - len(status)
	if statusPadding > 0 {
		statusColored += strings.Repeat(" ", statusPadding)
	}

	// Format skipped info
	mcRaw := fmt.Sprintf("%d", repoStat.SkippedMergeConf)
	dnmRaw := fmt.Sprintf("%d", repoStat.SkippedCriteria)

	// Get color for skipped metrics
	mcColor := getColorForValue(repoStat.SkippedMergeConf > 0)
	dnmColor := getColorForValue(repoStat.SkippedCriteria > 0)

	// Format with appropriate coloring
	skippedText, padding := formatSkippedText(
		mcRaw, dnmRaw,
		mcColor, dnmColor,
		colWidths[2],
	)

	return fmt.Sprintf(
		"%s %-*s %s %*d %s %s%s %s %s %s",
		tableVertLine,
		colWidths[0], repoStat.RepoName,
		tableVertLine,
		colWidths[1], repoStat.CombinedCount,
		tableVertLine,
		skippedText, padding,
		tableVertLine,
		statusColored,
		tableVertLine,
	)
}

// formatSkippedText formats the "skipped" cell with proper colors and padding
func formatSkippedText(mcRaw, dnmRaw string, mcColor, dnmColor string, colWidth int) (text, padding string) {
	skippedPlain := fmt.Sprintf("%s (%s), %s (%s)", mcRaw, labelMC, dnmRaw, labelDNM)

	var skippedDisplay string
	if noColor {
		skippedDisplay = skippedPlain
	} else {
		mcDisplay := mcColor + mcRaw + colorReset
		dnmDisplay := dnmColor + dnmRaw + colorReset
		skippedDisplay = fmt.Sprintf("%s (%s), %s (%s)", mcDisplay, labelMC, dnmDisplay, labelDNM)
	}

	// Calculate padding needed
	paddingLen := colWidth - len(skippedPlain)
	padding = ""
	if paddingLen > 0 {
		padding = strings.Repeat(" ", paddingLen)
	}

	return skippedDisplay, padding
}

// getStatusInfo returns the appropriate status text and color based on the repo stats
func getStatusInfo(repoStat *RepoStats) (string, string) {
	status := "OK"
	statusColor := colorGreen

	if repoStat.TotalPRs == 0 {
		status = "NO OPEN PRs"
	} else if repoStat.NotEnoughPRs {
		status = "NOT ENOUGH"
		statusColor = colorYellow
	}

	return status, statusColor
}

// getColorForValue returns the appropriate color based on a condition
func getColorForValue(isWarning bool) string {
	if isWarning {
		return colorYellow
	}
	return colorGreen
}

// displaySummaryTable prints a summary table with overall statistics
func displaySummaryTable(stats *StatsCollector) {
	// Table borders (predefined for simplicity)
	summaryTop := "╭───────────────┬───────────────┬───────────────────────┬───────────────╮"
	summaryHead := "│ Repos         │ Combined PRs  │ Skipped               │ Total PRs     │"
	summarySep := "├───────────────┼───────────────┼───────────────────────┼───────────────┤"
	summaryBot := "╰───────────────┴───────────────┴───────────────────────┴───────────────╯"

	// Format skipped metrics with colors
	mcSummaryRaw := fmt.Sprintf("%d", stats.PRsSkippedMergeConflict)
	dnmSummaryRaw := fmt.Sprintf("%d", stats.PRsSkippedCriteria)

	mcColor := getColorForValue(stats.PRsSkippedMergeConflict > 0)
	dnmColor := getColorForValue(stats.PRsSkippedCriteria > 0)

	skippedSummaryText, summaryPadding := formatSkippedText(
		mcSummaryRaw, dnmSummaryRaw,
		mcColor, dnmColor,
		21, // Fixed width for the skipped column
	)

	summaryRowPRCount := interface{}(len(stats.CombinedPRLinks))

	if summaryRowPRCount == 0 && dryRun {
		summaryRowPRCount = "DRY RUN"
	}

	// Generate the summary row
	summaryRow := fmt.Sprintf(
		"│ %-13d │ %-13d │ %s%s │ %-13v │",
		stats.ReposProcessed,
		stats.PRsCombined,
		skippedSummaryText,
		summaryPadding,
		summaryRowPRCount,
	)

	// Print the summary table
	fmt.Println()
	fmt.Println(summaryTop)
	fmt.Println(summaryHead)
	fmt.Println(summarySep)
	fmt.Println(summaryRow)
	fmt.Println(summaryBot)
}

// displayPRLinks prints the links to combined PRs
func displayPRLinks(links []string) {
	if len(links) == 0 {
		return
	}

	fmt.Println("\nLinks to Combined PRs:")
	for _, link := range links {
		if noColor {
			fmt.Println("-", link)
		} else {
			fmt.Printf("- %s%s%s\n", colorBlue, link, colorReset)
		}
	}
}

// displayJSONStats displays stats in JSON format
func displayJSONStats(stats *StatsCollector) {
	output := map[string]interface{}{
		"reposProcessed":          stats.ReposProcessed,
		"prsCombined":             stats.PRsCombined,
		"prsSkippedMergeConflict": stats.PRsSkippedMergeConflict,
		"prsSkippedCriteria":      stats.PRsSkippedCriteria,
		"executionTime":           stats.EndTime.Sub(stats.StartTime).String(),
		"combinedPRLinks":         stats.CombinedPRLinks,
		"perRepoStats":            stats.PerRepoStats,
	}
	jsonData, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(jsonData))
}

// displayPlainStats displays stats in plain text format
func displayPlainStats(stats *StatsCollector) {
	elapsed := stats.EndTime.Sub(stats.StartTime)

	// Print summary statistics
	fmt.Printf("Repositories Processed: %d\n", stats.ReposProcessed)
	fmt.Printf("PRs Combined: %d\n", stats.PRsCombined)
	fmt.Printf("PRs Skipped (Merge Conflicts): %d\n", stats.PRsSkippedMergeConflict)
	fmt.Printf("PRs Skipped (Did Not Match): %d\n", stats.PRsSkippedCriteria)
	fmt.Printf("Execution Time: %s\n", elapsed.Round(time.Second))

	// Print PR links
	fmt.Println("\nLinks to Combined PRs:")
	for _, link := range stats.CombinedPRLinks {
		fmt.Println("-", link)
	}

	// Print per-repository details
	fmt.Println("\nPer-Repository Details:")
	for _, repoStat := range stats.PerRepoStats {
		fmt.Printf("  %s\n", repoStat.RepoName)
		if repoStat.NotEnoughPRs {
			fmt.Println("    Not enough PRs to combine.")
			continue
		}
		fmt.Printf("    Combined: %d\n", repoStat.CombinedCount)
		fmt.Printf("    Skipped (Merge Conflicts): %d\n", repoStat.SkippedMergeConf)
		fmt.Printf("    Skipped (Did Not Match): %d\n", repoStat.SkippedCriteria)
		if repoStat.CombinedPRLink != "" {
			fmt.Printf("    Combined PR: %s\n", repoStat.CombinedPRLink)
		}
	}
}

// colorize adds color to text if colors are enabled
func colorize(s, color string) string {
	if noColor {
		return s
	}
	return color + s + colorReset
}
