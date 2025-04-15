package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// displayTableStats displays stats in a table format
func displayTableStats(stats *StatsCollector) {
	// ANSI color helpers
	green := "\033[32m"
	yellow := "\033[33m"
	reset := "\033[0m"
	colorize := func(s, color string) string {
		if noColor {
			return s
		}
		return color + s + reset
	}

	// Find max repo name length
	maxRepoLen := len("Repository")
	for _, repoStat := range stats.PerRepoStats {
		if l := len(repoStat.RepoName); l > maxRepoLen {
			maxRepoLen = l
		}
	}
	if maxRepoLen > 40 {
		maxRepoLen = 40 // hard cap for very long repo names
	}

	repoCol := maxRepoLen
	colWidths := []int{repoCol, 14, 20, 12}

	// Table border helpers
	top := "╭"
	sep := "├"
	bot := "╰"
	for i, w := range colWidths {
		top += pad("─", w+2) // +2 for padding spaces
		sep += pad("─", w+2)
		bot += pad("─", w+2)
		if i < len(colWidths)-1 {
			top += "┬"
			sep += "┼"
			bot += "┴"
		} else {
			top += "╮"
			sep += "┤"
			bot += "╯"
		}
	}

	headRepo := fmt.Sprintf("%-*s", repoCol, "Repository")
	headCombined := fmt.Sprintf("%*s", colWidths[1], "PRs Combined")
	headSkipped := fmt.Sprintf("%-*s", colWidths[2], "Skipped")
	headStatus := fmt.Sprintf("%-*s", colWidths[3], "Status")
	head := fmt.Sprintf(
		"│ %-*s │ %s │ %s │ %s │",
		repoCol, headRepo,
		headCombined,
		headSkipped,
		headStatus,
	)

	fmt.Println(top)
	fmt.Println(head)
	fmt.Println(sep)

	for _, repoStat := range stats.PerRepoStats {
		status := "OK"
		statusColor := green
		if repoStat.TotalPRs == 0 {
			status = "NO OPEN PRs"
			statusColor = green
		} else if repoStat.NotEnoughPRs {
			status = "NOT ENOUGH"
			statusColor = yellow
		}

		mcColor := green
		dnmColor := green
		if repoStat.SkippedMergeConf > 0 {
			mcColor = yellow
		}
		if repoStat.SkippedCriteria > 0 {
			dnmColor = yellow
			}
		
		// Construct skipped text without color codes first to get proper width
		mcRaw := fmt.Sprintf("%d", repoStat.SkippedMergeConf)
		dnmRaw := fmt.Sprintf("%d", repoStat.SkippedCriteria)
		skippedPlain := fmt.Sprintf("%s (MC), %s (DNM)", mcRaw, dnmRaw)
		
		// Then add color codes for display
		skippedDisplay := ""
		if noColor {
			skippedDisplay = skippedPlain
		} else {
			mcDisplay := mcColor + mcRaw + reset
			dnmDisplay := dnmColor + dnmRaw + reset
			skippedDisplay = fmt.Sprintf("%s (MC), %s (DNM)", mcDisplay, dnmDisplay)
		}
		
		// Ensure proper padding based on the plaintext width
		paddingLen := colWidths[2] - len(skippedPlain)
		padding := ""
		if paddingLen > 0 {
			padding = strings.Repeat(" ", paddingLen)
		}
		
		statusColored := colorize(status, statusColor)
		statusPadding := colWidths[3] - len(status)
		if statusPadding > 0 {
			statusColored += strings.Repeat(" ", statusPadding)
		}

		fmt.Printf(
			"│ %-*s │ %*d │ %s%s │ %s │\n",
			repoCol, repoStat.RepoName,
			colWidths[1], repoStat.CombinedCount,
			skippedDisplay, padding,
			statusColored,
		)
	}
	fmt.Println(bot)

	// Print summary mini-table with proper padding
	summaryTop := "╭───────────────┬───────────────┬───────────────────────┬───────────────╮"
	summaryHead := "│ Repos         │ Combined PRs  │ Skipped               │ Total PRs     │"
	summarySep := "├───────────────┼───────────────┼───────────────────────┼───────────────┤"
	
	// Use the same approach for summary table to ensure consistency
	mcSummaryRaw := fmt.Sprintf("%d", stats.PRsSkippedMergeConflict)
	dnmSummaryRaw := fmt.Sprintf("%d", stats.PRsSkippedCriteria)
	
	skippedSummaryPlain := fmt.Sprintf("%s (MC), %s (DNM)", mcSummaryRaw, dnmSummaryRaw)
	skippedSummaryDisplay := skippedSummaryPlain
	if !noColor {
		mcColor := green
		dnmColor := green
		if stats.PRsSkippedMergeConflict > 0 {
			mcColor = yellow
		}
		if stats.PRsSkippedCriteria > 0 {
			dnmColor = yellow
		}
		mcDisplay := mcColor + mcSummaryRaw + reset
		dnmDisplay := dnmColor + dnmSummaryRaw + reset
		skippedSummaryDisplay = fmt.Sprintf("%s (MC), %s (DNM)", mcDisplay, dnmDisplay)
	}
	
	summarySkippedPadding := 21 - len(skippedSummaryPlain)
	summaryPadding := ""
	if summarySkippedPadding > 0 {
		summaryPadding = strings.Repeat(" ", summarySkippedPadding)
	}
	
	summaryRow := fmt.Sprintf(
		"│ %-13d │ %-13d │ %s%s │ %-13d │",
		stats.ReposProcessed,
		stats.PRsCombined,
		skippedSummaryDisplay,
		summaryPadding,
		len(stats.CombinedPRLinks),
	)
	summaryBot := "╰───────────────┴───────────────┴───────────────────────┴───────────────╯"
	fmt.Println()
	fmt.Println(summaryTop)
	fmt.Println(summaryHead)
	fmt.Println(summarySep)
	fmt.Println(summaryRow)
	fmt.Println(summaryBot)

	// Print PR links block (blue color)
	if len(stats.CombinedPRLinks) > 0 {
		blue := "\033[34m"
		fmt.Println("\nLinks to Combined PRs:")
		for _, link := range stats.CombinedPRLinks {
			if noColor {
				fmt.Println("-", link)
			} else {
				fmt.Printf("- %s%s%s\n", blue, link, reset)
			}
		}
	}
	fmt.Println()
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
	fmt.Printf("Repositories Processed: %d\n", stats.ReposProcessed)
	fmt.Printf("PRs Combined: %d\n", stats.PRsCombined)
	fmt.Printf("PRs Skipped (Merge Conflicts): %d\n", stats.PRsSkippedMergeConflict)
	fmt.Printf("PRs Skipped (Did Not Match): %d\n", stats.PRsSkippedCriteria)
	fmt.Printf("Execution Time: %s\n", elapsed.Round(time.Second))

	fmt.Println("Links to Combined PRs:")
	for _, link := range stats.CombinedPRLinks {
		fmt.Println("-", link)
	}

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

// pad returns a string of n runes of s (usually "─")
func pad(s string, n int) string {
	if n <= 0 {
		return ""
	}
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
