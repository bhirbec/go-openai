package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// ExpirationPolicy represents the expiration policy for a vector store
type ExpirationPolicy struct {
	Anchor string `json:"anchor"`
	Days   int    `json:"days"`
}

// ChunkingStrategy represents the chunking strategy options for file processing in the vector store
type ChunkingStrategy struct {
	Type               string `json:"type"`
	MaxChunkSizeTokens int    `json:"max_chunk_size_tokens,omitempty"`
	ChunkOverlapTokens int    `json:"chunk_overlap_tokens,omitempty"`
}

// CreateVectorStoreParams defines parameters for creating a vector store
type CreateVectorStoreParams struct {
	Name             string            `json:"name,omitempty"`
	FileIDs          []string          `json:"file_ids,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	ExpiresAfter     *ExpirationPolicy `json:"expires_after,omitempty"`
	ChunkingStrategy *ChunkingStrategy `json:"chunking_strategy,omitempty"`
}

// VectorStore represents the response for retrieving or creating a vector store
type VectorStore struct {
	ID           string            `json:"id"`
	Object       string            `json:"object"`
	CreatedAt    int64             `json:"created_at"`
	Name         string            `json:"name"`
	UsageBytes   int64             `json:"usage_bytes"`
	Status       string            `json:"status"`
	ExpiresAfter *ExpirationPolicy `json:"expires_after,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	FileCounts   map[string]int    `json:"file_counts,omitempty"`
	ExpiresAt    *int64            `json:"expires_at,omitempty"`
	LastActiveAt *int64            `json:"last_active_at,omitempty"`
}

// CreateVectorStore creates a new vector store in OpenAI’s storage
func CreateVectorStore(params *CreateVectorStoreParams) (*VectorStore, error) {
	// Marshal the parameters to JSON
	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal vector store payload: %w", err)
	}

	// Send request to vector store API
	url := "https://api.openai.com/v1/vector_stores"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create vector store request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vector store request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("vector store creation failed with status %s: %s", resp.Status, string(body))
	}

	// Decode response to get vector store information
	var vectorStoreResp VectorStore
	if err := json.NewDecoder(resp.Body).Decode(&vectorStoreResp); err != nil {
		return nil, fmt.Errorf("failed to decode vector store response: %w", err)
	}

	fmt.Printf("Vector store created successfully with ID: %s\n", vectorStoreResp.ID)
	return &vectorStoreResp, nil
}

// VectorStoreListResponse represents the response from the list vector stores API
type VectorStoreListResponse struct {
	Data []VectorStore `json:"data"`
}

// ListVectorStores lists vector stores with optional parameters for pagination and sorting
func ListVectorStores(limit int, order, after, before string) ([]VectorStore, error) {
	// Prepare query parameters
	params := url.Values{}
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}
	if order != "" {
		params.Add("order", order)
	}
	if after != "" {
		params.Add("after", after)
	}
	if before != "" {
		params.Add("before", before)
	}

	// Build the request URL
	baseURL := "https://api.openai.com/v1/vector_stores"
	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Create the request
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list vector stores request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey) // Authorization header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list vector stores request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("list vector stores failed with status %s: %s", resp.Status, string(body))
	}

	// Parse the response
	var vectorStoreList VectorStoreListResponse
	if err := json.NewDecoder(resp.Body).Decode(&vectorStoreList); err != nil {
		return nil, fmt.Errorf("failed to decode list vector stores response: %w", err)
	}

	return vectorStoreList.Data, nil
}

// RetrieveVectorStore retrieves details of a specific vector store
func RetrieveVectorStore(vectorStoreID string) (*VectorStore, error) {
	// Build the request URL
	url := fmt.Sprintf("https://api.openai.com/v1/vector_stores/%s", vectorStoreID)

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create retrieve vector store request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("retrieve vector store request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("retrieve vector store failed with status %s: %s", resp.Status, string(body))
	}

	// Parse the response
	var vectorStore VectorStore
	if err := json.NewDecoder(resp.Body).Decode(&vectorStore); err != nil {
		return nil, fmt.Errorf("failed to decode retrieve vector store response: %w", err)
	}

	return &vectorStore, nil
}

// DeleteVectorStore deletes a specific vector store
func DeleteVectorStore(vectorStoreID string) error {
	// Build the request URL
	url := fmt.Sprintf("https://api.openai.com/v1/vector_stores/%s", vectorStoreID)

	// Create the request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete vector store request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("delete vector store request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("delete vector store failed with status %s: %s", resp.Status, string(body))
	}

	fmt.Printf("Vector store with ID %s deleted successfully\n", vectorStoreID)
	return nil
}
