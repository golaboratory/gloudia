package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

// newTestClient is defined in client_test.go (same package) and reused here.

func TestCreateImageAnalysis_ErrorPaths(t *testing.T) {
	t.Run("empty choices returns sentinel", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{}}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		_, err := client.CreateImageAnalysis(context.Background(), ImageAnalysisRequest{
			SystemPrompt: "Analyze",
			UserPrompt:   "What?",
			ImageURL:     "https://example.com/i.jpg",
		})
		assert.ErrorIs(t, err, ErrNoChoicesReturned)
	})

	t.Run("api error is wrapped", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":{"message":"server error"}}`))
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		_, err := client.CreateImageAnalysis(context.Background(), ImageAnalysisRequest{
			SystemPrompt: "Analyze",
			UserPrompt:   "What?",
			ImageURL:     "https://example.com/i.jpg",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to analyze image")
	})
}

func TestAnalyzeImage_PropagatesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"boom"}}`))
	}))
	defer server.Close()

	type Result struct {
		Name string `json:"name"`
	}
	client := newTestClient(server.URL)
	res, err := AnalyzeImage[Result](client, context.Background(), ImageAnalysisRequest{
		SystemPrompt: "Analyze",
		UserPrompt:   "What?",
		ImageURL:     "https://example.com/i.jpg",
	})
	assert.Nil(t, res)
	assert.Error(t, err)
	// CreateImageAnalysis のエラーがそのまま伝播し、unmarshal 段階には到達しないこと。
	assert.Contains(t, err.Error(), "failed to analyze image")
	assert.NotContains(t, err.Error(), "failed to unmarshal")
}
