package cmd

import (
	"fmt"
	"strings"

	"github.com/nemethk/ccsm/internal/ai"
	"github.com/nemethk/ccsm/internal/session"
	"github.com/spf13/cobra"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize <uuid-prefix>",
	Short: "Show a detailed AI summary of a session",
	Args:  cobra.ExactArgs(1),
	RunE:  runSummarize,
}

func runSummarize(cmd *cobra.Command, args []string) error {
	s, err := session.FindByPrefix(args[0])
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("no session found with prefix %q", args[0])
	}

	msgs, err := session.Messages(s.UUID, 20)
	if err != nil {
		return err
	}
	var clean []string
	for _, m := range msgs {
		if !strings.HasPrefix(strings.TrimSpace(m), "<") {
			clean = append(clean, m)
		}
	}
	if len(clean) == 0 {
		return fmt.Errorf("no content to summarize in session %s", s.UUID[:8])
	}

	fmt.Printf("Session:  %s\n", s.UUID)
	fmt.Printf("Project:  %s\n", shortenPath(s.ProjectPath))
	fmt.Printf("Date:     %s\n", s.FirstTime.Format("2006-01-02 15:04"))
	fmt.Printf("Turns:    %d\n\n", s.MsgCount)

	detail, err := ai.Detail(clean)
	if err != nil {
		return err
	}
	fmt.Println(detail)
	return nil
}
