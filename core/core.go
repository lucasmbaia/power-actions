package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucasmbaia/power-actions/config"
	"github.com/lucasmbaia/power-actions/core/github"
	"github.com/lucasmbaia/power-actions/core/openai"
	"github.com/lucasmbaia/power-actions/core/prompt"
)

func Run(diffPath string) (err error) {
	var (
		content        = make(map[string][]byte)
		chatCompletion openai.ChatCompletionRequest
		chatContent    string
		chatResponse   openai.ChatCompletionResponse
		reviews        github.Reviews
	)

	if err = filepath.Walk(diffPath, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if !info.IsDir() && strings.Contains(info.Name(), ".diff") {
			var b []byte
			if b, e = os.ReadFile(path); e != nil {
				return e
			}

			var filename = strings.Replace(info.Name(), ".diff", "", -1)
			filename = strings.Replace(filename, "_", "/", -1)
			content[filename] = b
		}

		return nil
	}); err != nil {
		return
	}

	for fileName, fileContent := range content {
		chatContent += fmt.Sprintf("%s\nFile Name: %s\nContent: \n%s\n%s", prompt.BEGIN_CONTENT, fileName, fileContent, prompt.END_CONTENT)
	}

	chatCompletion = openai.ChatCompletionRequest{
		Model: config.EnvConfig.OpenaiModel,
		Messages: []openai.ChatMessages{{
			Role:    "system",
			Content: prompt.INITIAL_PROMPT,
		}, {
			Role:    "user",
			Content: chatContent,
		}},
	}

	if chatResponse, err = config.EnvSingletons.OpenaiClient.CreateChatCompletion(chatCompletion); err != nil {
		return
	}

	reviewStr := strings.Replace(chatResponse.Choices[0].Message.Content, "```json", "", 1)
	reviewStr = strings.Replace(reviewStr, "```", "", 1)

	if err = json.Unmarshal([]byte(reviewStr), &reviews); err != nil {
		return
	}

	fmt.Println("CARALHOOOOPOO")
	if len(reviews.Review) > 0 {
		err = config.EnvSingletons.GithubClient.PullRequestReview(github.PullRequestReviewRequest{
			Comment:  "While reviewing the proposed modifications, I identified some opportunities for improvement that can further enhance the quality of our project. I am available to discuss these suggestions and find the best solutions together.",
			Owner:    config.EnvConfig.GithubRepoOwner,
			Repo:     config.EnvConfig.GithubRepoName,
			PrNumber: config.EnvConfig.GithubPrNumber,
			Reviews:  reviews,
		})
	}

	fmt.Println(err)
	return
}
