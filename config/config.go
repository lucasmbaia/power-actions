package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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

	MaxChangedLines int
	OpenaiModel     string
}

func LoadSingletons() {
	var err error

	if EnvSingletons.OpenaiClient, err = openai.NewClient(openai.Config{
		Key: os.Getenv("POWERPR_OPENAI_KEY"),
	}); err != nil {
		log.Fatalf("Error to initiate openai client: %s", err.Error())
	}

	EnvSingletons.GithubClient = github.NewClient(os.Getenv("GITHUB_TOKEN"))

	EnvConfig.GithubRepoOwner = os.Getenv("GITHUB_OWNER")
	EnvConfig.GithubRepoName = strings.Replace(os.Getenv("GITHUB_REPO"), fmt.Sprintf("%s/", EnvConfig.GithubRepoOwner), "", -1)
	EnvConfig.OpenaiModel = os.Getenv("OPENAI_MODEL")
	EnvConfig.MaxChangedLines = 500

	if EnvConfig.MaxChangedLines, err = getUnsignedIntEnv("MAX_CHANGED_LINES", 500); EnvConfig.MaxChangedLines <= 0 || err != nil {
		if EnvConfig.MaxChangedLines <= 0 {
			log.Fatalf("MAX_CHANGED_LINES need to be a positive integer")
		}
		log.Fatal(err)
	}

	if EnvConfig.GithubPrNumber, err = strconv.Atoi(os.Getenv("GITHUB_PR_NUMBER")); err != nil {
		log.Fatal(err)
	}
}

func getUnsignedIntEnv(varName string, defaultValue int) (int, error) {
	// Retrieve the value of the environment variable
	valueStr := os.Getenv(varName)
	if valueStr == "" {
		return defaultValue, nil
	}

	// Attempt to convert the string to an integer
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("environment variable %s is not a valid integer: %v", varName, err)
	}

	return value, nil
}
