package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

// File holds response data for a file upload
type File struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
	Bytes     int64  `json:"bytes"`
	FileName  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

// UploadFile uploads a file as a multi-part form data to ChatGPT
func UploadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Prepare a multi-part form
	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	// Add the "purpose" field required by the API
	purposeWriter, err := multiWriter.CreateFormField("purpose")
	if err != nil {
		return "", fmt.Errorf("failed to add purpose field: %w", err)
	}
	_, err = purposeWriter.Write([]byte("user_data")) // Replace "user_data" with the actual purpose if different
	if err != nil {
		return "", fmt.Errorf("failed to write purpose to form: %w", err)
	}

	// Add the file field
	fileWriter, err := multiWriter.CreateFormFile("file", filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	// Copy the actual file content into fileWriter
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return "", fmt.Errorf("failed to write file content to form: %w", err)
	}

	// Close the multi-part writer to set the correct boundary
	multiWriter.Close()

	// Create the request
	url := "https://api.openai.com/v1/files" // Replace with the actual endpoint
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %s: %s", resp.Status, string(body))
	}

	// Decode response to get file ID
	var f File
	if err := json.NewDecoder(resp.Body).Decode(&f); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("File %s uploaded successfully with ID: %s\n", filePath, f.ID)
	return f.ID, nil
}

// ListFiles retrieves a list of all files uploaded to ChatGPT
func ListFiles() ([]File, error) {
	url := "https://api.openai.com/v1/files"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("retrieving files failed with status %s: %s", resp.Status, string(body))
	}

	// Parse the response
	var response struct {
		Data []File `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data, nil
}

// RetrieveFile retrieves information about a specific file by file ID
func RetrieveFile(fileID string) (*File, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/files/%s", fileID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create retrieve file request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("file retrieval request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("file retrieval failed with status %s: %s", resp.Status, string(body))
	}

	var file File
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, fmt.Errorf("failed to decode file retrieval response: %w", err)
	}

	fmt.Printf("File %s retrieved successfully with ID: %s\n", file.FileName, file.ID)
	return &file, nil
}

// DeleteFile deletes a file from ChatGPT by file ID
func DeleteFile(fileID string) error {
	url := fmt.Sprintf("https://api.openai.com/v1/files/%s", fileID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("file deletion failed with status %s: %s", resp.Status, string(body))
	}

	fmt.Printf("File with ID %s deleted successfully.\n", fileID)
	return nil
}
