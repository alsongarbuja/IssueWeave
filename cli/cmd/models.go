package cmd

// Comment struct
type Comment struct {
	Author string `json:"author"`
	Body string `json:"body"`
	Date string `json:"date"`
}

// Single issue struct
type IssueData struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
	Date   string `json:"date,omitempty"`
	Comments []Comment `json:"comments"`
}

// WeaveResult
type WeaveResult struct {
	MainIssue IssueData `json:"main_issue"`
	DirectReferences []IssueData `json:"direct_references"`
	CommentReferences []IssueData `json:"comment_references"`
}
