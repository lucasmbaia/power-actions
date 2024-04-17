package openai

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lucasmbaia/power-actions/request"
)

type ErrorResponse struct {
	Error ErrorMessage `json:"error"`
}

type ErrorMessage struct {
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
}

type ChatCompletionRequest struct {
	Model       string         `json:"model"`
	Messages    []ChatMessages `json:"messages"`
	Temperature float32        `json:"temperature"`
}

type ChatMessages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	ID                string                 `json:"id"`
	Choices           []ChatCompletionChoice `json:"choices"`
	SystemFingerprint string                 `json:"system_fingerprint"`
}

type ChatCompletionChoice struct {
	Index        int          `json:"index"`
	Message      ChatMessages `json:"message"`
	FinishReason string       `json:"finish_reason"`
}

type Speech struct {
	Model string `json:"model"`
	Input string `json:"input"`
	Voice string `json:"voice"`
}

func (c *Client) CreateChatCompletion(chatCompletion ChatCompletionRequest) (response ChatCompletionResponse, err error) {
	var (
		httpResponse request.Response
	)

	if httpResponse, err = c.httpClient.Request(request.POST, fmt.Sprintf("%s/v1/chat/completions", c.openAiUrl), request.Options{
		Body: chatCompletion,
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", c.key),
			"Content-Type":  "application/json",
		},
	}); err != nil {
		return
	}

	if httpResponse.Code == http.StatusOK {
		err = json.Unmarshal(httpResponse.Body, &response)
	} else {
		var errorResponse ErrorResponse
		if err = json.Unmarshal(httpResponse.Body, &errorResponse); err != nil {
			err = fmt.Errorf(string(httpResponse.Body))
			return
		}

		err = fmt.Errorf("message: %s, type: %s", errorResponse.Error.Message, errorResponse.Error.Type)
	}

	return
}
