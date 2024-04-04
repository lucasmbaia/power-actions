package openai

import (
	"fmt"

	"github.com/lucasmbaia/power-actions/request"
)

type Client struct {
	key        string
	httpClient *request.Client
	openAiUrl  string
}

type Config struct {
	Key       string
	OpenAIUrl string
}

func NewClient(cfg Config) (c Client, err error) {
	c.openAiUrl = "https://api.openai.com"
	if cfg.Key == "" {
		err = fmt.Errorf("you must inform openai key")
		return
	}

	if cfg.OpenAIUrl != "" {
		c.openAiUrl = cfg.OpenAIUrl
	}

	if c.httpClient, err = request.NewClient(request.ClientConfiguration{}); err != nil {
		return
	}
	c.key = cfg.Key

	return
}
