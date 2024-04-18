package services

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type GitRepoInfo struct {
	CurrentBranch   string
	PrincipalBranch string
	RepositoryName  string
	RepositoryOwner string
	OriginURL       string
}

// getGitRepoInfo returns details about the current Git repository
func GetGitRepoInfo() (*GitRepoInfo, error) {
	info := &GitRepoInfo{}

	// Get current branch
	if output, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	} else {
		info.CurrentBranch = strings.TrimSpace(string(output))
	}

	// Get remote details including the principal branch
	cmd := exec.Command("git", "remote", "show", "origin")
	var cmdOutput bytes.Buffer
	cmd.Stdout = &cmdOutput
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get remote details: %w", err)
	}
	remoteDetails := strings.Split(cmdOutput.String(), "\n")
	for _, line := range remoteDetails {
		if strings.Contains(line, "HEAD branch") {
			info.PrincipalBranch = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}

	// Get the repository URL and parse the owner and repository name
	if output, err := exec.Command("git", "config", "--get", "remote.origin.url").Output(); err != nil {
		return nil, fmt.Errorf("failed to get remote origin URL: %w", err)
	} else {
		info.OriginURL = strings.TrimSpace(string(output))
		// Parse the owner and repository name from the URL
		repoURL := strings.TrimSpace(string(output))
		repoURL = strings.TrimSuffix(repoURL, ".git")
		// Typical repo URLs are either SSH format (git@github.com:owner/repo) or HTTPS format (https://github.com/owner/repo)
		if pos := strings.Index(repoURL, "@"); pos != -1 {
			// SSH format
			repoURL = repoURL[pos+1:]                       // Remove the 'git@'
			repoURL = strings.Replace(repoURL, ":", "/", 1) // Replace ':' with '/'
		}
		if pos := strings.Index(repoURL, "://"); pos != -1 {
			// HTTPS format
			repoURL = repoURL[pos+3:] // Remove 'https://'
		}
		parts := strings.Split(repoURL, "/")
		if len(parts) >= 3 {
			info.RepositoryOwner = parts[1]
			info.RepositoryName = parts[2]
		}
	}

	return info, nil
}
