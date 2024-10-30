package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Thread represents the response from creating or retrieving a thread
type Thread struct {
	ID            string                 `json:"id"`
	Object        string                 `json:"object"`
	CreatedAt     int64                  `json:"created_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ToolResources map[string]interface{} `json:"tool_resources,omitempty"`
}

// ThreadMessage represents the message structure in a thread
type ThreadMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CreateThreadParams defines the parameters for creating a thread
type CreateThreadParams struct {
	Messages       []ThreadMessage          `json:"messages,omitempty"`
	ToolResources  map[string]interface{}   `json:"tool_resources,omitempty"`
	VectorStoreIDs []string                 `json:"vector_store_ids,omitempty"`
	VectorStores   []map[string]interface{} `json:"vector_stores,omitempty"`
}

// CreateThread creates a new thread with the specified parameters
func CreateThread(params *CreateThreadParams) (*Thread, error) {
	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal thread payload: %w", err)
	}

	url := "https://api.openai.com/v1/threads"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create thread request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("thread request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("thread creation failed with status %s: %s", resp.Status, string(body))
	}

	var response Thread
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode thread response: %w", err)
	}

	fmt.Printf("Thread created successfully with ID: %s\n", response.ID)
	return &response, nil
}
