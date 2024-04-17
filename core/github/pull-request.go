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
	File               string `json:"file"`
	LineNumber         int    `json:"lineNumber"`
	ReviewComment      string `json:"reviewComment"`
	SuggestionComments string `json:"suggestionComments"`
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
		var comment string = ""
		if len(value.ReviewComment) > 0 {
			comment += value.ReviewComment
		}
		if len(value.SuggestionComments) > 0 {
			comment += "\n```suggestion\n" + value.SuggestionComments + "\n```"
		}
		if comment != "" {
			comments = append(comments, &gogithub.DraftReviewComment{
				Path:     gogithub.String(value.File),
				Position: gogithub.Int(value.LineNumber),
				Body:     gogithub.String(comment),
			})
		}
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
		pullrequest *gogithub.PullRequest
		commits     []*gogithub.RepositoryCommit
		comments    []*gogithub.PullRequestComment
	)

	if pullrequest, _, err = c.Client.PullRequests.Get(c.ctx, prr.Owner, prr.Repo, prr.PrNumber); err != nil {
		return
	}

	if commits, _, err = c.Client.PullRequests.ListCommits(c.ctx, prr.Owner, prr.Repo, prr.PrNumber, nil); err != nil {
		return
	}

	if comments, _, err = c.Client.PullRequests.ListComments(c.ctx, prr.Owner, prr.Repo, prr.PrNumber, nil); err != nil {
		return
	}

	contentPullRequest += fmt.Sprintf(
		"Pull request title: %s\nPull request description:\n%s\n%s\n%s\n",
		pullrequest.GetTitle(),
		prompt.PR_BODY_START,
		pullrequest.GetBody(),
		prompt.PR_BODY_END,
	)

	for _, commit := range commits {
		var commitInfos *gogithub.RepositoryCommit
		if commitInfos, _, err = c.Client.Repositories.GetCommit(c.ctx, prr.Owner, prr.Repo, commit.GetSHA()); err != nil {
			return
		}

		contentPullRequest += fmt.Sprintf("%s\nCommitID: %s\n", prompt.BEGIN_CONTENT, commit.GetSHA())

		for _, file := range commitInfos.Files {
			if *file.Changes > prr.MaxChangedLines {
				continue
			}

			contentPullRequest += fmt.Sprintf(
				"\nPrevious filename: %s\nFilename: %s\nAdditions: %d\nDeletions: %d\nChanges: %d\nStatus: %s\nPatch:\n%s\n%s\n%s\n",
				file.GetPreviousFilename(),
				file.GetFilename(),
				file.GetAdditions(),
				file.GetDeletions(),
				file.GetChanges(),
				file.GetStatus(),
				prompt.PATCH_START,
				file.GetPatch(),
				prompt.PATCH_END,
			)
			for _, comment := range comments {
				var prComments string

				if comment.GetPath() == file.GetFilename() && comment.GetCommitID() == commit.GetSHA() {
					var position int
					if comment.Position != nil {
						position = comment.GetPosition()
					}

					if comment.OriginalPosition != nil {
						position = comment.GetOriginalPosition()
					}

					prComments += fmt.Sprintf(
						"\tComment:\n\t\tLine: %d\n\t\tUser: %s\n\t\tBody:\n%s\n%s\n%s\n",
						position,
						comment.GetUser().GetLogin(),
						prompt.COMMENT_BODY_START,
						comment.GetBody(),
						prompt.COMMENT_BODY_END,
					)
				}

				if prComments != "" {
					contentPullRequest += fmt.Sprintf("Comments:\n%s", prComments)
				}
			}
		}

		contentPullRequest += prompt.END_CONTENT + "\n"
	}

	return
}
