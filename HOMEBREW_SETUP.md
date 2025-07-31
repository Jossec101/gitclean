# Homebrew Installation Setup

This document explains how to set up `gitclean` for installation via Homebrew using the same repository.

## Prerequisites

1. **GitHub Personal Access Token**: Create a personal access token with `repo` permissions for the automated formula updates (this is optional if you're using the same repository).

## Setup Steps

### 1. Using the Same Repository (Recommended for Single Tools)

The current configuration uses the same `gitclean` repository to host the Homebrew formula. This is simpler and doesn't require creating a separate tap repository.

### 2. Release Process

1. Create and push a new tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. The GitHub Action will automatically:
   - Build binaries for multiple platforms
   - Create a GitHub release
   - Create/update the Homebrew formula in the same repository

### 3. Installation for Users

Once set up, users can install your tool with:

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

## Manual Homebrew Formula (Alternative)

If you prefer not to use automated releases, you can create a manual formula:

1. In your `homebrew-tap` repository, create `Formula/gitclean.rb`:

```ruby
class Gitclean < Formula
  desc "Clean up local git branches that have been merged or squashed"
  homepage "https://github.com/Jossec101/gitclean"
  url "https://github.com/Jossec101/gitclean/archive/v1.0.0.tar.gz"
  sha256 "your-sha256-here"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    system "#{bin}/gitclean", "--help"
  end
end
```

## Testing

Test your formula locally:

```bash
# Install from your tap
brew install --build-from-source Jossec101/tap/gitclean

# Test the installation
gitclean --help
gitclean --version
```

## Notes

- The automated setup requires GoReleaser and GitHub Actions
- Make sure your repository is public for Homebrew distribution
- The formula will be automatically updated when you create new releases
- Users need `git` and `gh` CLI tools as dependencies (marked as optional in the formula)
