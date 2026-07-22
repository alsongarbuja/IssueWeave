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

// * command "run" to get the issues batched using cobra
var runCmd = &cobra.Command{
	Use:   "run [owner/repo] [issueId]",
	Short: "Weave issues for a specific repository and issue ID",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		repoArg := args[0]
		issueIdStr := args[1]

		// Get owner and repo from the first argument
		parts := strings.Split(repoArg, "/")
		if len(parts) != 2 {
			fmt.Println("Error: Repostory must be in format owner/repo ('e.g., torvalds/linux')")
			return
		}
		owner, repo := parts[0], parts[1]

		// Get the issue Id from the second argument
		issueId, err := strconv.Atoi(issueIdStr)
		if err != nil {
			fmt.Println("Error: Issue ID must be a number")
			return
		}

		fmt.Printf("Fetching chronological weave for %s/%s issue %d...\n", owner, repo, issueId)

		// Get the github client
		client := getGithubClient()

		// Fetch the main issue
		main_issue, err := fetchMainIssue(client, owner, repo, issueId)
		if err != nil {
			fmt.Printf("Error fetching main issue: %v\n", err)
			return
		}

		fmt.Print(main_issue)

		// Fetch cross referenced issues
		direct_refs, comment_refs, err := fetchCrossReferences(client, owner, repo, issueId)
		if err != nil {
			fmt.Printf("Error fetching timeline: %v\n", err)
			return
		}

		fmt.Print(direct_refs, comment_refs)
	},
}
