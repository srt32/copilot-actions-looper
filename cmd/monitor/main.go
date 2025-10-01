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
	runID := os.Getenv("GITHUB_RUN_ID")

	fmt.Printf("=== GitHub Actions Monitor Starting ===\n")
	fmt.Printf("Repository: %s\n", repository)
	fmt.Printf("Event Type: %s\n", eventName)
	fmt.Printf("Event Path: %s\n", eventPath)
	fmt.Printf("Run ID: %s\n", runID)
	fmt.Printf("=====================================\n\n")

	if eventPath == "" {
		log.Fatal("GITHUB_EVENT_PATH environment variable is required")
	}

	fmt.Printf("Reading event data from: %s\n", eventPath)
	eventData, err := os.ReadFile(eventPath)
	if err != nil {
		log.Fatalf("Failed to read event file: %v", err)
	}
	fmt.Printf("Successfully read %d bytes of event data\n\n", len(eventData))

	client := github.NewClient(token, repository)

	switch eventName {
	case "workflow_run":
		fmt.Printf("Processing workflow_run event...\n")
		var event github.WorkflowRunEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Fatalf("Failed to parse workflow_run event: %v", err)
		}
		fmt.Printf("Event action: %s\n", event.Action)
		if err := client.HandleWorkflowRun(&event); err != nil {
			log.Fatalf("Failed to handle workflow_run event: %v", err)
		}
	case "pull_request":
		fmt.Printf("Processing pull_request event...\n")
		var event github.PullRequestEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Fatalf("Failed to parse pull_request event: %v", err)
		}
		fmt.Printf("Event action: %s\n", event.Action)
		if err := client.HandlePullRequest(&event); err != nil {
			log.Fatalf("Failed to handle pull_request event: %v", err)
		}
	default:
		fmt.Printf("Ignoring event type: %s\n", eventName)
	}
	fmt.Printf("\n=== Monitor completed successfully ===\n")
}
