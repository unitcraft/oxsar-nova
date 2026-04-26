package portalsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CreditClient вызывает auth-service для списания кредитов.
type CreditClient struct {
	authServiceURL string
	httpClient     *http.Client
}

// NewCreditClient создаёт клиент для работы с кредитами.
func NewCreditClient(authServiceURL string) *CreditClient {
	return &CreditClient{
		authServiceURL: authServiceURL,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Spend списывает amount кредитов у пользователя через auth-service.
func (c *CreditClient) Spend(ctx context.Context, userID string, amount int64, reason, refID string) error {
	if c.authServiceURL == "" {
		return nil // dev-режим без auth-service
	}
	body, _ := json.Marshal(map[string]any{
		"user_id": userID,
		"amount":  amount,
		"reason":  reason,
		"ref_id":  refID,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.authServiceURL+"/auth/credits/spend", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		return ErrInsufficientCredits
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("auth-service credits/spend: status %d", resp.StatusCode)
	}
	return nil
}
