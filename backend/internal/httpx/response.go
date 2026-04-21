package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Error представляет стандартизированную ошибку API.
//
// В теле ответа — {"error":{"code":"...","message":"..."}}.
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// Стандартные сэндвичи, чтобы handler-ы возвращали их через errors.As.
var (
	ErrBadRequest   = &Error{Status: http.StatusBadRequest, Code: "bad_request", Message: "bad request"}
	ErrUnauthorized = &Error{Status: http.StatusUnauthorized, Code: "unauthorized", Message: "unauthorized"}
	ErrForbidden    = &Error{Status: http.StatusForbidden, Code: "forbidden", Message: "forbidden"}
	ErrNotFound     = &Error{Status: http.StatusNotFound, Code: "not_found", Message: "not found"}
	ErrConflict     = &Error{Status: http.StatusConflict, Code: "conflict", Message: "conflict"}
	ErrInternal     = &Error{Status: http.StatusInternalServerError, Code: "internal", Message: "internal error"}
)

// WriteJSON сериализует value как JSON. Ошибки маршалинга не возвращаются
// наружу (уже поздно), вместо этого логируются.
func WriteJSON(w http.ResponseWriter, r *http.Request, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if value == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.ErrorContext(r.Context(), "http_write_json_failed", slog.String("err", err.Error()))
	}
}

// WriteError унифицированно сериализует ошибку. Если err — это *Error, берёт
// его поля; иначе превращает в 500.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var e *Error
	if !errors.As(err, &e) {
		e = &Error{Status: http.StatusInternalServerError, Code: "internal", Message: err.Error()}
	}
	WriteJSON(w, r, e.Status, map[string]any{
		"error": map[string]any{
			"code":    e.Code,
			"message": e.Message,
		},
	})
}

// Wrap возвращает новую *Error с подставленным сообщением.
func Wrap(base *Error, message string) *Error {
	return &Error{Status: base.Status, Code: base.Code, Message: message}
}
