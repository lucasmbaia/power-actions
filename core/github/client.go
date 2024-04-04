package github

import (
	"context"

	gogithub "github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

type Client struct {
	ctx    context.Context
	token  string
	Client *gogithub.Client
}

func NewClient(token string) (c Client) {
	c.ctx = context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(c.ctx, ts)

	c.Client = gogithub.NewClient(tc)
	c.token = token

	return
}
