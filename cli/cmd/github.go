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
)

// getGitHubClient initializes and returns a GitHub API client.
// It automatically applies an authentication token if the GITHUB_TOKEN environment
// variable is set, otherwise returning an unauthenticated client subject to stricter
// rate limits.
func getGithubClient() *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return github.NewClient(nil)
	}

	return github.NewClient(nil).WithAuthToken(token)
}

// fetchMainIssue retrieves the details of the root issue from GitHub.
// It uses the provided client to query the specified owner and repo by issueId,
/// returning the issue data and error if api fails.
func fetchMainIssue(client *github.Client, owner, repo string, issueId int) (IssueData, error) {
	ctx := context.Background()

	issue, _, err := client.Issues.Get(ctx, owner, repo, issueId)
	if err != nil {
		return IssueData{}, err
	}

	comments, _ := fetchIssueComments(client, owner, repo, issueId)

	return IssueData{
		Number: issue.GetNumber(),
		Title: issue.GetTitle(),
		State: issue.GetState(),
		URL: issue.GetHTMLURL(),
		Date: issue.GetCreatedAt().Format("2006-01-02 15:04"),
		Comments: comments,
	}, nil
}

// fetchCrossReferences paginates through the GitHub Timeline API for a specific issue
// to identify and process both explicit cross-references and inline comment mentions.
// It returns cross referenced and comment referenced issues and an error if the network request or pagination fails.
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
				comments(first: 20) {
					nodes {
						author { login }
						body
						createdAt
					}
				}
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
		return comment_references, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return comment_references, fmt.Errorf("failed to create request: %w", err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return comment_references, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return comment_references, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return comment_references, fmt.Errorf("failed to parse JSON: %w", err)
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

		var comments []Comment
		if commentsData, ok := issueMap["comments"].(map[string]interface{}); ok {
			if nodes, ok := commentsData["nodes"].([]interface{}); ok {
				for _, nodeInter := range nodes {
					if nodeInter == nil { continue }
					node := nodeInter.(map[string]interface{})

					author := "Unknown"
					if authorMap, ok := node["author"].(map[string]interface{}); ok && authorMap != nil {
						author = fmt.Sprintf("%v", authorMap["login"])
					}

					comments = append(comments, Comment{
						Author: author,
						Body:   fmt.Sprintf("%v", node["body"]),
						Date:   fmt.Sprintf("%v", node["createdAt"]),
					})
				}
			}
		}

		comment_references = append(comment_references, IssueData{
			Number: issueNumber,
			Title:  fmt.Sprintf("%v", issueMap["title"]),
			State:  fmt.Sprintf("%v", issueMap["state"]),
			URL:    fmt.Sprintf("%v", issueMap["url"]),
		})
	}

	return comment_references, nil
}

// fetchIssueComments retreives the first 20 comments of the issue
// It uses the provided client to query the specified owner and repo by issueId,
// returning a list of comments and error if api fails.
func fetchIssueComments(client *github.Client, owner, repo string, issueNumber int) ([]Comment, error) {
	ctx := context.Background()
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 20},
	}

	githubComments, _, err := client.Issues.ListComments(ctx, owner, repo, issueNumber, opts)
	if err != nil {
		return nil, err
	}

	var comments []Comment
	for _, c := range githubComments {
		author := "Unknown"
		if c.GetUser() != nil {
			author = c.GetUser().GetLogin()
		}
		comments = append(comments, Comment {
			Author: author,
			Body: c.GetBody(),
			Date: c.GetCreatedAt().Format("2006-01-02 15:04"),
		})
	}

	return comments, nil
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

				comments, _ := fetchIssueComments(getGithubClient(), owner, repo, linkedIssue.GetNumber())

				direct_references = append(direct_references, IssueData{
					Number: linkedIssue.GetNumber(),
					Title:  linkedIssue.GetTitle(),
					State:  linkedIssue.GetState(),
					URL:    linkedIssue.GetHTMLURL(),
					Date:   event.GetCreatedAt().Format("2006-01-02 15:04"),
					Comments: comments,
				})
			}
		}

		if eventType == "commented" {
			matches := refRegex.FindAllStringSubmatch(parseIssueUrl(owner, repo, event.GetBody()), -1)
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


func parseIssueUrl(owner, repo, commentBody string) string {
	safeOwner := regexp.QuoteMeta(owner)
	safeRepo := regexp.QuoteMeta(repo)

	githubUrl := fmt.Sprintf(`https://github\.com/%s/%s/issues/(\d+/)?`, safeOwner, safeRepo)
	issueUrlRegex := regexp.MustCompile(githubUrl)

	updatedCommentBody := issueUrlRegex.ReplaceAllString(commentBody, "#$1")
	return updatedCommentBody
}
