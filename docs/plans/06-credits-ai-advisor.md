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

### F.2 Очки игрока (points scoring) (приоритет: LOW)

**Контекст:** В legacy (consts.php — Dominator) нестандартные коэффициенты очков:

| Параметр | Значение Dominator | Дефолт |
|----------|-------------------|--------|
| Очки за юниты (флот/оборону) | `2.0/1000` | `2.0/1000` |
| Очки за исследования | `0.5/1000` | `1.0/1000` |
| Очки за постройки | `0.05/1000` | `0.5/1000` |

Формула: `points += resource_cost × coefficient`.

В nova очки или не считаются, или не конфигурируемые. Нужно:
- Добавить `PointsCoefficients` в `config.go`
- Пересчитывать очки при постройке/исследовании/победе в бою
- Хранить в `users.points`, индексировать для рейтинга

**Шаг 1** — Миграция: `ALTER TABLE users ADD COLUMN points BIGINT DEFAULT 0`
**Шаг 2** — `scoring/service.go`: `AddBuildingPoints / AddResearchPoints / AddFleetPoints`
**Шаг 3** — Вызовы из `building/service.go`, `research/service.go`, `fleet/attack.go`
**Шаг 4** — `GET /api/leaderboard` — топ по очкам

**Проверка готовности:**
- [ ] `users.points` в БД
- [ ] Очки начисляются за постройки/исследования/флот
- [ ] `/api/leaderboard` endpoint
- [ ] `make test` зелёный

---

### F.3 AI-советник (план 07, приоритет: LOW — новая фича, не legacy-порт)

**Концепция:** AI-помощник, отвечает на вопросы об игре с учётом состояния игрока.
Два режима: Claude API (за кредиты) или Ollama (бесплатно, локально на сервере).

**Цены:**

| Модель | ID | Кредиты |
|--------|----|---------|
| Haiku | claude-haiku-4-5-20251001 | 5 кр |
| Sonnet | claude-sonnet-4-6 | 20 кр |
| Opus | claude-opus-4-7 | 80 кр |

Кредиты списываются **после** получения ответа (ошибка API → не списывается).

**Шаг 1** — Зависимость: `go get github.com/anthropics/anthropic-sdk-go`

**Шаг 2** — `backend/internal/config/config.go`:
```go
type AIAdvisorConfig struct {
    APIKey      string  // ANTHROPIC_API_KEY
    ProxyURL    string  // ANTHROPIC_PROXY_URL (опционально)
    OllamaURL   string  // OLLAMA_URL (если задан — использовать Ollama)
    OllamaModel string  // OLLAMA_MODEL, default "qwen2.5:3b"
    MaxPerDay   int     // AI_ADVISOR_MAX_PER_DAY, default 20
    MaxTokens   int     // AI_ADVISOR_MAX_TOKENS, default 1024
}
```

**Шаг 3** — `backend/internal/aiadvisor/`: интерфейс `LLMClient`, реализации
`AnthropicClient` (с поддержкой прокси) и `OllamaClient`.

**Шаг 4** — `aiadvisor/service.go`: `Ask(ctx, userID, model, question)`:
1. Проверить баланс кредитов
2. `buildPlayerContext` — планеты, здания, исследования, флоты, артефакты
3. `buildStaticGameKnowledge` — из configs/ (кэшируемая часть prompt)
4. Вызов LLM
5. Списать кредиты

**Шаг 5** — Миграция: `ai_advisor_log(id, user_id, question, answer, tokens_used, created_at)`
для rate limit (20 вопросов/день/игрок).

**Шаг 6** — Endpoint: `POST /api/ai-advisor/ask`, `GET /api/ai-advisor/estimate`

**Шаг 7** — Frontend: `AIAdvisorWidget.tsx` — floating кнопка, выбор модели,
превью стоимости, textarea с вопросом, отображение ответа.

**Ollama для локального режима:** `qwen2.5:3b` (2 GB) на VPS с 4 GB RAM, бесплатно
для игроков, ниже качество. При `OLLAMA_URL` задан — Claude не используется.

**Зависимость:** F.1 (платёжная система) — для полноценного использования кредитной
экономики. AI-советник можно запустить и без неё (только бесплатные кредиты).

**Проверка готовности:**
- [ ] `AIAdvisorConfig` в config.go
- [ ] `LLMClient` интерфейс + AnthropicClient + OllamaClient
- [ ] `aiadvisor/service.go`: Ask, buildPlayerContext, rate limit
- [ ] Миграция `ai_advisor_log`
- [ ] Endpoint `/api/ai-advisor/ask`
- [ ] `AIAdvisorWidget.tsx` с выбором модели и превью стоимости
- [ ] Тест rate limit, тест EstimateCost
