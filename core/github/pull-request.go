package github

import (
	"fmt"

	gogithub "github.com/google/go-github/v33/github"
	"github.com/lucasmbaia/power-actions/core/prompt"
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
	Comment         string
	Owner           string
	Repo            string
	PrNumber        int
	Reviews         Reviews
	MaxChangedLines int
}

func (c *Client) PullRequestReview(prr PullRequestReviewRequest) (err error) {
	var comments []*gogithub.DraftReviewComment

	for _, value := range prr.Reviews.Review {
		comments = append(comments, &gogithub.DraftReviewComment{
			Path:     gogithub.String(value.File),
			Position: gogithub.Int(value.LineNumber),
			Body:     gogithub.String(value.ReviewComment),
		})

		fmt.Println(value.File)
		fmt.Println(value.LineNumber)
		fmt.Println(value.ReviewComment)
	}

	var review = &gogithub.PullRequestReviewRequest{
		Body:     gogithub.String(prr.Comment),
		Event:    gogithub.String("COMMENT"),
		Comments: comments,
	}

	_, _, err = c.Client.PullRequests.CreateReview(c.ctx, prr.Owner, prr.Repo, prr.PrNumber, review)

	return
}

func (c *Client) GetPullRequestChanges(prr PullRequestReviewRequest) (contentPullRequest string, err error) {
	var (
		commits  []*gogithub.RepositoryCommit
		comments []*gogithub.PullRequestComment
	)

	if commits, _, err = c.Client.PullRequests.ListCommits(c.ctx, prr.Owner, prr.Repo, prr.PrNumber, nil); err != nil {
		return
	}

	if comments, _, err = c.Client.PullRequests.ListComments(c.ctx, prr.Owner, prr.Repo, prr.PrNumber, nil); err != nil {
		return
	}

	for _, commit := range commits {
		var commitInfos *gogithub.RepositoryCommit
		if commitInfos, _, err = c.Client.Repositories.GetCommit(c.ctx, prr.Owner, prr.Repo, *commit.SHA); err != nil {
			return
		}

		contentPullRequest += fmt.Sprintf("%s\nCommitID: %s\n", prompt.BEGIN_CONTENT, *commit.SHA)

		for _, file := range commitInfos.Files {

			if (*file.Additions + *file.Deletions + *file.Changes) > prr.MaxChangedLines {
				continue
			}

			contentPullRequest += fmt.Sprintf("Filename: %s\nAdditions: %d\nDeletions: %d\nChanges: %d\nStatus: %s\nContent: %s\n", *file.Filename, *file.Additions, *file.Deletions, *file.Changes, *file.Status, *file.Patch)
			for _, comment := range comments {
				var prComments string

				if *comment.Path == *file.Filename && *comment.CommitID == *commit.SHA {
					var position int
					if comment.Position != nil {
						position = *comment.Position
					}

					if comment.OriginalPosition != nil {
						position = *comment.OriginalPosition
					}

					prComments += fmt.Sprintf("Line: %d - User: %s - Comment: %s\n", position, *comment.User.Login, *comment.Body)
				}

				if prComments != "" {
					contentPullRequest += fmt.Sprintf("Comments: %s\n", prComments)
				}
			}
		}

		contentPullRequest += prompt.END_CONTENT + "\n"
	}

	return
}
