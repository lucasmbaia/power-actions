package github

import (
	"strings"

	gogithub "github.com/google/go-github/v33/github"
)

type Reviews struct {
	Review []Review `json:"reviews"`
}

type Review struct {
	File          string `json:"file"`
	LineNumber    int    `json:"lineNumber"`
	ReviewComment string `json:"reviewComment"`
}

type PullRequestReviewRequest struct {
	Comment  string
	Owner    string
	Repo     string
	PrNumber int
	Reviews  Reviews
}

func (c *Client) PullRequestReview(prr PullRequestReviewRequest) (err error) {
	var comments []*gogithub.DraftReviewComment

	for _, value := range prr.Reviews.Review {
		comments = append(comments, &gogithub.DraftReviewComment{
			Path:     gogithub.String(strings.Replace(value.File, "_", "/", -1)),
			Position: gogithub.Int(value.LineNumber),
			Body:     gogithub.String(value.ReviewComment),
		})
	}

	var review = &gogithub.PullRequestReviewRequest{
		Body:     gogithub.String(prr.Comment),
		Event:    gogithub.String("COMMENT"),
		Comments: comments,
	}

	_, _, err = c.Client.PullRequests.CreateReview(c.ctx, prr.Owner, prr.Repo, prr.PrNumber, review)

	return
}
