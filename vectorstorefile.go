package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Custom error for unsupported file types
type UnsupportedFileTypeError struct {
	FileName string
}

func (e *UnsupportedFileTypeError) Error() string {
	return fmt.Sprintf("unsupported file type for file: %s", e.FileName)
}

// VectorStoreFile represents the response for attaching a file to a vector store
type VectorStoreFile struct {
	ID               string                  `json:"id"`
	Object           string                  `json:"object"`
	UsageBytes       int64                   `json:"usage_bytes"`
	CreatedAt        int64                   `json:"created_at"`
	VectorStoreID    string                  `json:"vector_store_id"`
	Status           string                  `json:"status"`
	LastError        *map[string]interface{} `json:"last_error,omitempty"`
	ChunkingStrategy map[string]interface{}  `json:"chunking_strategy,omitempty"`
}

// CreateVectorStoreFile attaches a file to a vector store
func CreateVectorStoreFile(vectorStoreID, fileID string, chunkingStrategy map[string]interface{}) (*VectorStoreFile, error) {
	// Prepare payload for attaching file
	payload := map[string]interface{}{
		"file_id":           fileID,
		"chunking_strategy": chunkingStrategy,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal vector store file payload: %w", err)
	}

	// Set up request to attach file to vector store
	url := fmt.Sprintf("https://api.openai.com/v1/vector_stores/%s/files", vectorStoreID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create vector store file request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vector store file request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		body, _ := io.ReadAll(resp.Body)

		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("vector store file creation failed with status %s: %s", resp.Status, string(body))
		}

		if errorResp.Error.Code == "unsupported_file" {
			return nil, &UnsupportedFileTypeError{FileName: fileID}
		}

		return nil, fmt.Errorf("vector store file creation failed: %s", errorResp.Error.Message)
	}

	// Decode response to get file attachment details
	var vectorStoreFileResp VectorStoreFile
	if err := json.NewDecoder(resp.Body).Decode(&vectorStoreFileResp); err != nil {
		return nil, fmt.Errorf("failed to decode vector store file response: %w", err)
	}

	fmt.Printf("File attached successfully to vector store with ID: %s\n", vectorStoreFileResp.ID)
	return &vectorStoreFileResp, nil
}

// VectorStoreFileListResponse represents the response from the list vector store files API
type VectorStoreFileListResponse struct {
	Data []VectorStoreFile `json:"data"`
}

// ListVectorStoreFiles lists files attached to a specific vector store
func ListVectorStoreFiles(vectorStoreID string) ([]VectorStoreFile, error) {
	// Build the request URL
	url := fmt.Sprintf("https://api.openai.com/v1/vector_stores/%s/files?limit=100", vectorStoreID)

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list vector store files request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list vector store files request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list vector store files failed with status %s: %s", resp.Status, string(body))
	}

	// Parse the response
	var vectorStoreFileList VectorStoreFileListResponse
	if err := json.NewDecoder(resp.Body).Decode(&vectorStoreFileList); err != nil {
		return nil, fmt.Errorf("failed to decode list vector store files response: %w", err)
	}

	return vectorStoreFileList.Data, nil
}

// RetrieveVectorStoreFile retrieves details of a specific file attached to a vector store
func RetrieveVectorStoreFile(vectorStoreID, fileID string) (*VectorStoreFile, error) {
	// Build the request URL
	url := fmt.Sprintf("https://api.openai.com/v1/vector_stores/%s/files/%s", vectorStoreID, fileID)

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create retrieve vector store file request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("retrieve vector store file request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("retrieve vector store file failed with status %s: %s", resp.Status, string(body))
	}

	// Parse the response
	var vectorStoreFile VectorStoreFile
	if err := json.NewDecoder(resp.Body).Decode(&vectorStoreFile); err != nil {
		return nil, fmt.Errorf("failed to decode retrieve vector store file response: %w", err)
	}

	return &vectorStoreFile, nil
}

// DeleteVectorStoreFile deletes a specific file from a vector store
func DeleteVectorStoreFile(vectorStoreID, fileID string) error {
	// Build the request URL
	url := fmt.Sprintf("https://api.openai.com/v1/vector_stores/%s/files/%s", vectorStoreID, fileID)

	// Create the request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete vector store file request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("delete vector store file request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("delete vector store file failed with status %s: %s", resp.Status, string(body))
	}

	fmt.Printf("File with ID %s deleted successfully from vector store %s\n", fileID, vectorStoreID)
	return nil
}
