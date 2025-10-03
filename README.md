# copilot-actions-looper

A GitHub Actions-based system that monitors Copilot-created pull requests and automatically provides feedback on workflow successes and failures.

## Features

- 🤖 Automatically detects pull requests created by GitHub Copilot
- 🔍 Monitors GitHub Actions workflows running on Copilot PRs
- ✅ Posts success comments when workflows pass
- ❌ Posts detailed failure comments with:
  - Links to failed workflow runs
  - Snippets of error logs
  - @-mentions to prompt Copilot to fix issues
- 🚀 Written primarily in Go with minimal bash usage
- ✨ Easy to install - just copy one workflow file

## Installation

1. Copy the workflow file to your repository:
   ```bash
   mkdir -p .github/workflows
   cp .github/workflows/monitor-copilot-prs.yml .github/workflows/
   ```

2. Ensure your repository has the required permissions:
   - `pull-requests: write` - to post comments on PRs
   - `actions: read` - to read workflow run information
   - `checks: read` - to read check statuses

3. (Optional) To have comments authored by a specific user instead of `github-actions[bot]`:
   - **Recommended**: Create a [fine-grained Personal Access Token](https://github.com/settings/tokens?type=beta) with:
     - Repository access: Select the target repository
     - Permissions: `Pull requests` (Read and write), `Issues` (Read and write)
   - Alternatively, create a [Personal Access Token (Classic)](https://github.com/settings/tokens) with:
     - `repo` scope (or at minimum: `public_repo` for public repositories)
     - Note: Classic tokens are being deprecated by GitHub in favor of fine-grained tokens
   - Add it as a repository secret named `PR_AUTHOR_TOKEN`
   - Comments will now appear as authored by the PAT owner instead of the bot

4. Commit and push the workflow file to your repository.

## How It Works

1. When you assign an issue to Copilot, it creates a pull request (standard GitHub behavior)
2. The monitor workflow listens for:
   - New pull requests (to detect Copilot PRs)
   - Workflow run completions (to check for failures)
3. When a workflow run completes on a Copilot PR:
   - **If successful**: Posts a success comment
   - **If failed**: Posts a detailed failure comment with error logs and @-mentions Copilot

## Example Comments

### Success Comment
```
✅ **Workflow 'CI' completed successfully!**

[View workflow run](https://github.com/owner/repo/actions/runs/123)
```

### Failure Comment
```
❌ **Workflow 'CI' failed**

[View workflow run](https://github.com/owner/repo/actions/runs/123)

**Failed Jobs:**
- Build
- Test

**Error Logs:**

**Job: Build**
```
ERROR: Build failed at step xyz
```

**Job: Test**
```
FAIL: TestSomething failed
```

---
@copilot Please review the failure above and fix the issues to make the workflow pass.
```

## Development

### Prerequisites
- Go 1.21 or later

### Building
```bash
go build -o monitor ./cmd/monitor
```

### Testing
```bash
go test ./... -v
```

### Code Formatting
```bash
go fmt ./...
go vet ./...
```

## Project Structure

```
.
├── .github/
│   └── workflows/
│       ├── monitor-copilot-prs.yml  # Main workflow file
│       └── ci.yml                   # CI workflow for testing
├── cmd/
│   └── monitor/
│       └── main.go                   # Application entry point
├── pkg/
│   └── github/
│       ├── client.go                 # GitHub API client
│       ├── client_test.go            # Tests
│       └── types.go                  # Data structures
├── testapp/
│   ├── math.go                       # Example code for testing
│   └── math_test.go                  # Tests for example code
├── go.mod
├── go.sum
└── README.md
```

## Testing the Monitor

This repository includes a CI workflow (`.github/workflows/ci.yml`) that tests both the monitor code and a simple test application. The test application (`testapp/`) contains a basic `Add` function that can be used to verify the monitoring system works correctly.

To test the monitoring system:
1. The CI workflow runs automatically on push and pull requests
2. If tests fail, the monitor workflow should detect the failure and comment on Copilot PRs
3. You can intentionally break the `testapp/math.go` function to test failure detection

## License

MIT
