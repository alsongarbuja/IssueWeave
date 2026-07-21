package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [owner/repo] [issueId]",
	Short: "Weave issues for a specific repository and issue ID",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		repoArg := args[0]
		issueIdStr := args[1]

		parts := strings.Split(repoArg, "/")
		if len(parts) != 2 {
			fmt.Println("Error: Repostory must be in format owner/repo ('e.g., torvalds/linux')")
			return
		}
		owner, repo := parts[0], parts[1]

		issueId, err := strconv.Atoi(issueIdStr)
		if err != nil {
			fmt.Println("Error: Issue ID must be a number")
			return
		}

		fmt.Printf("Fetching chronological weave for %s/%s issue %d...\n", owner, repo, issueId)
	},
}
