package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	gogithub "github.com/google/go-github/v39/github"
	"github.com/lucasmbaia/power-actions/services"
	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CommitData represents the data of a single git commit
type CommitData struct {
	ID      string
	Message string
	Files   []string
	Diffs   map[string]string // Maps file names to their diffs
}

var logger *zap.SugaredLogger
var openAIClient *openai.Client

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Automate creation of a pull request on GitHub for current branch changes",
	Run: func(cmd *cobra.Command, args []string) {

		atom := zap.NewAtomicLevel()
		var level zapcore.Level
		err := level.UnmarshalText([]byte(logLevel))
		if err != nil {
			fmt.Printf("Error setting log level: %s\n", err.Error())
			return
		}
		atom.SetLevel(level)

		cfg, _ := configureLogger()
		logger = zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(*cfg),
			zapcore.Lock(os.Stdout),
			atom,
		)).Sugar()
		defer logger.Sync()

		// Initialize OpenAI client
		openAIClient = openai.NewClient(viper.GetString("OPENAI_KEY"))
		gitHubClient := services.NewGitHubClient(viper.GetString("GITHUB_KEY"))

		gitRepoInfo, err := services.GetGitRepoInfo()
		if err != nil {
			logger.Error("Error getting git repository info", zap.Error(err))
			return
		}

		commits, err := getCommits(gitRepoInfo.CurrentBranch, gitRepoInfo.PrincipalBranch)
		if err != nil {
			logger.Error("Error retrieving commits", zap.Error(err))
			return
		}

		summaries := make([]string, 0, len(commits))
		for _, commit := range commits {
			summary := processSingleCommit(commit)
			if summary != "" {
				summaries = append(summaries, summary)
			}
		}

		if len(summaries) == 0 {
			logger.Warn("No valid data to generate PR title and description.")
			return
		}

		finalPrompt := createFinalPrompt(summaries)
		prInfo, err := generatePRTitleAndDescription(finalPrompt)

		if err != nil {
			logger.Error("Error generating PR title and description", err)
			return
		}

		// Create a new pull request
		newPullRequest := &gogithub.NewPullRequest{
			Title: gogithub.String(prInfo.Title),
			Body:  gogithub.String(prInfo.Description),
			Base:  gogithub.String(gitRepoInfo.PrincipalBranch),
			Head:  gogithub.String(gitRepoInfo.CurrentBranch),
		}

		// Create a pull request
		pr, resp, err := gitHubClient.Client.PullRequests.Create(context.Background(), gitRepoInfo.RepositoryOwner, gitRepoInfo.RepositoryName, newPullRequest)

		if err != nil {
			logger.Error("Error creating pull request", zap.Error(err))
			return
		}

		if resp.StatusCode != 201 {
			logger.Error("Error creating pull request", zap.Any("response", resp))
			return
		}

		fmt.Printf("Pull request created successfully: %s\n", pr.GetHTMLURL())
	},
}

// processSingleCommit sends a single commit to the OpenAI API and returns a generated summary
func processSingleCommit(commit CommitData) string {
	fmt.Printf("Processing commit: %s\n", commit.ID)

	prompt := formatPromptForCommit(commit)
	resp, err := openAIClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		logger.Error("ChatCompletion error", zap.Error(err))
		return ""
	}
	return resp.Choices[0].Message.Content
}

// createFinalPrompt aggregates all summaries into a final prompt for the PR title and description
func createFinalPrompt(summaries []string) string {
	var builder strings.Builder
	builder.WriteString("Generate a single PR title and a rich and well descriptive and detailed description in markdown format based on the following summaries of changes and return it as an inline json with one field for the title and other field for a rich and well descriptive and detailed description. This description value should be scaped to avoid error to parse the json and should be in a markdown format. Use the best practices to create a bery good human readable description. Do it like a senior software engineer:\n\n")
	for _, summary := range summaries {
		builder.WriteString(summary + "\n")
	}
	return builder.String()
}

// generatePRTitleAndDescription sends the final prompt to generate the PR title and description
func generatePRTitleAndDescription(prompt string) (services.PRInfo, error) {
	resp, err := openAIClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return services.PRInfo{}, err
	}

	// JSON string
	jsonData := resp.Choices[0].Message.Content
	jsonData, _ = strings.CutPrefix(jsonData, "```json\n")
	jsonData, _ = strings.CutSuffix(jsonData, "\n```")
	jsonData = strings.ReplaceAll(jsonData, `\"`, `"`)

	// Variable to hold the unmarshalled data
	var prInfo services.PRInfo

	// Unmarshal the JSON into the struct
	err = json.Unmarshal([]byte(jsonData), &prInfo)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return services.PRInfo{}, err
	}

	return prInfo, nil
}

// formatPromptForCommit creates a prompt for a single commit
func formatPromptForCommit(commit CommitData) string {
	diffs := make([]string, 0, len(commit.Diffs))
	for _, diff := range commit.Diffs {
		diffs = append(diffs, diff)
	}
	return fmt.Sprintf("Analyze the following commit and summarize its impact and changes:\nCommit ID: %s\nMessage: %s\nFiles Changed: %s\nDiffs:\n%s",
		commit.ID, commit.Message, strings.Join(commit.Files, ", "), strings.Join(diffs, "\n"))
}

func getCommits(currentBranch string, mainBranch string) ([]CommitData, error) {
	// Retrieve commits that are only in the current branch compared to master
	output, err := exec.Command("git", "log", fmt.Sprintf("%s..%s", mainBranch, currentBranch), "--pretty=%H %s").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve git branch commits: %w", err)
	}

	commitLines := strings.Split(string(output), "\n")
	var commits []CommitData

	for _, line := range commitLines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue // Skip invalid lines
		}
		commitID := parts[0]
		message := parts[1]

		fileOutput, err := exec.Command("git", "show", "--name-only", "--format=", commitID).Output()
		if err != nil {
			continue // Skip commits that fail to retrieve files
		}
		files := strings.Split(strings.TrimSpace(string(fileOutput)), "\n")

		diffs := make(map[string]string)
		for _, file := range files {
			diffOutput, err := exec.Command("git", "diff", commitID+"^!", "--", file).Output()
			if err != nil {
				continue // Skip files that fail to retrieve diffs
			}
			diffs[file] = string(diffOutput)
		}

		commits = append(commits, CommitData{
			ID:      commitID,
			Message: message,
			Files:   files,
			Diffs:   diffs,
		})
	}
	return commits, nil
}

var openAIKey string
var gitHubKey string
var logLevel string

func init() {
	rootCmd.AddCommand(createCmd)

	// Viper setup for environment variable
	viper.AutomaticEnv()          // Automatically read from environment variables
	viper.SetEnvPrefix("POWERPR") // Set a prefix for environment variables to avoid conflicts
	viper.BindEnv("OPENAI_KEY")   // Bind the environment variable to a key
	viper.BindEnv("GITHUB_KEY")   // Bind the environment variable to a key
	viper.BindEnv("LOG_LEVEL")    // Bind the environment variable to a key

	// Optionally set a default in case the env isn't set
	//viper.SetDefault("KEY", "your-default-key")

	// Binding flags to viper
	createCmd.Flags().StringVarP(&openAIKey, "openAIKey", "o", "", "OpenAI API key")
	viper.BindPFlag("OPENAI_KEY", createCmd.Flags().Lookup("openAIKey"))
	createCmd.Flags().StringVarP(&gitHubKey, "gitHubKey", "g", "", "GitHub key")
	viper.BindPFlag("GITHUB_KEY", createCmd.Flags().Lookup("gitHubKey"))

	createCmd.Flags().StringVarP(&logLevel, "logLevel", "l", "debug", "Log level (debug, info, warn, error, dpanic, panic, fatal)")
	viper.BindPFlag("LOG_LEVEL", createCmd.Flags().Lookup("logLevel"))

}

func configureLogger() (*zapcore.EncoderConfig, error) {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "timestamp"
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	return &cfg, nil
}
