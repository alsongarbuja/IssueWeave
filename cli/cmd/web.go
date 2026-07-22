package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(webCommand)
}

var webCommand = &cobra.Command {
	Use: "web [owner/repo] [issueId]",
	Short: "Start a local server to serve the issue data",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		repoArgs := args[0]
		issueIdStr := args[1]

		parts := strings.Split(repoArgs, "/")
		if len(parts) != 2 {
			fmt.Println("Error: Format of first argument must be owner/repo")
			return
		}

		owner, repo := parts[0], parts[1]
		issueId, _ := strconv.Atoi(issueIdStr)

		http.HandleFunc("/api/weave", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/json")

			client := getGithubClient()

			rootIssue, err := fetchMainIssue(client, owner, repo, issueId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			d_refs, c_refs, err := fetchCrossReferences(client, owner, repo, issueId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			result := WeaveResult {
				MainIssue: rootIssue,
				DirectReferences: d_refs,
				CommentReferences: c_refs,
			}

			json.NewEncoder(w).Encode(result)
		})

		fmt.Printf("API Server running...\n")
		fmt.Printf("Data available at: http://localhost:8080/api/weave\n")
		http.ListenAndServe(":8080", nil)
	},
}
