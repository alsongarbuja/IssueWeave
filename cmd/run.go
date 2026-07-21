package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v59/github"
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
		err = fetchMainIssue(client, owner, repo, issueId)
		if err != nil {
			fmt.Printf("Error fetching main issue: %v\n", err)
			return
		}

		// Fetch cross referenced issues
		err = fetchCrossReferences(client, owner, repo, issueId)
		if err != nil {
			fmt.Printf("Error fetching timeline: %v\n", err)
			return
		}
	},
}

// fetchMainIssue retrieves the details of the root issue from GitHub.
// It uses the provided client to query the specified owner and repo by issueId,
// returning an error if the API request fails.
func fetchMainIssue(client *github.Client, owner, repo string, issueId int) error {
	ctx := context.Background()

	issue, _, err := client.Issues.Get(ctx, owner, repo, issueId)
	if err != nil {
		return err
	}

	fmt.Println("========================================")
	fmt.Println("              ROOT ISSUE                ")
	fmt.Println("========================================")
	fmt.Printf("[%s] Created\n", issue.GetCreatedAt().Format("2006-01-02 15:04"))
	fmt.Printf(" - Issue #%d: %s\n", issue.GetNumber(), issue.GetTitle())
	fmt.Printf(" - State: %s\n", issue.GetState())
	fmt.Printf(" - URL: %s\n\n", issue.GetHTMLURL())

	return nil
}

// fetchCrossReferences paginates through the GitHub Timeline API for a specific issue
// to identify and process both explicit cross-references and inline comment mentions.
// It returns an error if the network request or pagination fails.
func fetchCrossReferences(client *github.Client, owner, repo string, issueId int) error {
	ctx := context.Background()
	opts := &github.ListOptions{PerPage: 100}

	var allEvents []*github.Timeline

	for {
		events, resp, err := client.Issues.ListIssueTimeline(ctx, owner, repo, issueId, opts)
		if err != nil {
			return err
		}

		allEvents = append(allEvents, events...)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return processEvents(allEvents, owner, repo)
}

// fetchBatchedIssues dynamically constructs and executes a single GraphQL query
// to retrieve metadata for multiple issues simultaneously, bypassing REST rate limits.
// It takes a map of deduplicated issue IDs and returns an error if the HTTP request
// or JSON parsing fails.
func fetchBatchedIssues(owner, repo string, uniqueIssues map[string]bool) error {
	if len(uniqueIssues) == 0 {
		return nil
	}

	var queryBuilder strings.Builder
	fmt.Fprintf(&queryBuilder, `query {
		repository(owner: "%s", name: "%s") {`, owner, repo)

	for issueNumStr := range uniqueIssues {
		fmt.Fprintf(&queryBuilder, `
			issue%s: issue(number: %s) {
				title
				url
				state
			}
		`, issueNumStr, issueNumStr)
	}

	queryBuilder.WriteString(`
		}
	}
	`)

	payload := map[string]string {
		"query": queryBuilder.String(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	dataMap, ok := result["data"].(map[string]interface{})
	if !ok || dataMap == nil {
		return fmt.Errorf("invalid GraphQL response format: %w", err)
	}

	repoMap, ok := dataMap["repository"].(map[string]interface{})
	if !ok || repoMap == nil {
		return fmt.Errorf("repository data not found: %w", err)
	}

	for key, issueDataInfer := range repoMap {
		if issueDataInfer == nil {
			continue
		}

		issueMap := issueDataInfer.(map[string]interface{})
		issueNumber := strings.TrimPrefix(key, "issue")

		fmt.Printf(" - Issue #%s: %s\n", issueNumber, issueMap["title"])
		fmt.Printf("   State: %s\n", issueMap["state"])
		fmt.Printf("   URL: %s\n\n", issueMap["url"])
	}

	return nil
}

// processEvents iterates through an issue's timeline to extract and display
// explicitly referenced issues, as well as outbound mentions found within comments.
// It uses the provided owner and repo to execute a batch fetch for any newly
// discovered issue IDs.
func processEvents(events []*github.Timeline, owner, repo string) error {
	foundReferences := false
	foundDirectReferences := 0

	uniqueIssues := make(map[string]bool)
	refRegex := regexp.MustCompile(`(?i)(?:^|\s)#(\d+)\b`)

	for _, event := range events {
		eventType := event.GetEvent()
		if eventType == "cross-referenced" || eventType == "referenced" {
			source := event.GetSource()

			if (foundDirectReferences == 1) {
				fmt.Println("----------------------------------------")
				fmt.Println("           CROSS-REFERENCES             ")
				fmt.Println("----------------------------------------")
			}

			if source != nil && source.GetIssue() != nil {
				foundReferences = true
				foundDirectReferences += 1
				linkedIssue := source.GetIssue()

				fmt.Printf("[%s] Cross-reference found:\n", event.GetCreatedAt().Format("2006-01-02 15:04"))
				fmt.Printf(" - Type: %T\n", linkedIssue)
				fmt.Printf(" - Issue #%d: %s\n", linkedIssue.GetNumber(), linkedIssue.GetTitle())
				fmt.Printf(" - State: %s\n", linkedIssue.GetState())
				fmt.Printf(" - URL: %s\n\n", linkedIssue.GetHTMLURL())
			}
		}

		if eventType == "commented" {
			body := event.GetBody()
			matches := refRegex.FindAllStringSubmatch(body, -1)
			for _, match := range matches {
				issueNum := match[1]
				uniqueIssues[issueNum] = true
			}
		}
	}

	if len(uniqueIssues) > 0 {
		foundReferences = true
		fmt.Println("----------------------------------------")
		fmt.Println("       FOUND IN COMMENTS (BATCHED)      ")
		fmt.Println("----------------------------------------")

		err := fetchBatchedIssues(owner, repo, uniqueIssues)
		if err != nil {
			fmt.Printf("Error fetching batched issues: %v\n", err)
		}
	}

	if !foundReferences {
		fmt.Println("No cross-references or comment mentions found for this issue.")
	}

	return nil
}
