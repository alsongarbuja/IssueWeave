package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rateCmd)
}

var rateCmd = &cobra.Command{
	Use:   "rate",
	Short: "Check your remaining GitHub API rate limits",
	Run: func(cmd *cobra.Command, args []string) {
		client := getGithubClient()
		ctx := context.Background()

		limits, _, err := client.RateLimit.Get(ctx)
		if err != nil {
			fmt.Printf("Error fetching rate limits: %v\n", err)
			return
		}

		core := limits.Core
		graphql := limits.GraphQL

		fmt.Println("========================================")
		fmt.Println("          GITHUB API RATE LIMITS        ")
		fmt.Println("========================================")

		fmt.Printf("REST API (Core):\n")
		fmt.Printf(" - Remaining: %d / %d\n", core.Remaining, core.Limit)
		fmt.Printf(" - Resets At: %s (in %s)\n\n",
			core.Reset.Time.Format("15:04:05 MST"),
			time.Until(core.Reset.Time).Round(time.Minute))

		fmt.Printf("GraphQL API:\n")
		fmt.Printf(" - Remaining: %d / %d\n", graphql.Remaining, graphql.Limit)
		fmt.Printf(" - Resets At: %s (in %s)\n",
			graphql.Reset.Time.Format("15:04:05 MST"),
			time.Until(graphql.Reset.Time).Round(time.Minute))
	},
}
