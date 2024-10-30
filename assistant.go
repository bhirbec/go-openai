package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Assistant represents an individual assistant's information
type Assistant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Model       string `json:"model"`
	CreatedAt   int64  `json:"created_at"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// ListAssistants retrieves a list of all assistants
func ListAssistants() ([]Assistant, error) {
	url := "https://api.openai.com/v1/assistants"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("retrieving assistants failed with status %s: %s", resp.Status, string(body))
	}

	// Parse the response
	var response struct {
		Data []Assistant `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data, nil
}

type CreateAssistantParams struct {
	Name           string                 `json:"name,omitempty"`
	Description    string                 `json:"description,omitempty"`
	Model          string                 `json:"model"`
	Instructions   string                 `json:"instructions,omitempty"`
	Tools          []Tool                 `json:"tools,omitempty"`
	ToolResources  map[string]interface{} `json:"tool_resources,omitempty"`
	Temperature    *float64               `json:"temperature,omitempty"`
	TopP           *float64               `json:"top_p,omitempty"`
	ResponseFormat interface{}            `json:"response_format,omitempty"`
	Metadata       map[string]string      `json:"metadata,omitempty"`
}

type Tool struct {
	Type            string                 `json:"type"`
	FileSearch      *FileSearchConfig      `json:"file_search,omitempty"`
	CodeInterpreter *CodeInterpreterConfig `json:"code_interpreter,omitempty"`
}

type FileSearchConfig struct {
	MaxNumResults  int             `json:"max_num_results,omitempty"`
	RankingOptions *RankingOptions `json:"ranking_options,omitempty"`
}

type RankingOptions struct {
	Ranker         string  `json:"ranker,omitempty"`
	ScoreThreshold float64 `json:"score_threshold"`
}

type CodeInterpreterConfig struct {
	FileIDs []string `json:"file_ids,omitempty"`
}

// CreateAssistant creates an assistant with the provided configuration
func CreateAssistant(params *CreateAssistantParams) (string, error) {
	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("failed to marshal assistant payload: %w", err)
	}

	url := "https://api.openai.com/v1/assistants"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create assistant request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("assistant request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("assistant creation failed with status %s: %s", resp.Status, string(body))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode assistant response: %w", err)
	}
	assistantID, _ := response["id"].(string)
	fmt.Printf("Assistant created successfully with ID: %s\n", assistantID)
	return assistantID, nil
}

// DeleteAssistant deletes an assistant by its ID
func DeleteAssistant(assistantID string) error {
	url := fmt.Sprintf("https://api.openai.com/v1/assistants/%s", assistantID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2") // Extra header for beta features

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("assistant deletion failed with status %s: %s", resp.Status, string(body))
	}

	fmt.Printf("Assistant with ID %s deleted successfully.\n", assistantID)
	return nil
}
