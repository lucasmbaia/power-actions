package services

// PRInfo represents the data of a single pull request
type PRInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
