package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/lucasmbaia/power-actions/config"
	"github.com/lucasmbaia/power-actions/core/openai"
)

func Run() (err error) {
	var (
		chatCompletion     openai.ChatCompletionRequest
		chatResponse       openai.ChatCompletionResponse
		contentPullRequest string
		//reviews            github.Reviews
		//prr                github.PullRequestReviewRequest
		body []byte
	)

	/*prr = github.PullRequestReviewRequest{
		Owner:    config.EnvConfig.GithubRepoOwner,
		Repo:     config.EnvConfig.GithubRepoName,
		PrNumber: config.EnvConfig.GithubPrNumber,
	}

	if contentPullRequest, err = config.EnvSingletons.GithubClient.GetPullRequestChanges(prr); err != nil {
		return
	}*/

	fmt.Println(contentPullRequest)

	if body, err = os.ReadFile("./mock/pr-content"); err != nil {
		return
	}
	contentPullRequest = string(body)

	if body, err = os.ReadFile("./mock/prompt"); err != nil {
		return
	}
	initialPrompt := string(body)

	chatCompletion = openai.ChatCompletionRequest{
		Model: config.EnvConfig.OpenaiModel,
		Messages: []openai.ChatMessages{{
			Role:    "system",
			Content: initialPrompt,
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
	/*if err = json.Unmarshal([]byte(reviewStr), &reviews); err != nil {
		return
	}

	if len(reviews.Review) > 0 {
		prr.Comment = "While reviewing the proposed modifications, I identified some opportunities for improvement that can further enhance the quality of our project. I am available to discuss these suggestions and find the best solutions together."
		prr.Reviews = reviews
		err = config.EnvSingletons.GithubClient.PullRequestReview(prr)
	}*/

	//fmt.Println(err)
	return
}
