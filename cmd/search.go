package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/nemethk/ccsm/internal/session"
	"github.com/spf13/cobra"
)

var searchJSON bool

var searchCmd = &cobra.Command{
	Use:   "search <term>",
	Short: "Search sessions by first message or project path",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "output as JSON")
}

func runSearch(cmd *cobra.Command, args []string) error {
	term := strings.ToLower(args[0])

	sessions, err := session.Load()
	if err != nil {
		return err
	}

	type result struct {
		session.Session
		snippet string
		turn    int
	}

	var matched []result
	for _, s := range sessions {
		snippet, turn := session.MatchText(s.FilePath, term)
		if turn == 0 && !strings.Contains(strings.ToLower(s.ProjectPath), term) {
			continue
		}
		matched = append(matched, result{s, snippet, turn})
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].LastTime.After(matched[j].LastTime)
	})

	if searchJSON {
		type jsonResult struct {
			session.Session
			MatchSnippet string `json:"match"`
			MatchTurn    int    `json:"turn"`
		}
		out := make([]jsonResult, len(matched))
		for i, r := range matched {
			out[i] = jsonResult{r.Session, r.snippet, r.turn}
		}
		return json.NewEncoder(os.Stdout).Encode(out)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DATE\tTIME\tUUID\tPROJECT\tTURN\tMATCH")
	for _, r := range matched {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t#%d\t%s\n",
			r.FirstTime.Format("2006-01-02"),
			r.FirstTime.Format("15:04"),
			r.UUID,
			shortenPath(r.ProjectPath),
			r.turn,
			r.snippet,
		)
	}
	return w.Flush()
}
