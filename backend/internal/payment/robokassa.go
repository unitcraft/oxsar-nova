package payment

import (
	"context"
	"crypto/md5" //nolint:gosec // Робокасса использует MD5 по своему протоколу
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const robokassaBaseURL = "https://auth.robokassa.ru/Merchant/Index.aspx"

// RobokassaGateway реализует Gateway для Робокассы.
type RobokassaGateway struct {
	login string
	pass1 string
	pass2 string
}

func NewRobokassaGateway(login, pass1, pass2 string) *RobokassaGateway {
	return &RobokassaGateway{login: login, pass1: pass1, pass2: pass2}
}

// BuildPayURL формирует ссылку оплаты.
// Подпись: MD5(login:OutSum:InvId:pass1) — по документации Робокассы.
func (g *RobokassaGateway) BuildPayURL(_ context.Context, orderID, description string, amountKop int, returnURL string) (string, error) {
	outSum := fmt.Sprintf("%.2f", float64(amountKop)/100)
	sig := robokassaMD5(g.login, outSum, orderID, g.pass1)

	q := url.Values{}
	q.Set("MerchantLogin", g.login)
	q.Set("OutSum", outSum)
	q.Set("InvId", orderID)
	q.Set("Description", description)
	q.Set("SignatureValue", sig)
	q.Set("Encoding", "utf-8")
	if returnURL != "" {
		q.Set("ReturnUrl", returnURL)
	}

	return robokassaBaseURL + "?" + q.Encode(), nil
}

// VerifyWebhook проверяет подпись входящего POST от Робокассы.
// Ожидаемые поля: OutSum, InvId, SignatureValue.
// Подпись: MD5(OutSum:InvId:pass2).
func (g *RobokassaGateway) VerifyWebhook(r *http.Request) (orderID string, amountKop int, err error) {
	if err = r.ParseForm(); err != nil {
		return "", 0, ErrWebhookInvalid
	}

	outSum := r.FormValue("OutSum")
	invID := r.FormValue("InvId")
	sig := r.FormValue("SignatureValue")

	if outSum == "" || invID == "" || sig == "" {
		return "", 0, ErrWebhookInvalid
	}

	expected := robokassaMD5(outSum, invID, g.pass2)
	if !strings.EqualFold(expected, sig) {
		return "", 0, ErrWebhookInvalid
	}

	var rub float64
	if _, scanErr := fmt.Sscanf(outSum, "%f", &rub); scanErr != nil {
		return "", 0, ErrWebhookInvalid
	}

	return invID, int(rub * 100), nil
}

// SuccessResponse пишет ответ, ожидаемый Робокассой: "OK{InvId}".
func (g *RobokassaGateway) SuccessResponse(w http.ResponseWriter, orderID string) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "OK%s", orderID)
}

func robokassaMD5(parts ...string) string {
	h := md5.New() //nolint:gosec
	h.Write([]byte(strings.Join(parts, ":")))
	return fmt.Sprintf("%x", h.Sum(nil))
}
