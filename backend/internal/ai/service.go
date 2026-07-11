package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Mars-60/project4/backend/internal/core"
)

type Client interface {
	Complete(ctx context.Context, messages []core.AIMessage) (string, error)
}

type GroqClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
	maxRetries int
}

func NewGroqClient(baseURL string, apiKey string, model string, timeout time.Duration, maxRetries int) *GroqClient {
	return &GroqClient{baseURL: baseURL, apiKey: apiKey, model: model, httpClient: &http.Client{Timeout: timeout}, maxRetries: maxRetries}
}

func (c *GroqClient) Complete(ctx context.Context, messages []core.AIMessage) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("groq api key is not configured")
	}
	payload := map[string]any{
		"model":       c.model,
		"messages":    mapMessages(messages),
		"temperature": 0.2,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleep(ctx, time.Duration(attempt)*300*time.Millisecond); err != nil {
				return "", err
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		responseBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return "", readErr
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError {
			lastErr = fmt.Errorf("groq retryable status %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode >= http.StatusBadRequest {
			return "", fmt.Errorf("groq status %d: %s", resp.StatusCode, string(responseBody))
		}
		var parsed groqResponse
		if err := json.Unmarshal(responseBody, &parsed); err != nil {
			return "", err
		}
		if len(parsed.Choices) == 0 {
			return "", fmt.Errorf("groq returned no choices")
		}
		return parsed.Choices[0].Message.Content, nil
	}
	return "", lastErr
}

type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func mapMessages(messages []core.AIMessage) []map[string]string {
	out := make([]map[string]string, 0, len(messages)+1)
	out = append(out, map[string]string{
		"role":    "system",
		"content": "You are TradePilot AI. Assist with analysis, education, risk review, and explanations. Never place trades or claim that you executed an order.",
	})
	for _, message := range messages {
		out = append(out, map[string]string{"role": message.Role, "content": message.Content})
	}
	return out
}

func sleep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type PromptManager struct{}

func (PromptManager) PortfolioAnalysis(summary core.PortfolioSummary) core.AIMessage {
	return core.AIMessage{Role: "user", Content: fmt.Sprintf("Analyze this portfolio risk and performance. Net %.2f, cash %.2f, exposure %.2f, unrealized PnL %.2f.", summary.NetValue, summary.Cash, summary.Exposure, summary.UnrealizedPnL)}
}

func (PromptManager) StrategyExplanation(strategy core.StrategyDefinition) core.AIMessage {
	return core.AIMessage{Role: "user", Content: fmt.Sprintf("Explain this strategy and risk controls: %s - %s", strategy.Name, strategy.Description)}
}

func (PromptManager) MarketExplanation(snapshot core.MarketSnapshot) core.AIMessage {
	return core.AIMessage{Role: "user", Content: fmt.Sprintf("Explain this market snapshot without giving trade execution instructions. Symbol %s, last %.2f, open %.2f, high %.2f, low %.2f, close %.2f, volume %d.", snapshot.TradingSymbol, snapshot.LastPrice, snapshot.Open, snapshot.High, snapshot.Low, snapshot.Close, snapshot.Volume)}
}

func (PromptManager) RiskExplanation(summary core.PortfolioSummary) core.AIMessage {
	return core.AIMessage{Role: "user", Content: fmt.Sprintf("Explain the risk profile and risk controls for this portfolio. Net %.2f, cash %.2f, exposure %.2f, realized PnL %.2f, unrealized PnL %.2f. Do not place trades.", summary.NetValue, summary.Cash, summary.Exposure, summary.RealizedPnL, summary.UnrealizedPnL)}
}

type ConversationService struct {
	client Client
	repo   core.ConversationRepository
	clock  core.Clock
	ids    core.IDGenerator
}

func NewConversationService(client Client, repo core.ConversationRepository, clock core.Clock, ids core.IDGenerator) *ConversationService {
	return &ConversationService{client: client, repo: repo, clock: clock, ids: ids}
}

func (s *ConversationService) Ask(ctx context.Context, userID core.ID, prompt string) (core.AIConversation, error) {
	return s.askWithMessages(ctx, userID, "Trading Assistant", []core.AIMessage{{Role: "user", Content: prompt, CreatedAt: s.clock.Now()}})
}

func (s *ConversationService) AnalyzePortfolio(ctx context.Context, userID core.ID, summary core.PortfolioSummary) (core.AIConversation, error) {
	message := PromptManager{}.PortfolioAnalysis(summary)
	message.CreatedAt = s.clock.Now()
	return s.askWithMessages(ctx, userID, "Portfolio Analysis", []core.AIMessage{message})
}

func (s *ConversationService) ExplainStrategy(ctx context.Context, userID core.ID, strategy core.StrategyDefinition) (core.AIConversation, error) {
	message := PromptManager{}.StrategyExplanation(strategy)
	message.CreatedAt = s.clock.Now()
	return s.askWithMessages(ctx, userID, "Strategy Explanation", []core.AIMessage{message})
}

func (s *ConversationService) ExplainMarket(ctx context.Context, userID core.ID, snapshot core.MarketSnapshot) (core.AIConversation, error) {
	message := PromptManager{}.MarketExplanation(snapshot)
	message.CreatedAt = s.clock.Now()
	return s.askWithMessages(ctx, userID, "Market Explanation", []core.AIMessage{message})
}

func (s *ConversationService) ExplainRisk(ctx context.Context, userID core.ID, summary core.PortfolioSummary) (core.AIConversation, error) {
	message := PromptManager{}.RiskExplanation(summary)
	message.CreatedAt = s.clock.Now()
	return s.askWithMessages(ctx, userID, "Risk Explanation", []core.AIMessage{message})
}

func (s *ConversationService) askWithMessages(ctx context.Context, userID core.ID, title string, messages []core.AIMessage) (core.AIConversation, error) {
	now := s.clock.Now()
	answer, err := s.client.Complete(ctx, messages)
	if err != nil {
		return core.AIConversation{}, err
	}
	messages = append(messages, core.AIMessage{Role: "assistant", Content: answer, CreatedAt: s.clock.Now()})
	conversation := core.AIConversation{ID: s.ids.NewID(), UserID: userID, Title: title, Messages: messages, CreatedAt: now, UpdatedAt: s.clock.Now()}
	return conversation, s.repo.SaveConversation(ctx, conversation)
}
