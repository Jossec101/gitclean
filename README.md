# gitclean

A tool to clean up local git branches that have already been merged or squashed into a target branch (default: `main`).

## Features
- Checks each branch for a merged PR using `gh pr view` and excludes branches with open PRs from deletion
- Detects branches merged or squashed into `main`/`master` (or any target branch)
- Uses GitHub CLI (`gh`) to check for open/merged PRs for extra safety
- Backs up your repo before deleting branches
- Supports dry-run mode (shows what would be deleted)
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
