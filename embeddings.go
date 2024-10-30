package openai

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	openai "github.com/sashabaranov/go-openai"
)

// / createEmbedding reads the content of a file, uploads it to OpenAI embeddings, and returns a generated ID for the embedding
func CreateEmbedding(filePath string) (string, error) {
	// Read the file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Initialize OpenAI client
	client := openai.NewClient(openaiAPIKey)
	ctx := context.Background()

	// Create embedding request
	embeddingReq := openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2,     // Model for generating embeddings
		Input: []string{string(content)}, // Convert content to string and pass as array
	}

	// Generate embedding using OpenAI client
	resp, err := client.CreateEmbeddings(ctx, embeddingReq)
	if err != nil {
		return "", fmt.Errorf("error creating embedding: %w", err)
	}

	if len(resp.Data) == 0 || len(resp.Data[0].Embedding) == 0 {
		return "", fmt.Errorf("no embedding data returned")
	}

	// Generate a unique ID for this embedding using SHA-1 hash of the content
	hash := sha1.New()
	hash.Write(content)
	embeddingID := hex.EncodeToString(hash.Sum(nil))

	// Optionally, store `resp.Data[0].Embedding` here if needed
	// You could also consider persisting this in a vector database for further use
	return embeddingID, nil
}

// EmbeddingResponse represents the response from the embeddings API
type EmbeddingResponse struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// CreateVectorForFile generates an embedding for the file content and returns a unique ID based on the embedding
func CreateVectorForFile(filePath string) (string, error) {
	// Read the file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Prepare payload for embedding request
	payload := map[string]interface{}{
		"input": string(content),          // Convert content to string for embedding input
		"model": "text-embedding-ada-002", // Embedding model
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal embedding payload: %w", err)
	}

	// Send request to embeddings API
	url := "https://api.openai.com/v1/embeddings"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create embedding request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("embedding creation failed with status %s: %s", resp.Status, string(body))
	}

	// Decode response to get embedding data
	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return "", fmt.Errorf("failed to decode embedding response: %w", err)
	}

	if len(embeddingResp.Embedding) == 0 {
		return "", fmt.Errorf("no embedding data returned for file '%s'", filePath)
	}

	// Generate a unique ID for this embedding using SHA-1 hash of the content
	hash := sha1.New()
	hash.Write([]byte(content))
	embeddingID := hex.EncodeToString(hash.Sum(nil))

	fmt.Printf("Embedding created successfully for %s with ID: %s\n", filePath, embeddingID)
	return embeddingID, nil
}
