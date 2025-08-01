package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func getLocalBranches(target string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-l")
	cmd.Env = append(os.Environ(), "GIT_PAGER=cat") // Prevent pager
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	branches := []string{}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		branch := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "* "))
		if branch != target {
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

func isSquashed(branch, target string) (bool, error) {
	cmd := exec.Command("git", "merge", "--no-commit", "--no-ff", branch)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GIT_MERGE_AUTOEDIT=no")
	err := cmd.Run()
	if err == nil {
		// No changes, likely squashed
		abortCmd := exec.Command("git", "merge", "--abort")
		if abortErr := abortCmd.Run(); abortErr != nil {
			return true, fmt.Errorf("merge abort failed: %v", abortErr)
		}
		return true, nil
	}
	abortCmd := exec.Command("git", "merge", "--abort")
	_ = abortCmd.Run() // ignore error here
	return false, nil
}

func backupRepo() (string, error) {
	var backupDir string
	var globalBackupDir string
	if runtime.GOOS == "windows" {
		tmp := os.Getenv("TEMP")
		if tmp == "" {
			tmp = "C:\\Temp"
		}
		globalBackupDir = filepath.Join(tmp, "gitclean-backups")
	} else {
		globalBackupDir = "/tmp/gitclean-backups"
	}
	if err := os.MkdirAll(globalBackupDir, 0755); err != nil {
		return "", err
	}
	backupDir = filepath.Join(globalBackupDir, "active")
	// Remove any existing backup before copying
	if err := os.RemoveAll(backupDir); err != nil {
		return "", fmt.Errorf("failed to remove existing backup directory: %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		cmd := exec.Command("xcopy", cwd, backupDir, "/E", "/I", "/Q", "/Y", "/EXCLUDE:.gitignore")
		if err := cmd.Run(); err != nil {
			return "", err
		}
	} else {
		cmd := exec.Command("rsync", "-a", "--exclude=.gitignore", cwd+"/", backupDir)
		if err := cmd.Run(); err != nil {
			return "", err
		}
	}
	return backupDir, nil
}

func checkCLITool(tool string) bool {
	cmd := exec.Command(tool, "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	app := &cli.App{
		Name:    "gitclean",
		Usage:   "Clean up local git branches that have been merged or squashed into a target branch.",
		Version: version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dryrun",
				Usage: "Show what would be deleted, but do not delete anything.",
			},
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Delete branches without asking for confirmation.",
			},
			&cli.StringFlag{
				Name:  "target",
				Usage: "Target branch to compare against (default: origin/main).",
				Value: "origin/main",
			},
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "Set log level (debug, info, warn, error). Default: info.",
				Value: "info",
			},
		},
		Action: func(c *cli.Context) error {
			if !checkCLITool("git") {
				log.Error("🛑 'git' CLI not found. Please install git and ensure it is in your PATH.")
				return fmt.Errorf("git CLI not found")
			}
			if !checkCLITool("gh") {
				log.Error("🛑 'gh' CLI not found. Please install GitHub CLI and ensure it is in your PATH.")
				return fmt.Errorf("gh CLI not found")
			}
			levelStr := strings.ToLower(c.String("log-level"))
			logLevel := log.InfoLevel
			switch levelStr {
			case "debug":
				logLevel = log.DebugLevel
			case "info":
				logLevel = log.InfoLevel
			case "warn":
				logLevel = log.WarnLevel
			case "error":
				logLevel = log.ErrorLevel
			default:
				log.Warnf("Unknown log level: %s, defaulting to info", levelStr)
				logLevel = log.InfoLevel
			}
			log.SetLevel(logLevel)
			log.Debugf("Log level set to: %s", logLevel.String())

			target := c.String("target")
			branchName := strings.TrimPrefix(target, "origin/")
			if !strings.HasPrefix(target, "origin/") {
				target = "origin/" + target
			}
			dryrun := c.Bool("dryrun")
			force := c.Bool("force")

			log.Infof("🔄 Fetching latest changes for %s from origin...", branchName)
			fetchCmd := exec.Command("git", "fetch", "origin", branchName)
			if err := fetchCmd.Run(); err != nil {
				log.WithError(err).Warnf("🛑 git fetch origin %s failed, continuing with local refs.", branchName)
			}

			log.Debug("Listing local branches...")
			branches, err := getLocalBranches(branchName)
			if err != nil {
				log.WithError(err).Error("🛑 Error listing branches")
				return err
			}
			log.Debugf("Detected local branches: %v", branches)
			log.Infof("%d local branches detected", len(branches))

			log.Debug("Selecting candidate branches for deletion...")
			candidates := []struct {
				name       string
				confidence string
			}{}
			mergedSet := make(map[string]bool)
			for _, branch := range branches {
				// Check if branch has a PR using gh pr view
				cmd := exec.Command("gh", "pr", "view", branch, "--json", "mergedAt,title,body,state")
				out, err := cmd.Output()
				if err == nil {
					var pr struct {
						MergedAt *string `json:"mergedAt"`
						State    string  `json:"state"`
					}
					err = json.Unmarshal(out, &pr)
					if err == nil {
						if pr.MergedAt != nil {
							log.Debugf("Branch %s: merged PR detected", branch)
							candidates = append(candidates, struct {
								name       string
								confidence string
							}{branch, "🟢 high (merged PR)"})
							continue
						} else if pr.State == "OPEN" {
							log.Infof("Branch %s: open PR detected, skipping", branch)
							continue
						}
					}
				}
				if mergedSet[branch] {
					log.Debugf("Branch %s: merged", branch)
					candidates = append(candidates, struct {
						name       string
						confidence string
					}{branch, "🟢 high (merged)"})
				} else {
					log.Debugf("Checking if branch %s is squashed...", branch)
					squashed, err := isSquashed(branch, target)
					log.Debugf("Branch %s squashed: %v, error: %v", branch, squashed, err)
					if err != nil {
						log.WithError(err).Warnf("🛑 Error checking if branch %s is squashed", branch)
						continue
					}
					if squashed {
						log.Debugf("Branch %s: squashed", branch)
						candidates = append(candidates, struct {
							name       string
							confidence string
						}{branch, "🟡 medium (squashed)"})
					}
				}
			}

			log.Infof("Branches that would be deleted (%d):", len(candidates))
			if len(branches) > 0 {
				percent := float64(len(candidates)) / float64(len(branches)) * 100
				log.Infof("%.1f%% of local branches are candidates for deletion", percent)
			}
			for _, c := range candidates {
				log.Infof("- %s [%s]", c.name, c.confidence)
			}

			if dryrun {
				log.Info("🧹 Dry run: no branches deleted.")
				return nil
			}

			log.Debug("Backing up repository before deletion...")
			backupPath, err := backupRepo()
			if err != nil {
				log.WithError(err).Error("🛑 Backup failed, aborting deletion.")
				return err
			}
			log.Infof("💾 Backup created at: %s", backupPath)
			log.Info("💾 A backup of your repository has been created before deletion.")

			if !force {
				if len(branches) > 0 {
					percent := float64(len(candidates)) / float64(len(branches)) * 100
					log.Infof("About to delete %.1f%% of local branches", percent)
				}
				log.Info("Proceed with deletion? (y/N): ")
				var resp string
				scan := bufio.NewScanner(os.Stdin)
				if scan.Scan() {
					resp = strings.ToLower(strings.TrimSpace(scan.Text()))
				}
				if resp != "y" && resp != "yes" {
					log.Info("Aborted.")
					return nil
				}
			}

			deleted := 0
			for _, c := range candidates {
				log.Debugf("Deleting branch: %s", c.name)
				cmd := exec.Command("git", "branch", "-D", c.name)
				if err := cmd.Run(); err == nil {
					deleted++
					log.Infof("🗑️ Deleted branch: %s", c.name)
				} else {
					log.WithError(err).Errorf("🛑 Failed to delete branch %s", c.name)
				}
			}
			log.Infof("Total branches deleted: %d", deleted)
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Fatal("Application error")
	}
}
