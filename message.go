package openai

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Message represents a single message in a thread
type Message struct {
	ID          string                 `json:"id"`
	Object      string                 `json:"object"`
	CreatedAt   int64                  `json:"created_at"`
	AssistantID *string                `json:"assistant_id,omitempty"`
	ThreadID    string                 `json:"thread_id"`
	RunID       *string                `json:"run_id,omitempty"`
	Role        string                 `json:"role"`
	Content     []MessageContent       `json:"content"`
	Attachments []interface{}          `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MessageContent represents the content structure within a message
type MessageContent struct {
	Type string      `json:"type"`
	Text ContentText `json:"text"`
}

// ContentText holds the textual content of a message
type ContentText struct {
	Value       string        `json:"value"`
	Annotations []interface{} `json:"annotations,omitempty"`
}

// ListMessages retrieves a list of messages from a given thread with optional query parameters
func ListMessages(threadID string, limit int, order, after, before, runID string) ([]Message, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages", threadID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to list messages: %w", err)
	}

	// Set query parameters based on provided values
	q := req.URL.Query()
	if limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", limit))
	}
	if order != "" {
		q.Add("order", order)
	}
	if after != "" {
		q.Add("after", after)
	}
	if before != "" {
		q.Add("before", before)
	}
	if runID != "" {
		q.Add("run_id", runID)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to list messages failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list messages with status %s: %s", resp.Status, string(body))
	}

	var result struct {
		Data []Message `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode messages response: %w", err)
	}

	return result.Data, nil
}
