package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// テスト用ヘルパー: モックサーバーのURLを使うクライアントを作成
func newTestClient(serverURL string) *Client {
	config := openai.DefaultConfig("test-key")
	config.BaseURL = serverURL
	return &Client{client: openai.NewClientWithConfig(config)}
}

func TestNewClient(t *testing.T) {
	cfg := Config{
		APIKey: "dummy_key",
		OrgID:  "dummy_org",
	}
	client := NewClient(cfg)

	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
}

func TestCreateChatCompletion(t *testing.T) {
	t.Run("正常にチャット補完を返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: "こんにちは"}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		result, err := client.CreateChatCompletion(context.Background(), ChatRequest{
			Messages: []ChatMessage{{Role: "user", Content: "hello"}},
		})

		require.NoError(t, err)
		assert.Equal(t, "こんにちは", result)
	})

	t.Run("空の Choices でエラーを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{}}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		_, err := client.CreateChatCompletion(context.Background(), ChatRequest{
			Messages: []ChatMessage{{Role: "user", Content: "hello"}},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no choices returned")
	})

	t.Run("APIエラー時にラップされたエラーを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":{"message":"server error"}}`))
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		_, err := client.CreateChatCompletion(context.Background(), ChatRequest{
			Messages: []ChatMessage{{Role: "user", Content: "hello"}},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create chat completion")
	})
}

func TestCreateImageAnalysis(t *testing.T) {
	t.Run("正常に画像解析結果を返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: `{"description":"テスト画像"}`}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := newTestClient(server.URL)
		result, err := client.CreateImageAnalysis(context.Background(), ImageAnalysisRequest{
			SystemPrompt: "Analyze",
			UserPrompt:   "What is this?",
			ImageURL:     "https://example.com/image.jpg",
		})

		require.NoError(t, err)
		assert.Contains(t, result, "テスト画像")
	})
}

func TestAnalyzeImage(t *testing.T) {
	t.Run("ジェネリクスによる型安全なデコード", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: `{"name":"テスト","count":42}`}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		type TestResult struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}

		client := newTestClient(server.URL)
		result, err := AnalyzeImage[TestResult](client, context.Background(), ImageAnalysisRequest{
			SystemPrompt: "Analyze",
			UserPrompt:   "Count items",
			ImageURL:     "https://example.com/image.jpg",
		})

		require.NoError(t, err)
		assert.Equal(t, "テスト", result.Name)
		assert.Equal(t, 42, result.Count)
	})

	t.Run("不正なJSONでデコードエラーを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: "invalid json"}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		type TestResult struct {
			Name string `json:"name"`
		}

		client := newTestClient(server.URL)
		_, err := AnalyzeImage[TestResult](client, context.Background(), ImageAnalysisRequest{
			SystemPrompt: "Analyze",
			UserPrompt:   "What?",
			ImageURL:     "https://example.com/image.jpg",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})
}
