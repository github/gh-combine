package cmd

import (
	"testing"
	"time"
)

func TestDisplayTableStats(t *testing.T) {
	stats := &StatsCollector{
		ReposProcessed:          2,
		PRsCombined:             5,
		PRsSkippedMergeConflict: 1,
		PRsSkippedCriteria:      2,
		CombinedPRLinks:         []string{"http://example.com/pr1", "http://example.com/pr2"},
		PerRepoStats: map[string]*RepoStats{
			"repo1": {
				RepoName:         "repo1",
				CombinedCount:    3,
				SkippedMergeConf: 1,
				SkippedCriteria:  0,
				CombinedPRLink:   "http://example.com/pr1",
				NotEnoughPRs:     false,
				TotalPRs:         5,
			},
			"repo2": {
				RepoName:         "repo2",
				CombinedCount:    2,
				SkippedMergeConf: 0,
				SkippedCriteria:  2,
				CombinedPRLink:   "http://example.com/pr2",
				NotEnoughPRs:     false,
				TotalPRs:         4,
			},
		},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(2 * time.Minute),
	}

	displayTableStats(stats)
	// Add assertions or manual verification as needed
}

func TestDisplayJSONStats(t *testing.T) {
	stats := &StatsCollector{
		ReposProcessed:          2,
		PRsCombined:             5,
		PRsSkippedMergeConflict: 1,
		PRsSkippedCriteria:      2,
		CombinedPRLinks:         []string{"http://example.com/pr1", "http://example.com/pr2"},
		PerRepoStats: map[string]*RepoStats{
			"repo1": {
				RepoName:         "repo1",
				CombinedCount:    3,
				SkippedMergeConf: 1,
				SkippedCriteria:  0,
				CombinedPRLink:   "http://example.com/pr1",
				NotEnoughPRs:     false,
				TotalPRs:         5,
			},
			"repo2": {
				RepoName:         "repo2",
				CombinedCount:    2,
				SkippedMergeConf: 0,
				SkippedCriteria:  2,
				CombinedPRLink:   "http://example.com/pr2",
				NotEnoughPRs:     false,
				TotalPRs:         4,
			},
		},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(2 * time.Minute),
	}

	displayJSONStats(stats)
	// Add assertions or manual verification as needed
}

func TestDisplayPlainStats(t *testing.T) {
	stats := &StatsCollector{
		ReposProcessed:          2,
		PRsCombined:             5,
		PRsSkippedMergeConflict: 1,
		PRsSkippedCriteria:      2,
		CombinedPRLinks:         []string{"http://example.com/pr1", "http://example.com/pr2"},
		PerRepoStats: map[string]*RepoStats{
			"repo1": {
				RepoName:         "repo1",
				CombinedCount:    3,
				SkippedMergeConf: 1,
				SkippedCriteria:  0,
				CombinedPRLink:   "http://example.com/pr1",
				NotEnoughPRs:     false,
				TotalPRs:         5,
			},
			"repo2": {
				RepoName:         "repo2",
				CombinedCount:    2,
				SkippedMergeConf: 0,
				SkippedCriteria:  2,
				CombinedPRLink:   "http://example.com/pr2",
				NotEnoughPRs:     false,
				TotalPRs:         4,
			},
		},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(2 * time.Minute),
	}

	displayPlainStats(stats)
	// Add assertions or manual verification as needed
}