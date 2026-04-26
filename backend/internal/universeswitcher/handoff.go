package universeswitcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// HandoffReceive — GET /auth/handoff?code=<token>
// Вызывается браузером при переходе в эту вселенную из другой.
// Обменивает handoff-токен на full JWT через auth-service и
// устанавливает access_token в localStorage через JS-страницу.
func (h *Handler) HandoffReceive(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	tokens, err := h.exchangeHandoffCode(r.Context(), code)
	if err != nil {
		http.Error(w, "invalid or expired handoff token", http.StatusUnauthorized)
		return
	}

	// Возвращаем мини-HTML, который кладёт токен в localStorage и редиректит на /
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><html><body><script>
localStorage.setItem('access_token', %q);
localStorage.setItem('refresh_token', %q);
window.location.replace('/');
</script></body></html>`, tokens.Access, tokens.Refresh)
}

type tokenPair struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

func (h *Handler) exchangeHandoffCode(ctx context.Context, code string) (tokenPair, error) {
	body, _ := json.Marshal(map[string]string{"code": code})
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		h.authServiceURL+"/auth/token/exchange",
		bytes.NewReader(body),
	)
	if err != nil {
		return tokenPair{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return tokenPair{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tokenPair{}, fmt.Errorf("exchange returned %d", resp.StatusCode)
	}

	var out struct {
		Tokens tokenPair `json:"tokens"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return tokenPair{}, err
	}
	return out.Tokens, nil
}
