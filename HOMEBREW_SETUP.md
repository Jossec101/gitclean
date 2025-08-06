# Homebrew Installation Setup

This document explains how to set up `gitclean` for proper Homebrew distribution using a dedicated tap repository.

## Prerequisites

1. **Create a Homebrew Tap Repository**: You need to create a separate GitHub repository named `homebrew-gitclean`.
2. **GitHub Personal Access Token**: Create a personal access token with `repo` permissions for automated formula updates.

## Setup Steps

### 1. Create the Homebrew Tap Repository

1. Go to GitHub and create a new repository named `homebrew-gitclean`
2. Initialize it with a README
3. Clone it locally and create the basic structure:

```bash
git clone https://github.com/Jossec101/homebrew-gitclean.git
cd homebrew-gitclean
mkdir Formula
echo "# Homebrew Tap for gitclean" > README.md
git add .
git commit -m "Initial commit with Formula directory"
git push origin main
```

### 2. Set up GitHub Secrets

In your main `gitclean` repository settings, add the following secret:

1. Go to Settings → Secrets and variables → Actions
2. Add a new repository secret:
   - Name: `HOMEBREW_TAP_GITHUB_TOKEN`
   - Value: Your GitHub personal access token with repo permissions

### 3. Release Process

1. Create and push a new tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. The GitHub Action will automatically:
   - Build binaries for multiple platforms
   - Create a GitHub release
   - Update the Homebrew formula in the `homebrew-gitclean` repository

### 4. Installation for Users

Once set up, users can install your tool with proper Homebrew commands:

```bash
# Add your tap (one-time setup)
brew tap Jossec101/gitclean

# Install gitclean
brew install gitclean
```

Or in one command:
```bash
brew install Jossec101/gitclean/gitclean
```

## Testing

Test your Homebrew setup by building from source:

```bash
# Add the tap
brew tap Jossec101/gitclean

# Install from source (builds locally)
brew install --build-from-source gitclean

# Test the installation
gitclean --help
gitclean --version

# Verify it's properly installed via Homebrew
brew list gitclean
```

Note: `--build-from-source` will clone your repository and build the binary locally, which is perfect for testing the Homebrew setup before creating official releases.

## Notes

- The automated setup requires GoReleaser and GitHub Actions
- Make sure your repository is public for Homebrew distribution
- The formula will be automatically updated when you create new releases
- Users need `git` and `gh` CLI tools as dependencies