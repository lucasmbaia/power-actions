package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/lucasmbaia/power-actions/core/github"
	"github.com/lucasmbaia/power-actions/core/openai"
)

var (
	EnvSingletons Singletons
	EnvConfig     Config
)

type Singletons struct {
	OpenaiClient openai.Client
	GithubClient github.Client
}

type Config struct {
	GithubRepoOwner string
	GithubRepoName  string
	GithubPrNumber  int

	OpenaiModel string
}

func LoadSingletons() {
	var err error

	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Fatalf("It's mandatory set a github token in GITHUB_TOKEN environment variable")
	}

	if EnvSingletons.OpenaiClient, err = openai.NewClient(openai.Config{
		Key: os.Getenv("OPENAI_TOKEN"),
	}); err != nil {
		log.Fatalf("Error to initiate openai client: %s", err.Error())
	}

	EnvSingletons.GithubClient = github.NewClient(os.Getenv("GITHUB_TOKEN"))

	EnvConfig.GithubRepoOwner = os.Getenv("GITHUB_OWNER")
	EnvConfig.GithubRepoName = os.Getenv("GITHUB_REPO")
	EnvConfig.OpenaiModel = os.Getenv("OPENAI_MODEL")
	if EnvConfig.GithubPrNumber, err = strconv.Atoi(os.Getenv("GITHUB_PR_NUMBER")); err != nil {
		log.Fatal(err)
	}

	fmt.Println("********* CONFIG ***********")
	fmt.Println(EnvConfig)
}
