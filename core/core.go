package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lucasmbaia/power-actions/config"
	"github.com/lucasmbaia/power-actions/core/github"
	"github.com/lucasmbaia/power-actions/core/openai"
	"github.com/lucasmbaia/power-actions/core/prompt"
)

func Run() (err error) {
	var (
		chatCompletion     openai.ChatCompletionRequest
		chatResponse       openai.ChatCompletionResponse
		contentPullRequest string
		reviews            github.Reviews
		prr                github.PullRequestReviewRequest
	)

	prr = github.PullRequestReviewRequest{
		Owner:    config.EnvConfig.GithubRepoOwner,
		Repo:     config.EnvConfig.GithubRepoName,
		PrNumber: config.EnvConfig.GithubPrNumber,
	}

	if contentPullRequest, err = config.EnvSingletons.GithubClient.GetPullRequestChanges(prr); err != nil {
		return
	}

	chatCompletion = openai.ChatCompletionRequest{
		Model: config.EnvConfig.OpenaiModel,
		Messages: []openai.ChatMessages{{
			Role:    "system",
			Content: prompt.INITIAL_PROMPT,
		}, {
			Role:    "user",
			Content: contentPullRequest,
		}},
	}

	if chatResponse, err = config.EnvSingletons.OpenaiClient.CreateChatCompletion(chatCompletion); err != nil {
		return
	}

	reviewStr := strings.Replace(chatResponse.Choices[0].Message.Content, "```json", "", 1)
	reviewStr = strings.Replace(reviewStr, "```", "", 1)

	fmt.Println(reviewStr)
	if err = json.Unmarshal([]byte(reviewStr), &reviews); err != nil {
		return
	}

	if len(reviews.Review) > 0 {
		prr.Comment = "While reviewing the proposed modifications, I identified some opportunities for improvement that can further enhance the quality of our project. I am available to discuss these suggestions and find the best solutions together."
		prr.Reviews = reviews
		err = config.EnvSingletons.GithubClient.PullRequestReview(prr)
	}

	fmt.Println(err)
	return
}
