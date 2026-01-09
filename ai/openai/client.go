package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/newmo-oss/ergo"
	"github.com/sashabaranov/go-openai"
)

// Client は OpenAI API へのアクセスクライアントです。
type Client struct {
	client *openai.Client
}

// Config は OpenAI クライアントの設定です。
type Config struct {
	APIKey string
	OrgID  string // オプショナル
}

// NewClient は新しい OpenAI クライアントを作成します。
func NewClient(cfg Config) *Client {
	config := openai.DefaultConfig(cfg.APIKey)
	if cfg.OrgID != "" {
		config.OrgID = cfg.OrgID
	}

	client := openai.NewClientWithConfig(config)
	return &Client{
		client: client,
	}
}

// ChatMessage はチャット補完リクエストのメッセージ構造です。
type ChatMessage struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// ChatRequest はチャット補完リクエストのパラメータです。
type ChatRequest struct {
	Model       string
	Messages    []ChatMessage
	Temperature *float32 // 0.0 ~ 2.0 (デフォルト: 1.0)
	MaxTokens   int      // 0の場合はモデルの上限まで
}

// CreateChatCompletion はチャット補完APIを呼び出します。
func (c *Client) CreateChatCompletion(ctx context.Context, req ChatRequest) (string, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	// デフォルトパラメータの調整
	var temperature float32 = 1.0
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	model := req.Model
	if model == "" {
		model = openai.GPT4o // デフォルトモデル (2026年時点での推奨に合わせて変更してください)
	}

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: temperature,
			MaxTokens:   req.MaxTokens,
		},
	)

	if err != nil {
		return "", ergo.Wrap(err, "failed to create chat completion")
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

// ImageAnalysisRequest は画像解析リクエストのパラメータです。
type ImageAnalysisRequest struct {
	Model        string
	SystemPrompt string // AIへの指示（JSONスキーマの説明など）
	UserPrompt   string // ユーザーからの質問
	ImageURL     string // 画像のURL (http://... または data:image/jpeg;base64,...)
	Temperature  *float32
	MaxTokens    int
}

// CreateImageAnalysis は画像を解析し、JSON形式で結果を返します。
// システムプロンプトで必ず "Return JSON" と指示し、スキーマを明示することを推奨します。
func (c *Client) CreateImageAnalysis(ctx context.Context, req ImageAnalysisRequest) (string, error) {
	// デフォルトパラメータ
	model := req.Model
	if model == "" {
		model = openai.GPT4o
	}
	var temperature float32 = 0.0 // 解析タスクなのでランダム性を下げる
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1000
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		},
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: req.UserPrompt,
				},
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL:    req.ImageURL,
						Detail: openai.ImageURLDetailAuto,
					},
				},
			},
		},
	}

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: temperature,
			MaxTokens:   maxTokens,
			// 強制的にJSONモードを有効化 (モデルが対応している必要あり: gpt-4-turbo, gpt-4oなど)
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return "", ergo.Wrap(err, "failed to analyze image")
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

// AnalyzeImage は画像を解析し、結果を指定された構造体 T にデコードして返します。
// Go言語の仕様上、メソッドに型パラメータを持たせることができないため、パッケージ関数として提供します。
//
// 使用例:
//
//	type Result struct {
//		Description string `json:"description"`
//	}
//	result, err := openai.AnalyzeImage[Result](client, ctx, req)
func AnalyzeImage[T any](c *Client, ctx context.Context, req ImageAnalysisRequest) (*T, error) {
	jsonStr, err := c.CreateImageAnalysis(ctx, req)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, ergo.Wrap(err, "failed to unmarshal analysis result")
	}

	return &result, nil
}
