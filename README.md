# gitclean

A tool to clean up local git branches that have already been merged or squashed into a target branch (default: `main`).

## Requirements
- `git` (command-line tool)
- GitHub CLI (`gh`)

## Features
- Safely deletes local branches that have been merged or squashed into the target branch
- Checks each branch for a merged pull request using GitHub CLI (`gh pr view`); skips branches with open PRs
- Backs up your repository before deleting any branches
- Dry-run mode: preview branches that would be deleted without making changes
- Force mode: delete branches without confirmation
- Specify any target branch to compare against (default: origin/main)
- Configurable log level (debug, info, warn, error) with emoji and timestamped logs
- Fast: only fetches the target branch from origin

## Installation

```sh
go install github.com/jossec101/gitclean@latest
```

## Usage

```sh
gitclean --dryrun           # Show what would be deleted, but do not delete anything
gitclean --force            # Delete branches without asking for confirmation
gitclean --target=main      # Use a different target branch (default: origin/main)
gitclean --log-level=debug  # Set log level (debug, info, warn, error)
```

## License
MIT
