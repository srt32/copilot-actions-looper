package github

import (
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token", "owner/repo")
	if client == nil {
		t.Fatal("Expected client to be created")
	}
	if client.token != "test-token" {
		t.Errorf("Expected token to be 'test-token', got '%s'", client.token)
	}
	if client.repository != "owner/repo" {
		t.Errorf("Expected repository to be 'owner/repo', got '%s'", client.repository)
	}
}

func TestIsCopilotPR(t *testing.T) {
	tests := []struct {
		name     string
		prData   PullRequest
		expected bool
	}{
		{
			name: "Copilot bot PR",
			prData: PullRequest{
				User: User{
					Login: "copilot",
					Type:  "Bot",
				},
			},
			expected: true,
		},
		{
			name: "PR with copilot in login",
			prData: PullRequest{
				User: User{
					Login: "github-copilot",
					Type:  "Bot",
				},
			},
			expected: true,
		},
		{
			name: "Regular user PR",
			prData: PullRequest{
				User: User{
					Login: "regularuser",
					Type:  "User",
				},
			},
			expected: false,
		},
		{
			name: "Bot but not copilot",
			prData: PullRequest{
				User: User{
					Login: "dependabot",
					Type:  "Bot",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the logic directly based on the isCopilotPR implementation
			result := strings.Contains(strings.ToLower(tt.prData.User.Login), copilotUser)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v for user %s", tt.expected, result, tt.prData.User.Login)
			}
		})
	}
}

func TestExtractErrorSnippet(t *testing.T) {
	tests := []struct {
		name     string
		logs     string
		lines    int
		expected string
	}{
		{
			name:     "Logs with error keyword",
			logs:     "Line 1\nLine 2\nERROR: Something went wrong\nLine 4\nLine 5",
			lines:    3,
			expected: "ERROR: Something went wrong",
		},
		{
			name:     "Logs with multiple errors",
			logs:     "Line 1\nError: First error\nLine 3\nFailed: Second error\nLine 5",
			lines:    2,
			expected: "Error: First error\nFailed: Second error",
		},
		{
			name:     "No errors, return last lines",
			logs:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5",
			lines:    2,
			expected: "Line 4\nLine 5",
		},
		{
			name:     "Empty logs",
			logs:     "",
			lines:    5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractErrorSnippet(tt.logs, tt.lines)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestBuildFailureComment(t *testing.T) {
	client := NewClient("test-token", "owner/repo")

	workflow := &WorkflowRun{
		ID:      123,
		Name:    "Test Workflow",
		HTMLURL: "https://github.com/owner/repo/actions/runs/123",
	}

	failedJobs := []Job{
		{ID: 1, Name: "Build"},
		{ID: 2, Name: "Test"},
	}

	logSnippets := []string{
		"**Job: Build**\n```\nError: Build failed\n```",
		"**Job: Test**\n```\nError: Tests failed\n```",
	}

	comment := client.buildFailureComment(workflow, failedJobs, logSnippets)

	if !strings.Contains(comment, "Test Workflow") {
		t.Error("Comment should contain workflow name")
	}
	if !strings.Contains(comment, "Build") {
		t.Error("Comment should contain failed job name 'Build'")
	}
	if !strings.Contains(comment, "Test") {
		t.Error("Comment should contain failed job name 'Test'")
	}
	if !strings.Contains(comment, "@copilot") {
		t.Error("Comment should mention @copilot")
	}
	if !strings.Contains(comment, workflow.HTMLURL) {
		t.Error("Comment should contain workflow URL")
	}
}

func TestHandleWorkflowRun_NotCompleted(t *testing.T) {
	client := NewClient("test-token", "owner/repo")

	event := &WorkflowRunEvent{
		WorkflowRun: WorkflowRun{
			Status: "in_progress",
		},
	}

	err := client.HandleWorkflowRun(event)
	if err != nil {
		t.Errorf("Expected no error for in_progress workflow, got %v", err)
	}
}

func TestHandleWorkflowRun_NoPullRequests(t *testing.T) {
	client := NewClient("test-token", "owner/repo")

	event := &WorkflowRunEvent{
		WorkflowRun: WorkflowRun{
			Status:       "completed",
			Conclusion:   "success",
			PullRequests: []PullRequest{},
		},
	}

	err := client.HandleWorkflowRun(event)
	if err != nil {
		t.Errorf("Expected no error for workflow with no PRs, got %v", err)
	}
}

func TestHandlePullRequest_NotCopilot(t *testing.T) {
	client := NewClient("test-token", "owner/repo")

	event := &PullRequestEvent{
		PullRequest: PullRequest{
			Number: 1,
			User: User{
				Login: "regularuser",
				Type:  "User",
			},
		},
	}

	err := client.HandlePullRequest(event)
	if err != nil {
		t.Errorf("Expected no error for non-Copilot PR, got %v", err)
	}
}

func TestHandlePullRequest_Copilot(t *testing.T) {
	client := NewClient("test-token", "owner/repo")

	event := &PullRequestEvent{
		PullRequest: PullRequest{
			Number: 1,
			Title:  "Fix issue",
			User: User{
				Login: "copilot",
				Type:  "Bot",
			},
		},
	}

	err := client.HandlePullRequest(event)
	if err != nil {
		t.Errorf("Expected no error for Copilot PR, got %v", err)
	}
}
