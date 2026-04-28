package exchange

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
)

// newTestHandler — service+handler с in-memory fakeRepo и fakeExec.
func newTestHandler(t *testing.T) (*Handler, *fakeRepo) {
	t.Helper()
	svc, fr := newSvc(t)
	return NewHandler(svc), fr
}

// withUser — кладёт user_id в context (минимальный auth-bypass для теста).
func withUser(r *http.Request, uid string) *http.Request {
	ctx := context.WithValue(r.Context(), auth.UserIDKey, uid)
	return r.WithContext(ctx)
}

func TestHandler_Create_RequiresIdempotencyKey(t *testing.T) {
	h, _ := newTestHandler(t)
	body := `{"artifact_unit_id":1,"quantity":1,"price_oxsarit":100,"expires_in_hours":24}`
	req := httptest.NewRequest("POST", "/api/exchange/lots", strings.NewReader(body))
	req = withUser(req, "user-A")
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	body := `{"artifact_unit_id":1,"quantity":1,"price_oxsarit":100,"expires_in_hours":24}`
	req := httptest.NewRequest("POST", "/api/exchange/lots", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "k1")
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestHandler_Create_HappyPath(t *testing.T) {
	h, fr := newTestHandler(t)
	seller := "user-A"
	fr.addUser(seller, 0, "planet-A")
	for i := 0; i < 3; i++ {
		fr.addArtefact(seller, 100)
	}
	body := `{"artifact_unit_id":100,"quantity":2,"price_oxsarit":1000,"expires_in_hours":24}`
	req := httptest.NewRequest("POST", "/api/exchange/lots", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "k1")
	req = withUser(req, seller)
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201, body=%s", w.Code, w.Body.String())
	}
	var resp map[string]map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["lot"]["status"] != "active" {
		t.Errorf("status = %v, want active", resp["lot"]["status"])
	}
}

func TestHandler_Create_BadJSON(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest("POST", "/api/exchange/lots", strings.NewReader("not json"))
	req.Header.Set("Idempotency-Key", "k1")
	req = withUser(req, "user-A")
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandler_Create_InsufficientArtefacts(t *testing.T) {
	h, fr := newTestHandler(t)
	seller := "user-A"
	fr.addUser(seller, 0, "planet-A")
	body := `{"artifact_unit_id":100,"quantity":2,"price_oxsarit":1000,"expires_in_hours":24}`
	req := httptest.NewRequest("POST", "/api/exchange/lots", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "k1")
	req = withUser(req, seller)
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandler_Buy_HappyPath(t *testing.T) {
	h, fr := newTestHandler(t)
	seller := "user-S"
	buyer := "user-B"
	fr.addUser(seller, 0, "planet-S")
	fr.addUser(buyer, 5000, "planet-B")
	fr.addArtefact(seller, 100)

	// Сначала создаём лот через сервис (минуя handler).
	svc, _ := newSvc(t)
	svc.repo = fr
	lot, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: 100,
		Quantity: 1, PriceOxsarit: 500, ExpiresInHours: 24,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Используем тот же service для handler'а.
	h = NewHandler(svc)

	r := chi.NewRouter()
	h.Routes(r)
	req := httptest.NewRequest("POST", "/api/exchange/lots/"+lot.ID+"/buy", nil)
	req.Header.Set("Idempotency-Key", "k2")
	req = withUser(req, buyer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_Buy_InsufficientOxsarits(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	buyer := "user-B"
	fr.addUser(seller, 0, "planet-S")
	fr.addUser(buyer, 100, "planet-B")
	fr.addArtefact(seller, 100)
	lot, _ := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: 100,
		Quantity: 1, PriceOxsarit: 1000, ExpiresInHours: 24,
	})
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.Routes(r)
	req := httptest.NewRequest("POST", "/api/exchange/lots/"+lot.ID+"/buy", nil)
	req.Header.Set("Idempotency-Key", "k2")
	req = withUser(req, buyer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusPaymentRequired {
		t.Errorf("status = %d, want 402", w.Code)
	}
}

func TestHandler_Cancel_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)
	r := chi.NewRouter()
	h.Routes(r)
	req := httptest.NewRequest("DELETE", "/api/exchange/lots/nonexistent", nil)
	req = withUser(req, "user-A")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestHandler_Cancel_HappyPath(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	fr.addUser(seller, 0, "planet-S")
	fr.addArtefact(seller, 100)
	lot, _ := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: 100,
		Quantity: 1, PriceOxsarit: 100, ExpiresInHours: 24,
	})
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.Routes(r)
	req := httptest.NewRequest("DELETE", "/api/exchange/lots/"+lot.ID, nil)
	req = withUser(req, seller)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

func TestHandler_Cancel_NotASeller(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	stranger := "user-X"
	fr.addUser(seller, 0, "planet-S")
	fr.addUser(stranger, 0, "planet-X")
	fr.addArtefact(seller, 100)
	lot, _ := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: 100,
		Quantity: 1, PriceOxsarit: 100, ExpiresInHours: 24,
	})
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.Routes(r)
	req := httptest.NewRequest("DELETE", "/api/exchange/lots/"+lot.ID, nil)
	req = withUser(req, stranger)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestHandler_List(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-S"
	fr.addUser(seller, 0, "planet-S")
	for i := 0; i < 3; i++ {
		fr.addArtefact(seller, 100)
	}
	if _, err := svc.CreateLot(context.Background(), CreateLotInput{
		SellerUserID: seller, ArtifactUnitID: 100,
		Quantity: 2, PriceOxsarit: 1000, ExpiresInHours: 24,
	}); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("GET", "/api/exchange/lots?artifact_unit_id=100", nil)
	req = withUser(req, "viewer")
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Lots []map[string]any `json:"lots"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Lots) != 1 {
		t.Errorf("got %d lots, want 1", len(resp.Lots))
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)
	r := chi.NewRouter()
	h.Routes(r)
	req := httptest.NewRequest("GET", "/api/exchange/lots/nonexistent", nil)
	req = withUser(req, "viewer")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// idempotency-проверка: повторный POST с одним и тем же Idempotency-Key
// при отсутствии Redis-middleware — два разных лота. Это поведение
// «raw handler» (middleware применяется на уровне router'а в main.go,
// не в чистом тесте). Документируем фактическое поведение.
func TestHandler_Create_IdempotencyHeaderOnly(t *testing.T) {
	svc, fr := newSvc(t)
	seller := "user-A"
	fr.addUser(seller, 0, "planet-A")
	for i := 0; i < 2; i++ {
		fr.addArtefact(seller, 100)
	}
	h := NewHandler(svc)
	body := `{"artifact_unit_id":100,"quantity":1,"price_oxsarit":100,"expires_in_hours":24}`
	doReq := func() int {
		req := httptest.NewRequest("POST", "/api/exchange/lots", bytes.NewBufferString(body))
		req.Header.Set("Idempotency-Key", "same-key")
		req = withUser(req, seller)
		w := httptest.NewRecorder()
		h.Create(w, req)
		return w.Code
	}
	if c := doReq(); c != http.StatusCreated {
		t.Fatalf("first create: %d", c)
	}
	if c := doReq(); c != http.StatusCreated {
		t.Fatalf("second create (no redis middleware): %d", c)
	}
}
