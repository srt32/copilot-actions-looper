package github

import "time"

// WorkflowRunEvent represents a workflow_run event from GitHub
type WorkflowRunEvent struct {
	Action      string      `json:"action"`
	WorkflowRun WorkflowRun `json:"workflow_run"`
	Repository  Repository  `json:"repository"`
}

// PullRequestEvent represents a pull_request event from GitHub
type PullRequestEvent struct {
	Action      string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
}

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	ID           int64         `json:"id"`
	Name         string        `json:"name"`
	HeadBranch   string        `json:"head_branch"`
	HeadSHA      string        `json:"head_sha"`
	Status       string        `json:"status"`
	Conclusion   string        `json:"conclusion"`
	HTMLURL      string        `json:"html_url"`
	PullRequests []PullRequest `json:"pull_requests"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
	Title   string `json:"title"`
	User    User   `json:"user"`
	Head    Head   `json:"head"`
	Base    Base   `json:"base"`
}

// User represents a GitHub user
type User struct {
	Login string `json:"login"`
	Type  string `json:"type"`
}

// Head represents the head of a PR
type Head struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// Base represents the base of a PR
type Base struct {
	Ref string `json:"ref"`
}

// Repository represents a GitHub repository
type Repository struct {
	FullName string `json:"full_name"`
	Owner    Owner  `json:"owner"`
	Name     string `json:"name"`
}

// Owner represents a repository owner
type Owner struct {
	Login string `json:"login"`
}

// Job represents a GitHub Actions job
type Job struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	StartedAt  time.Time `json:"started_at"`
	Steps      []Step    `json:"steps"`
}

// Step represents a step in a GitHub Actions job
type Step struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Number     int    `json:"number"`
}

// JobsResponse represents the response from the jobs API
type JobsResponse struct {
	TotalCount int   `json:"total_count"`
	Jobs       []Job `json:"jobs"`
}

// Comment represents a GitHub comment
type Comment struct {
	Body string `json:"body"`
}
