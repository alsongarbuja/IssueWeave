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

// fetchMainIssue retrieves the details of the root issue from GitHub.
// It uses the provided client to query the specified owner and repo by issueId,
// returning an error if the API request fails.
func fetchMainIssue(client *github.Client, owner, repo string, issueId int) (IssueData, error) {
	ctx := context.Background()

	issue, _, err := client.Issues.Get(ctx, owner, repo, issueId)
	if err != nil {
		return IssueData{}, err
	}

	return IssueData{
		Number: issue.GetNumber(),
		Title: issue.GetTitle(),
		State: issue.GetState(),
		URL: issue.GetHTMLURL(),
		Date: issue.GetCreatedAt().Format("2006-01-02 15:04"),
	}, nil
}

// fetchCrossReferences paginates through the GitHub Timeline API for a specific issue
// to identify and process both explicit cross-references and inline comment mentions.
// It returns an error if the network request or pagination fails.
func fetchCrossReferences(client *github.Client, owner, repo string, issueId int) ([]IssueData, []IssueData, error) {
	ctx := context.Background()
	opts := &github.ListOptions{PerPage: 100}

	var allEvents []*github.Timeline

	for {
		events, resp, err := client.Issues.ListIssueTimeline(ctx, owner, repo, issueId, opts)
		if err != nil {
			return []IssueData{}, []IssueData{}, err
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
func fetchBatchedIssues(owner, repo string, uniqueIssues map[string]bool) ([]IssueData, error) {
	var comment_references []IssueData
	if len(uniqueIssues) == 0 {
		return comment_references, nil
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
		fmt.Errorf("failed to marshal payload: %w", err)
		return comment_references, err
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Errorf("failed to create request: %w", err)
		return comment_references, err
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Errorf("failed to execute request: %w", err)
		return comment_references, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Errorf("failed to read response: %w", err)
		return comment_references, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		fmt.Errorf("failed to parse JSON: %w", err)
		return comment_references, err
	}

	dataMap, ok := result["data"].(map[string]interface{})
	if !ok || dataMap == nil {
		return comment_references, fmt.Errorf("invalid GraphQL response format")
	}

	repoMap, ok := dataMap["repository"].(map[string]interface{})
	if !ok || repoMap == nil {
		return comment_references, fmt.Errorf("repository data not found: %w", err)
	}

	for key, issueDataInfer := range repoMap {
		if issueDataInfer == nil {
			continue
		}

		issueMap := issueDataInfer.(map[string]interface{})
		issueNumberStr := strings.TrimPrefix(key, "issue")
		issueNumber, _ := strconv.Atoi(issueNumberStr)

		comment_references = append(comment_references, IssueData{
			Number: issueNumber,
			Title:  fmt.Sprintf("%v", issueMap["title"]),
			State:  fmt.Sprintf("%v", issueMap["state"]),
			URL:    fmt.Sprintf("%v", issueMap["url"]),
		})
	}

	return comment_references, nil
}

// processEvents iterates through an issue's timeline to extract and display
// explicitly referenced issues, as well as outbound mentions found within comments.
// It uses the provided owner and repo to execute a batch fetch for any newly
// discovered issue IDs.
func processEvents(events []*github.Timeline, owner, repo string) ([]IssueData, []IssueData, error) {
	var direct_references []IssueData
	var comment_references []IssueData

	uniqueIssues := make(map[string]bool)
	refRegex := regexp.MustCompile(`(?i)(?:^|\s)#(\d+)\b`)

	for _, event := range events {
		eventType := event.GetEvent()
		if eventType == "cross-referenced" || eventType == "referenced" {
			source := event.GetSource()

			if source != nil && source.GetIssue() != nil {
				linkedIssue := source.GetIssue()

				direct_references = append(direct_references, IssueData{
					Number: linkedIssue.GetNumber(),
					Title:  linkedIssue.GetTitle(),
					State:  linkedIssue.GetState(),
					URL:    linkedIssue.GetHTMLURL(),
					Date:   event.GetCreatedAt().Format("2006-01-02 15:04"),
				})
			}
		}

		if eventType == "commented" {
			matches := refRegex.FindAllStringSubmatch(event.GetBody(), -1)
			for _, match := range matches {
				uniqueIssues[match[1]] = true
			}
		}
	}

	if len(uniqueIssues) > 0 {
		var err error
		comment_references, err = fetchBatchedIssues(owner, repo, uniqueIssues)
		if err != nil {
			return nil, nil, err
		}
	}

	return direct_references, comment_references, nil
}
