package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/nemethk/ccsm/internal/cache"
	"github.com/nemethk/ccsm/internal/session"
	"github.com/spf13/cobra"
)

var (
	listJSON     bool
	listProject  string
	listSince    string
	listAI       bool
	listMinTurns int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Claude Code sessions",
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "output as JSON")
	listCmd.Flags().StringVar(&listProject, "project", "", "filter by project path substring")
	listCmd.Flags().StringVar(&listSince, "since", "", "filter sessions on or after date (YYYY-MM-DD)")
	listCmd.Flags().BoolVar(&listAI, "ai", false, "generate AI summaries via claude CLI (cached for future runs)")
	listCmd.Flags().IntVar(&listMinTurns, "min-turns", 0, "hide sessions with fewer than N user messages")
}

func runList(cmd *cobra.Command, args []string) error {
	sessions, err := session.Load()
	if err != nil {
		return err
	}

	sessions = applyFilters(sessions, listProject, listSince, listMinTurns)

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastTime.After(sessions[j].LastTime)
	})

	if listAI {
		var pending []session.Session
		for _, s := range sessions {
			if _, ok := cache.Get(s.UUID); !ok {
				pending = append(pending, s)
			}
		}
		if len(pending) > 0 {
			fmt.Fprintf(os.Stderr, "Generating AI summaries for %d session(s)...\n", len(pending))
			generated, skipped := 0, 0
			for i, s := range pending {
				fmt.Fprintf(os.Stderr, "  [%d/%d] %s", i+1, len(pending), s.UUID[:8])
				switch ensureSummary(s) {
				case summarizeGenerated:
					fmt.Fprintln(os.Stderr, " ✓")
					generated++
				case summarizeNoContent:
					fmt.Fprintln(os.Stderr, " (no content)")
					skipped++
				case summarizeError:
					fmt.Fprintln(os.Stderr, " (error)")
					skipped++
				}
			}
			fmt.Fprintf(os.Stderr, "Done: %d generated, %d skipped\n", generated, skipped)
		}
	}

	if listJSON {
		return json.NewEncoder(os.Stdout).Encode(sessions)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DATE\tTIME\tUUID\tPROJECT\tSUMMARY")
	for _, s := range sessions {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			s.FirstTime.Format("2006-01-02"),
			s.FirstTime.Format("15:04"),
			s.UUID,
			shortenPath(s.ProjectPath),
			sessionSummary(s),
		)
	}
	return w.Flush()
}

func applyFilters(sessions []session.Session, project, since string, minTurns int) []session.Session {
	var sinceTime time.Time
	if since != "" {
		sinceTime, _ = time.Parse("2006-01-02", since)
	}

	out := sessions[:0]
	for _, s := range sessions {
		if project != "" && !strings.Contains(s.ProjectPath, project) {
			continue
		}
		if !sinceTime.IsZero() && s.FirstTime.Before(sinceTime) {
			continue
		}
		if minTurns > 0 && s.MsgCount < minTurns {
			continue
		}
		out = append(out, s)
	}
	return out
}
