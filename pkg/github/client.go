package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	githubAPIURL = "https://api.github.com"
)

var copilotBotPatterns = []string{
	"copilot",
	"github-copilot",
	"copilot-preview",
}

// Client handles interactions with the GitHub API
type Client struct {
	token      string
	repository string
	httpClient *http.Client
}

// NewClient creates a new GitHub API client
func NewClient(token, repository string) *Client {
	return &Client{
		token:      token,
		repository: repository,
		httpClient: &http.Client{},
	}
}

// HandleWorkflowRun processes a workflow_run event
func (c *Client) HandleWorkflowRun(event *WorkflowRunEvent) error {
	fmt.Printf("\n--- Processing Workflow Run Event ---\n")
	fmt.Printf("Workflow Name: %s\n", event.WorkflowRun.Name)
	fmt.Printf("Workflow Run ID: %d\n", event.WorkflowRun.ID)
	fmt.Printf("Status: %s\n", event.WorkflowRun.Status)
	fmt.Printf("Conclusion: %s\n", event.WorkflowRun.Conclusion)
	fmt.Printf("Branch: %s\n", event.WorkflowRun.HeadBranch)
	fmt.Printf("Workflow URL: %s\n", event.WorkflowRun.HTMLURL)

	// Only process completed workflow runs
	if event.WorkflowRun.Status != "completed" {
		fmt.Printf("⏸️  Workflow run %d is not completed (status: %s), skipping\n",
			event.WorkflowRun.ID, event.WorkflowRun.Status)
		return nil
	}

	// Check if there are associated pull requests
	fmt.Printf("Associated PRs: %d\n", len(event.WorkflowRun.PullRequests))
	if len(event.WorkflowRun.PullRequests) == 0 {
		fmt.Printf("❌ Workflow run %d has no associated pull requests, skipping\n",
			event.WorkflowRun.ID)
		return nil
	}

	// Check each PR to see if it's from Copilot
	for _, pr := range event.WorkflowRun.PullRequests {
		fmt.Printf("\nChecking PR #%d...\n", pr.Number)
		fmt.Printf("Fetching PR details from GitHub API...\n")
		isCopilotPR, err := c.isCopilotPR(pr.Number)
		if err != nil {
			fmt.Printf("❌ Error checking if PR #%d is from Copilot: %v\n", pr.Number, err)
			continue
		}

		if !isCopilotPR {
			fmt.Printf("❌ PR #%d is not from Copilot, skipping\n", pr.Number)
			continue
		}

		fmt.Printf("✅ Confirmed Copilot PR #%d\n", pr.Number)
		fmt.Printf("Processing workflow conclusion: %s\n", event.WorkflowRun.Conclusion)

		// Handle based on workflow conclusion
		if event.WorkflowRun.Conclusion == "failure" {
			fmt.Printf("🔴 Handling failed workflow...\n")
			if err := c.handleFailedWorkflow(pr.Number, &event.WorkflowRun); err != nil {
				return fmt.Errorf("failed to handle failed workflow: %w", err)
			}
		} else if event.WorkflowRun.Conclusion == "success" {
			fmt.Printf("🟢 Handling successful workflow...\n")
			if err := c.handleSuccessfulWorkflow(pr.Number, &event.WorkflowRun); err != nil {
				return fmt.Errorf("failed to handle successful workflow: %w", err)
			}
		} else {
			fmt.Printf("ℹ️  Workflow conclusion '%s' - no action needed\n", event.WorkflowRun.Conclusion)
		}
	}

	return nil
}

// HandlePullRequest processes a pull_request event
func (c *Client) HandlePullRequest(event *PullRequestEvent) error {
	fmt.Printf("\n--- Processing Pull Request Event ---\n")
	fmt.Printf("PR Number: #%d\n", event.PullRequest.Number)
	fmt.Printf("PR Title: %s\n", event.PullRequest.Title)
	fmt.Printf("PR Author: %s (type: %s)\n", event.PullRequest.User.Login, event.PullRequest.User.Type)
	fmt.Printf("PR URL: %s\n", event.PullRequest.HTMLURL)
	fmt.Printf("Base Branch: %s\n", event.PullRequest.Base.Ref)
	fmt.Printf("Head Branch: %s\n", event.PullRequest.Head.Ref)

	// Check if the PR is from Copilot
	fmt.Printf("Checking if PR is from Copilot...\n")
	if !isCopilotUser(event.PullRequest.User.Login) {
		fmt.Printf("❌ PR #%d is NOT from Copilot (user: %s, type: %s), skipping\n",
			event.PullRequest.Number, event.PullRequest.User.Login, event.PullRequest.User.Type)
		return nil
	}

	fmt.Printf("✅ Detected Copilot PR #%d: %s\n", event.PullRequest.Number, event.PullRequest.Title)
	fmt.Printf("This PR will be monitored for workflow runs\n")
	return nil
}

// isCopilotPR checks if a PR was created by Copilot
func (c *Client) isCopilotPR(prNumber int) (bool, error) {
	url := fmt.Sprintf("%s/repos/%s/pulls/%d", githubAPIURL, c.repository, prNumber)
	fmt.Printf("  → API call: GET %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	fmt.Printf("  → API response: %d\n", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return false, err
	}

	isCopilot := isCopilotUser(pr.User.Login)
	fmt.Printf("  → User '%s' is Copilot: %v\n", pr.User.Login, isCopilot)
	return isCopilot, nil
}

// isCopilotUser checks if a username matches known Copilot bot patterns
func isCopilotUser(login string) bool {
	lowerLogin := strings.ToLower(login)

	// Check for exact matches or known patterns
	for _, pattern := range copilotBotPatterns {
		if lowerLogin == pattern ||
			lowerLogin == pattern+"[bot]" ||
			strings.HasPrefix(lowerLogin, pattern+"-") {
			return true
		}
	}

	return false
}

// handleFailedWorkflow handles a failed workflow run
func (c *Client) handleFailedWorkflow(prNumber int, workflow *WorkflowRun) error {
	fmt.Printf("\n--- Handling Failed Workflow ---\n")
	fmt.Printf("Workflow: '%s'\n", workflow.Name)
	fmt.Printf("PR: #%d\n", prNumber)

	// Get failed jobs
	fmt.Printf("Fetching workflow jobs...\n")
	jobs, err := c.getWorkflowJobs(workflow.ID)
	if err != nil {
		return fmt.Errorf("failed to get workflow jobs: %w", err)
	}
	fmt.Printf("Found %d total jobs\n", len(jobs))

	// Find failed jobs
	var failedJobs []Job
	for _, job := range jobs {
		if job.Conclusion == "failure" {
			failedJobs = append(failedJobs, job)
			fmt.Printf("  ❌ Failed job: %s (ID: %d)\n", job.Name, job.ID)
		} else {
			fmt.Printf("  ✅ Job: %s - %s\n", job.Name, job.Conclusion)
		}
	}

	if len(failedJobs) == 0 {
		fmt.Printf("No failed jobs found for workflow run %d\n", workflow.ID)
		return nil
	}

	// Get logs for failed jobs
	fmt.Printf("\nFetching logs for %d failed job(s)...\n", len(failedJobs))
	var logSnippets []string
	for _, job := range failedJobs {
		fmt.Printf("  → Fetching logs for job '%s' (ID: %d)...\n", job.Name, job.ID)
		logs, err := c.getJobLogs(job.ID)
		if err != nil {
			fmt.Printf("    ⚠️  Warning: failed to get logs for job %d: %v\n", job.ID, err)
			continue
		}
		fmt.Printf("    → Retrieved %d bytes of logs\n", len(logs))

		snippet := extractErrorSnippet(logs, 20)
		if snippet != "" {
			fmt.Printf("    → Extracted error snippet (%d chars)\n", len(snippet))
			logSnippets = append(logSnippets, fmt.Sprintf("**Job: %s**\n```\n%s\n```", job.Name, snippet))
		}
	}

	// Create comment
	fmt.Printf("\nBuilding failure comment...\n")
	comment := c.buildFailureComment(workflow, failedJobs, logSnippets)
	fmt.Printf("Comment length: %d characters\n", len(comment))

	fmt.Printf("Posting comment to PR #%d...\n", prNumber)
	if err := c.createComment(prNumber, comment); err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	fmt.Printf("✅ Successfully posted failure comment to PR #%d\n", prNumber)
	return nil
}

// handleSuccessfulWorkflow handles a successful workflow run
func (c *Client) handleSuccessfulWorkflow(prNumber int, workflow *WorkflowRun) error {
	fmt.Printf("\n--- Handling Successful Workflow ---\n")
	fmt.Printf("Workflow: '%s'\n", workflow.Name)
	fmt.Printf("PR: #%d\n", prNumber)

	comment := fmt.Sprintf("✅ **Workflow '%s' completed successfully!**\n\n[View workflow run](%s)",
		workflow.Name, workflow.HTMLURL)

	fmt.Printf("Building success comment...\n")
	fmt.Printf("Comment length: %d characters\n", len(comment))

	fmt.Printf("Posting comment to PR #%d...\n", prNumber)
	if err := c.createComment(prNumber, comment); err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	fmt.Printf("✅ Successfully posted success comment to PR #%d\n", prNumber)
	return nil
}

// getWorkflowJobs retrieves all jobs for a workflow run
func (c *Client) getWorkflowJobs(runID int64) ([]Job, error) {
	url := fmt.Sprintf("%s/repos/%s/actions/runs/%d/jobs", githubAPIURL, c.repository, runID)
	fmt.Printf("  → API call: GET %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("  → API response: %d\n", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var jobsResp JobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobsResp); err != nil {
		return nil, err
	}

	return jobsResp.Jobs, nil
}

// getJobLogs retrieves logs for a specific job
func (c *Client) getJobLogs(jobID int64) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/actions/jobs/%d/logs", githubAPIURL, c.repository, jobID)
	fmt.Printf("    → API call: GET %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("    → API response: %d\n", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// createComment creates a comment on a pull request
func (c *Client) createComment(prNumber int, body string) error {
	url := fmt.Sprintf("%s/repos/%s/issues/%d/comments", githubAPIURL, c.repository, prNumber)
	fmt.Printf("  → API call: POST %s\n", url)

	comment := Comment{Body: body}
	jsonData, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("  → API response: %d\n", resp.StatusCode)
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// buildFailureComment builds a formatted comment for workflow failures
func (c *Client) buildFailureComment(workflow *WorkflowRun, failedJobs []Job, logSnippets []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("❌ **Workflow '%s' failed**\n\n", workflow.Name))
	sb.WriteString(fmt.Sprintf("[View workflow run](%s)\n\n", workflow.HTMLURL))

	sb.WriteString("**Failed Jobs:**\n")
	for _, job := range failedJobs {
		sb.WriteString(fmt.Sprintf("- %s\n", job.Name))
	}
	sb.WriteString("\n")

	if len(logSnippets) > 0 {
		sb.WriteString("**Error Logs:**\n\n")
		for _, snippet := range logSnippets {
			sb.WriteString(snippet)
			sb.WriteString("\n\n")
		}
	}

	sb.WriteString("---\n")
	sb.WriteString("@copilot Please review the failure above and fix the issues to make the workflow pass.\n")

	return sb.String()
}

// extractErrorSnippet extracts the last N lines from logs, focusing on errors
func extractErrorSnippet(logs string, lines int) string {
	logLines := strings.Split(logs, "\n")

	// Find lines containing error keywords
	var errorLines []string
	keywords := []string{"error", "failed", "failure", "exception", "fatal"}

	for _, line := range logLines {
		lower := strings.ToLower(line)
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				errorLines = append(errorLines, line)
				break
			}
		}
	}

	// If we found error lines, return those (limited to 'lines' count)
	if len(errorLines) > 0 {
		if len(errorLines) > lines {
			return strings.Join(errorLines[len(errorLines)-lines:], "\n")
		}
		return strings.Join(errorLines, "\n")
	}

	// Otherwise, return last N lines
	if len(logLines) > lines {
		return strings.Join(logLines[len(logLines)-lines:], "\n")
	}
	return strings.Join(logLines, "\n")
}
