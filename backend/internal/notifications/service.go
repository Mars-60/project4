package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Mars-60/project4/backend/internal/core"
)

type Provider interface {
	Channel() string
	Send(ctx context.Context, notification core.Notification) error
}

type Service struct {
	mu        sync.RWMutex
	providers map[string]Provider
	repo      core.NotificationRepository
	clock     core.Clock
	ids       core.IDGenerator
}

func NewService(repo core.NotificationRepository, clock core.Clock, ids core.IDGenerator) *Service {
	return &Service{providers: make(map[string]Provider), repo: repo, clock: clock, ids: ids}
}

func (s *Service) Register(provider Provider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[provider.Channel()] = provider
}

func (s *Service) Notify(ctx context.Context, userID core.ID, channel string, subject string, body string) (core.Notification, error) {
	notification := core.Notification{ID: s.ids.NewID(), UserID: userID, Channel: channel, Subject: subject, Body: body, Status: "queued", CreatedAt: s.clock.Now()}
	s.mu.RLock()
	provider, ok := s.providers[channel]
	s.mu.RUnlock()
	if ok {
		if err := provider.Send(ctx, notification); err != nil {
			notification.Status = "failed"
			_, _ = s.repo.CreateNotification(ctx, notification)
			return notification, err
		}
		notification.Status = "sent"
	}
	if !ok {
		notification.Status = "queued"
	}
	return s.repo.CreateNotification(ctx, notification)
}

type LogProvider struct {
	channel string
}

func NewLogProvider(channel string) LogProvider {
	return LogProvider{channel: channel}
}

func (p LogProvider) Channel() string { return p.channel }

func (p LogProvider) Send(ctx context.Context, notification core.Notification) error {
	if notification.Body == "" {
		return fmt.Errorf("notification body is empty")
	}
	return nil
}

type WebhookProvider struct {
	channel string
	url     string
	client  *http.Client
}

func NewWebhookProvider(channel string, url string, timeout time.Duration) WebhookProvider {
	return WebhookProvider{channel: channel, url: url, client: &http.Client{Timeout: timeout}}
}

func (p WebhookProvider) Channel() string { return p.channel }

func (p WebhookProvider) Send(ctx context.Context, notification core.Notification) error {
	if p.url == "" {
		return fmt.Errorf("%s provider is not configured", p.channel)
	}
	payload, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("%s provider returned status %d", p.channel, resp.StatusCode)
	}
	return nil
}
