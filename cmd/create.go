package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// CommitData represents the data of a single git commit
type CommitData struct {
	ID      string
	Message string
	Files   []string
	Diffs   map[string]string // Maps file names to their diffs
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Automate creation of a pull request on GitHub for current branch changes",
	Long:  "This command streamlines the process of creating a pull request by automatically capturing and uploading changes from the current development branch to GitHub. It evaluates the latest commits, prepares them for review, and initiates a pull request. This tool is designed to enhance workflow efficiency by ensuring that branch updates are promptly and accurately proposed for integration and review.",
	Run: func(cmd *cobra.Command, args []string) {

		commits, err := getCommits()
		if err != nil {
			fmt.Printf("Error retrieving commits: %v\n", err)
			return
		}

		// Process commits in chunks to manage token limits
		chunkSize := 5 // Adjust chunk size based on average token count per commit
		for i := 0; i < len(commits); i += chunkSize {
			end := i + chunkSize
			if end > len(commits) {
				end = len(commits)
			}
			processCommitsChunk(commits[i:end])
		}
	},
}

// getCommits retrieves the local git branch commits and structures them into a slice of CommitData
func getCommits() ([]CommitData, error) {
	output, err := exec.Command("git", "log", "--pretty=%H %s").Output()
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

// Process a chunk of commits, sending each chunk to the OpenAI API
func processCommitsChunk(commits []CommitData) {
	client := openai.NewClient(key) // Ensure you have a valid OpenAI client setup
	prompt := formatPromptForPR(commits)
	resp, err := client.CreateChatCompletion(
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
		return
	}

	fmt.Println("PR Title and Description for chunk:")
	fmt.Println(resp.Choices[0].Message.Content)
}

// Format the commits into a prompt for generating a PR title and description
func formatPromptForPR(commits []CommitData) string {
	prompt := "Based on the following commit details, generate a GitHub PR title and a detailed description in Markdown format:\n\n"
	for _, commit := range commits {
		prompt += fmt.Sprintf("### Commit %s\n", commit.ID)
		prompt += fmt.Sprintf("**Message:** %s\n", commit.Message)
		prompt += "**Files changed:**\n"
		for _, file := range commit.Files {
			prompt += fmt.Sprintf("- %s\n", file)
			if diff, ok := commit.Diffs[file]; ok && len(diff) > 0 {
				diffLines := strings.Split(diff, "\n")
				// Only include a part of the diff if it's too long
				maxDiffLines := 10
				if len(diffLines) > maxDiffLines {
					diff = strings.Join(diffLines[:maxDiffLines], "\n") + "\n... (truncated)"
				}
				prompt += fmt.Sprintf("```diff\n%s\n```\n", diff)
			}
		}
		prompt += "\n"
	}
	return prompt
}

// Key for OpenAI API
var key string

func init() {
	rootCmd.AddCommand(createCmd)

	// Viper setup for environment variable
	viper.AutomaticEnv()          // Automatically read from environment variables
	viper.SetEnvPrefix("POWERPR") // Set a prefix for environment variables to avoid conflicts
	viper.BindEnv("KEY")          // Bind the environment variable KEY

	// Optionally set a default in case the env isn't set
	//viper.SetDefault("KEY", "your-default-key")

	// Binding the API key flag with viper
	createCmd.Flags().StringVarP(&key, "key", "k", "", "OpenAI API key")
	viper.BindPFlag("KEY", createCmd.Flags().Lookup("key"))

}
