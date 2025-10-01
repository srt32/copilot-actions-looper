package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/srt32/copilot-actions-looper/pkg/github"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	eventName := os.Getenv("GITHUB_EVENT_NAME")
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	repository := os.Getenv("GITHUB_REPOSITORY")

	if eventPath == "" {
		log.Fatal("GITHUB_EVENT_PATH environment variable is required")
	}

	eventData, err := os.ReadFile(eventPath)
	if err != nil {
		log.Fatalf("Failed to read event file: %v", err)
	}

	client := github.NewClient(token, repository)

	switch eventName {
	case "workflow_run":
		var event github.WorkflowRunEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Fatalf("Failed to parse workflow_run event: %v", err)
		}
		if err := client.HandleWorkflowRun(&event); err != nil {
			log.Fatalf("Failed to handle workflow_run event: %v", err)
		}
	case "pull_request":
		var event github.PullRequestEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Fatalf("Failed to parse pull_request event: %v", err)
		}
		if err := client.HandlePullRequest(&event); err != nil {
			log.Fatalf("Failed to handle pull_request event: %v", err)
		}
	default:
		fmt.Printf("Ignoring event type: %s\n", eventName)
	}
}
