package urlshortener

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	ctxKeyOp     = "operation"
	opShortenURL = "clients.ShortenURL"
)

// LoggingInterceptor - интерсептор для логирования
type LoggingInterceptor struct {
	next http.RoundTripper
	log  *slog.Logger
}

func NewLoggingInterceptor(log *slog.Logger, next http.RoundTripper) *LoggingInterceptor {
	if next == nil {
		next = http.DefaultTransport
	}
	return &LoggingInterceptor{
		next: next,
		log:  log,
	}
}

func (i *LoggingInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Логируем исходящий запрос
	i.log.Debug("HTTP request started",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.String("operation", req.Context().Value(ctxKeyOp).(string)),
	)

	// Выполняем запрос
	resp, err := i.next.RoundTrip(req)

	duration := time.Since(start)

	if err != nil {
		i.log.Error("HTTP request failed",
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
			slog.String("error", err.Error()),
			slog.Duration("duration", duration),
		)
		return nil, err
	}

	// Логируем успешный ответ
	i.log.Debug("HTTP request completed",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Int("status_code", resp.StatusCode),
		slog.Duration("duration", duration),
	)

	return resp, nil
}

// RetryInterceptor - интерсептор для повторных попыток
type RetryInterceptor struct {
	next       http.RoundTripper
	maxRetries int
	timeout    time.Duration
	log        *slog.Logger
}

func NewRetryInterceptor(log *slog.Logger, maxRetries int, timeout time.Duration, next http.RoundTripper) *RetryInterceptor {
	if next == nil {
		next = http.DefaultTransport
	}
	return &RetryInterceptor{
		next:       next,
		maxRetries: maxRetries,
		timeout:    timeout,
		log:        log,
	}
}

func (i *RetryInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastErr error
	var lastResp *http.Response

	for attempt := 0; attempt <= i.maxRetries; attempt++ {
		if attempt > 0 {
			i.log.Debug("retrying HTTP request",
				slog.Int("attempt", attempt),
				slog.String("url", req.URL.String()),
			)

			// Ждем перед повторной попыткой (exponential backoff можно добавить)
			select {
			case <-time.After(i.timeout / 2):
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}

			// Пересоздаем тело запроса, т.к. оно могло быть прочитано
			if req.GetBody != nil {
				if body, err := req.GetBody(); err == nil {
					req.Body = body
				}
			}
		}

		// Выполняем запрос
		resp, err := i.next.RoundTrip(req)

		if err != nil {
			lastErr = err
			i.log.Warn("HTTP request failed, will retry",
				slog.String("error", err.Error()),
				slog.Int("attempt", attempt),
			)
			continue
		}

		// Проверяем нужно ли повторять на основе статус кода
		if i.shouldRetry(resp.StatusCode) {
			lastResp = resp
			i.log.Warn("HTTP request returned retryable status",
				slog.Int("status_code", resp.StatusCode),
				slog.Int("attempt", attempt),
			)
			continue
		}

		// Успешный ответ или ошибка которая не требует retry
		return resp, nil
	}

	// Все попытки исчерпаны
	if lastResp != nil {
		return lastResp, nil
	}
	return nil, fmt.Errorf("all %d attempts failed: %w", i.maxRetries, lastErr)
}

func (i *RetryInterceptor) shouldRetry(statusCode int) bool {
	// Retry на временных ошибках сервера и too many requests
	return statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusInternalServerError ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout
}

type Client struct {
	httpClient       *http.Client
	shortenerBaseUrl string
	log              *slog.Logger
}

type ShortenRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

type ShortenResponse struct {
	URL   string `json:"url"`
	Alias string `json:"alias"`
}

func (r *ShortenResponse) GetURL() string {
	return r.URL
}

func (r *ShortenResponse) GetAlias() string {
	return r.Alias
}

// NewClient create new client and connect interceptors
func NewClient(log *slog.Logger, timeout time.Duration, maxRetries int, baseUrl string) *Client {
	var transport = http.DefaultTransport

	transport = NewRetryInterceptor(log, maxRetries, timeout, transport)

	transport = NewLoggingInterceptor(log, transport)

	return &Client{
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		shortenerBaseUrl: baseUrl,
	}
}

func (c *Client) ShortenURL(ctx context.Context, originalURL, alias string) (*ShortenResponse, error) {
	ctx = context.WithValue(ctx, ctxKeyOp, opShortenURL)

	requestBody := ShortenRequest{
		URL:   originalURL,
		Alias: alias,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opShortenURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.shortenerBaseUrl+"", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opShortenURL, err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opShortenURL, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opShortenURL, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d: %s", opShortenURL, resp.StatusCode, string(body))
	}

	var shortenResponse ShortenResponse
	if err := json.Unmarshal(body, &shortenResponse); err != nil {
		return nil, fmt.Errorf("%s: %w", opShortenURL, err)
	}

	c.log.Debug("URL Shortened successfully",
		slog.String("alias", shortenResponse.Alias),
		slog.String("url", originalURL),
	)

	return &shortenResponse, nil
}
