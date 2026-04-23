# План F: Кредиты и AI-советник

---

## История (завершено)

### ✅ Базовая экономика кредитов (план 11, итерации ~38–45)
- Поле `credit numeric(15,2)` в `users`, стартовое значение 5.00
- Отображение баланса в шапке UI (💳 N cr)
- `migration/0042_daily_credit.sql` — ежедневный бонус при входе (+10/день)
- Кредиты за победу в бою (`fleet/attack.go`)
- Кредиты за достижения (`achievement/service.go`)
- Кредиты в исходе экспедиции `credit` (`fleet/expedition.go`)

---

## Открытые задачи

### F.1 Платёжная система (приоритет: MEDIUM — нужна для монетизации)

**Контекст:** Курс 1000 кр = 100 руб (1 кр = 0.1 руб). Пакеты:

| Пакет | Кредиты | Цена | Бонус |
|-------|---------|------|-------|
| **Пробный** | **400** | **49 руб** | — |
| Стартовый | 1 000 | 100 руб | — |
| Средний | 3 000 | 250 руб | +200 кр |
| Большой | 7 000 | 500 руб | +500 кр |
| Максимальный | 15 000 | 1 000 руб | +2 000 кр |

**Пробный пакет (49 руб)** — снижает психологический барьер первой покупки.
Ниже 50 руб нет в типичных F2P. (§15.8 balance-analysis.md)

**Шаги:**
1. Таблица `credit_purchases(id, user_id, amount_credits, price_rub, created_at)` —
   нужна также для `expCredit` в expedition.go (B.2)
2. Интеграция платёжного шлюза (YooMoney / CloudPayments / Stripe)
3. Webhook handler: подтверждение оплаты → зачисление кредитов
4. Реферальный хук: `referral/service.ProcessPurchaseReferral(ctx, buyerID, amount)` (D.3)
5. UI: страница пополнения с пакетами

**Блокер:** Выбор платёжного шлюза (зависит от юрисдикции и наличия договора).

---

### ✅ F.2 Очки игрока (points scoring)

- `PointsCoefficients` в config.go (kBld=0.00005, kRes=0.0005, kUnt=0.002)
- `score/service.go`: `NewServiceWithCoeffs`, `Top`, `PlayerRank`, `PlayerScore`
- `GET /api/highscore`, `GET /api/highscore/me` эндпоинты

---

### ✅ F.3 AI-советник (backend)

- `config.AIAdvisorConfig`: APIKey, ProxyURL, OllamaURL, OllamaModel, MaxPerDay, MaxTokens (из ENV)
- `aiadvisor/llm.go`: `LLMClient` интерфейс; `AnthropicClient` (net/http, без внешних SDK);
  `OllamaClient` (Ollama HTTP API). Цены: Haiku=5кр, Sonnet=20кр, Opus=80кр.
- `aiadvisor/service.go`: `Ask` (rate limit, баланс, LLM вызов, списание кредитов);
  `Estimate` (стоимость без запроса); `buildPlayerContext`; `buildStaticGameKnowledge`
- `migrations/0048_ai_advisor_log.sql`: таблица для rate limit и истории
- `POST /api/ai-advisor/ask`, `GET /api/ai-advisor/estimate`
- Режим: если `OLLAMA_URL` задан — Ollama (бесплатно); иначе Anthropic API (за кредиты)

Остаётся: `AIAdvisorWidget.tsx` (frontend — отдельная задача)
