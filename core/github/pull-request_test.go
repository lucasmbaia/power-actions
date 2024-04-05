package github

import (
	"fmt"
	"os"
	"testing"
)

func Test_GetPullRequestChanges(t *testing.T) {
	var (
		c       Client
		err     error
		content string
	)

	c = NewClient(os.Getenv("GITHUB_TOKEN"))

	if content, err = c.GetPullRequestChanges(PullRequestReviewRequest{
		Owner:    os.Getenv("GITHUB_OWNER"),
		Repo:     os.Getenv("GITHUB_REPO"),
		PrNumber: 6,
	}); err != nil {
		t.Fatal(err)
	}

	fmt.Println(content)
}
