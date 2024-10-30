package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CreateRunParams struct {
	AssistantID            string                   `json:"assistant_id"`
	Model                  *string                  `json:"model,omitempty"`
	Instructions           *string                  `json:"instructions,omitempty"`
	AdditionalInstructions *string                  `json:"additional_instructions,omitempty"`
	AdditionalMessages     []ThreadMessage          `json:"additional_messages,omitempty"`
	Tools                  []map[string]interface{} `json:"tools,omitempty"`
	Metadata               map[string]string        `json:"metadata,omitempty"`
	Temperature            *float64                 `json:"temperature,omitempty"`
	TopP                   *float64                 `json:"top_p,omitempty"`
	Stream                 *bool                    `json:"stream,omitempty"`
	MaxPromptTokens        *int                     `json:"max_prompt_tokens,omitempty"`
	MaxCompletionTokens    *int                     `json:"max_completion_tokens,omitempty"`
	TruncationStrategy     *map[string]interface{}  `json:"truncation_strategy,omitempty"`
	ToolChoice             *map[string]interface{}  `json:"tool_choice,omitempty"`
	ParallelToolCalls      *bool                    `json:"parallel_tool_calls,omitempty"`
	ResponseFormat         *interface{}             `json:"response_format,omitempty"`
}

type Run struct {
	ID           string  `json:"id"`
	Object       string  `json:"object"`
	CreatedAt    int64   `json:"created_at"`
	AssistantID  string  `json:"assistant_id"`
	ThreadID     string  `json:"thread_id"`
	Status       string  `json:"status"`
	StartedAt    *int64  `json:"started_at,omitempty"`
	ExpiresAt    *int64  `json:"expires_at,omitempty"`
	CancelledAt  *int64  `json:"cancelled_at,omitempty"`
	FailedAt     *int64  `json:"failed_at,omitempty"`
	CompletedAt  *int64  `json:"completed_at,omitempty"`
	LastError    *string `json:"last_error,omitempty"`
	Model        string  `json:"model"`
	Instructions *string `json:"instructions,omitempty"`
	// Tools             []map[string]string    `json:"tools,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	IncompleteDetails *string                `json:"incomplete_details,omitempty"`
	Usage             struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Temperature         *float64               `json:"temperature,omitempty"`
	TopP                *float64               `json:"top_p,omitempty"`
	MaxPromptTokens     *int                   `json:"max_prompt_tokens,omitempty"`
	MaxCompletionTokens *int                   `json:"max_completion_tokens,omitempty"`
	TruncationStrategy  map[string]interface{} `json:"truncation_strategy,omitempty"`
	ResponseFormat      string                 `json:"response_format"`
	ToolChoice          string                 `json:"tool_choice"`
	ParallelToolCalls   *bool                  `json:"parallel_tool_calls,omitempty"`
}

// CreateRun creates a run in a specified thread using the given parameters
func CreateRun(threadID string, params *CreateRunParams, include []string) (*Run, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/runs", threadID)
	if len(include) > 0 {
		queryParams := "?include=" + include[0]
		for _, field := range include[1:] {
			queryParams += "&include=" + field
		}
		url += queryParams
	}

	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal run payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create run request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("run request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("run creation failed with status %s: %s", resp.Status, string(body))
	}

	// Decode the JSON response
	var response Run
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode run response: %w", err)
	}

	fmt.Printf("Run created successfully with ID: %s, Status: %s\n", response.ID, response.Status)
	return &response, nil
}

// RetrieveRun retrieves the status and details of a specific run within a thread
func RetrieveRun(threadID, runID string) (*Run, error) {
	// Construct the request URL
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/runs/%s", threadID, runID)

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get run request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("run retrieval request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("run retrieval failed with status %s: %s", resp.Status, string(body))
	}

	// Decode the JSON response into a Run struct
	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("failed to decode run response: %w", err)
	}

	return &run, nil
}
