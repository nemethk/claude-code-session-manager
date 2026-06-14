package main

import "github.com/nemethk/claude-code-session-manager/cmd"

var version = "dev"

func main() {
	cmd.Execute(version)
}
