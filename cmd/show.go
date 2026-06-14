package cmd

import (
	"fmt"

	"github.com/nemethk/claude-code-session-manager/internal/session"
	"github.com/spf13/cobra"
)

var showTurns int

var showCmd = &cobra.Command{
	Use:   "show <uuid>",
	Short: "Show the first N user turns from a session",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	showCmd.Flags().IntVar(&showTurns, "turns", 5, "number of user turns to show")
}

func runShow(cmd *cobra.Command, args []string) error {
	msgs, err := session.Messages(args[0], showTurns)
	if err != nil {
		return err
	}
	if len(msgs) == 0 {
		return fmt.Errorf("session not found: %s", args[0])
	}
	for i, msg := range msgs {
		fmt.Printf("=== Turn %d ===\n%s\n\n", i+1, msg)
	}
	return nil
}
