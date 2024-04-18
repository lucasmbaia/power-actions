package services

import (
	"context"

	gogithub "github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

// Client represents a GitHub client
type GitHubClient struct {
	ctx    context.Context
	token  string
	Client *gogithub.Client
}

// New creates a new GitHub client with the provided token
func NewGitHubClient(token string) (c GitHubClient) {
	c.ctx = context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(c.ctx, ts)

	c.Client = gogithub.NewClient(tc)
	c.token = token

	return c
}
