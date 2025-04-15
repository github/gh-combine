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
		mcRaw := fmt.Sprintf("%d", repoStat.SkippedMergeConf)
		dnmRaw := fmt.Sprintf("%d", repoStat.SkippedCriteria)
		skippedRaw := fmt.Sprintf("%s (MC), %s (DNM)", mcRaw, dnmRaw)
		skippedPadded := fmt.Sprintf("%-*s", colWidths[2], skippedRaw)
		mcIdx := strings.Index(skippedPadded, mcRaw)
		dnmIdx := strings.Index(skippedPadded, dnmRaw)
		skippedColored := skippedPadded
		if mcIdx != -1 {
			skippedColored = skippedColored[:mcIdx] + colorize(mcRaw, mcColor) + skippedColored[mcIdx+len(mcRaw):]
		}
		if dnmIdx != -1 {
			dnmIdx = strings.Index(skippedColored, dnmRaw)
			skippedColored = skippedColored[:dnmIdx] + colorize(dnmRaw, dnmColor) + skippedColored[dnmIdx+len(dnmRaw):]
		}
		statusColored := colorize(status, statusColor)
		statusColored = fmt.Sprintf("%-*s", colWidths[3]+len(statusColored)-len(status), statusColored)

		fmt.Printf(
			"│ %-*s │ %s │ %s │ %s │\n",
			repoCol, repoStat.RepoName,
			fmt.Sprintf("%*d", colWidths[1], repoStat.CombinedCount),
			skippedColored,
			statusColored,
		)
	}
	fmt.Println(bot)

	// Print summary mini-table with proper padding
	summaryTop := "╭───────────────┬───────────────┬───────────────────────┬───────────────╮"
	summaryHead := "│ Repos         │ Combined PRs  │ Skipped               │ Total PRs     │"
	summarySep := "├───────────────┼───────────────┼───────────────────────┼───────────────┤"
	skippedRaw := fmt.Sprintf("%d (MC), %d (DNM)", stats.PRsSkippedMergeConflict, stats.PRsSkippedCriteria)
	summaryRow := fmt.Sprintf(
		"│ %-13d │ %-13d │ %-21s │ %-13d │",
		stats.ReposProcessed,
		stats.PRsCombined,
		skippedRaw,
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

// truncate shortens a string to maxLen runes, adding … if truncated
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-1]) + "…"
}