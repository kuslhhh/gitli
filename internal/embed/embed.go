package embed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

// DefaultOllamaURL is the default Ollama API endpoint.
const DefaultOllamaURL = "http://localhost:11434"

// DefaultModel is the recommended embedding model.
const DefaultModel = "nomic-embed-text"

// Embedder generates embeddings for text.
type Embedder struct {
	client    *http.Client
	baseURL   string
	model     string
}

// New creates a new Embedder that calls the Ollama API.
func New(baseURL, model string) *Embedder {
	if baseURL == "" {
		baseURL = DefaultOllamaURL
	}
	if model == "" {
		model = DefaultModel
	}
	return &Embedder{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
	}
}

// ollamaEmbedRequest is the request body for Ollama's /api/embed endpoint.
type ollamaEmbedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// ollamaEmbedResponse is the response from Ollama's /api/embed endpoint.
type ollamaEmbedResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
}

// Embed generates an embedding vector for the given text using Ollama.
// Returns nil if Ollama is not available.
func (e *Embedder) Embed(text string) ([]float64, error) {
	url := e.baseURL + "/api/embed"

	reqBody := ollamaEmbedRequest{
		Model: e.model,
		Input: text,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := e.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("call Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}

	var result ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("Ollama returned no embeddings")
	}

	return result.Embeddings[0], nil
}

// IsAvailable quickly checks whether Ollama is running with a short timeout.
func (e *Embedder) IsAvailable() bool {
	// Use a fast connection check with a 1s timeout to avoid blocking
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(e.baseURL + "/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// CosineSimilarity computes the cosine similarity between two vectors.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
