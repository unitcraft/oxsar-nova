# План: AI-советник (игровой ассистент — Claude API + Ollama)

## Контекст

ИИ-помощник для игроков Oxsar — отвечает на вопросы по игре на человеческом языке,
учитывая текущее состояние конкретного игрока (планеты, ресурсы, флот, исследования).

**Не путать с чатом** (`/chat` — общение игроков между собой, уже реализован
в `backend/internal/chat/` + `ChatScreen.tsx`). Это отдельный AI-советник.

**Legacy:** в oxsar2 ничего подобного нет — это улучшение над legacy.

**Два режима работы:**
- **Claude API** (облако) — высокое качество, стоит кредиты игрока, требует `ANTHROPIC_API_KEY`
- **Ollama** (локально на сервере) — бесплатно для игроков, требует машину с 8+ GB RAM, качество ниже

---

## Что умеет советник

**Примеры вопросов:**
- "Что мне строить дальше?" → анализирует планету и даёт конкретный совет
- "Почему я проигрываю бои?" → смотрит технологии, корабли, опыт
- "Как быстрее накопить металл?" → анализирует здания, производство, артефакты
- "Что такое артефакт Catalyst?" → объясняет механику из каталога игры
- "Когда мне атаковать?" → смотрит флот, рейтинг, цели в галактике
- "Объясни боевую механику" → рассказывает про battle engine

---

## Модели и стоимость в кредитах

Игрок сам выбирает модель при отправке вопроса. Каждый вопрос **предварительно оценивается в кредитах** (игровая валюта) — перед отправкой показывается цена, игрок подтверждает.

### Доступные модели

| Модель | ID | Кредиты за вопрос | Качество |
|--------|-----|-------------------|---------|
| **Haiku** | claude-haiku-4-5-20251001 | 5 cr | Быстро, базово |
| **Sonnet** | claude-sonnet-4-6 | 20 cr | Баланс качества и цены |
| **Opus** | claude-opus-4-7 | 80 cr | Максимальное качество |

### Как рассчитывается цена

Цена предварительно оценивается **до отправки** на основе:
1. Выбранной модели (базовая стоимость)
2. Длины вопроса (больше токенов → дороже)
3. Сложности контекста игрока (размер данных о планетах, флотах)

```go
func EstimateCost(model string, questionLen int, contextSize int) int {
    base := map[string]int{
        "claude-haiku-4-5-20251001": 5,
        "claude-sonnet-4-6":         20,
        "claude-opus-4-7":           80,
    }
    // Доплата за длинный вопрос (>200 символов)
    extra := 0
    if questionLen > 200 {
        extra = base[model] / 2
    }
    return base[model] + extra
}
```

### UI: превью стоимости

```
┌─────────────────────────────────────┐
│ 🤖 AI Советник                      │
│                                     │
│ Выбери модель:                      │
│ ○ Haiku    — 5 cr   (быстро)        │
│ ● Sonnet   — 20 cr  (рекомендуется) │
│ ○ Opus     — 80 cr  (максимум)      │
│                                     │
│ Твой вопрос:                        │
│ [Что мне строить дальше?          ] │
│                                     │
│ 💳 Стоимость: 20 кредитов           │
│ 💳 Баланс после: 480 кредитов       │
│                                     │
│ [Отмена]        [Спросить за 20 cr] │
└─────────────────────────────────────┘
```

### Списание кредитов

Кредиты списываются **после получения ответа** (не до):
- Если Claude API вернул ошибку → кредиты не списываются
- Если ответ получен → `UPDATE users SET credit = credit - N WHERE id = $1`
- Если кредитов недостаточно → ошибка до отправки в Claude

---

## Архитектура

```
Игрок → выбор модели + вопрос
        ↓
GET /api/ai-advisor/estimate?model=sonnet&question=...
        ↓
Показать превью стоимости (N кредитов)
        ↓
Игрок подтверждает → POST /api/ai-advisor/ask
        ↓
aichat/service.go:
  1. Проверить баланс кредитов (>= cost)
  2. Загрузить контекст игрока из БД
     (планеты, здания, исследования, флоты, достижения, рейтинг)
  3. Сформировать system prompt
  4. Вызвать Claude API (выбранная модель)
  5. Списать кредиты (после успешного ответа)
        ↓
Ответ на русском языке → игрок
```

---

## Шаги реализации

### 1. Зависимости

Добавить Anthropic Go SDK:
```bash
go get github.com/anthropics/anthropic-sdk-go
```

Конфигурация через ENV (в соответствии с `backend/internal/config/config.go`):

```go
// В config.go — добавить в Config struct:
type AIAdvisorConfig struct {
    APIKey        string        // ANTHROPIC_API_KEY (обязательный)
    ProxyURL      string        // ANTHROPIC_PROXY_URL (опциональный, напр. "http://proxy:8888")
    MaxPerDay     int           // AI_ADVISOR_MAX_PER_DAY, default 20
    MaxTokens     int           // AI_ADVISOR_MAX_TOKENS, default 1024
}
```

```go
// В Load():
AIAdvisor: AIAdvisorConfig{
    APIKey:    mustEnv("ANTHROPIC_API_KEY"),
    ProxyURL:  env("ANTHROPIC_PROXY_URL", ""),
    MaxPerDay: envInt("AI_ADVISOR_MAX_PER_DAY", 20),
    MaxTokens: envInt("AI_ADVISOR_MAX_TOKENS", 1024),
},
```

В `.env` / `docker-compose.yml`:
```env
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_PROXY_URL=http://proxy.example.com:8888   # опционально
```

### 1a. Инициализация клиента с прокси

API `api.anthropic.com` может быть недоступен напрямую — прокси настраивается через `ANTHROPIC_PROXY_URL`.

```go
// backend/internal/aiadvisor/client.go

func NewAnthropicClient(cfg config.AIAdvisorConfig) *anthropic.Client {
    httpClient := &http.Client{
        Timeout: 60 * time.Second,
    }

    if cfg.ProxyURL != "" {
        proxyURL, err := url.Parse(cfg.ProxyURL)
        if err != nil {
            panic(fmt.Sprintf("invalid ANTHROPIC_PROXY_URL: %v", err))
        }
        httpClient.Transport = &http.Transport{
            Proxy: http.ProxyURL(proxyURL),
        }
    }

    return anthropic.NewClient(
        cfg.APIKey,
        anthropic.WithHTTPClient(httpClient),
    )
}
```

Если `ANTHROPIC_PROXY_URL` не задан — используется стандартный `http.Client` без прокси. Прокси поддерживает HTTP/HTTPS/SOCKS5 (через стандартный Go `net/http`).

### 2. `backend/internal/aiadvisor/service.go`

```go
package aiadvisor

type Service struct {
    db        repo.Exec
    catalog   *config.Catalog
    client    *anthropic.Client
    maxPerDay int
    maxTokens int
}

func (s *Service) Ask(ctx context.Context, userID, question string) (string, error) {
    // Rate limit: проверить количество вопросов сегодня
    if err := s.checkRateLimit(ctx, userID); err != nil {
        return "", err
    }

    // Собрать контекст игрока
    playerCtx, err := s.buildPlayerContext(ctx, userID)
    if err != nil {
        return "", err
    }

    // Вызвать Claude API
    resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
        Model:     anthropic.ModelClaudeSonnet4_6,
        MaxTokens: int64(s.maxPerDay),
        System: []anthropic.TextBlockParam{
            {Text: s.buildSystemPrompt(playerCtx)},
        },
        Messages: []anthropic.MessageParam{
            anthropic.NewUserMessage(question),
        },
    })
    if err != nil {
        return "", fmt.Errorf("claude api: %w", err)
    }

    answer := resp.Content[0].Text

    // Сохранить в лог (для аналитики и rate limit)
    s.logQuestion(ctx, userID, question, answer)

    return answer, nil
}
```

### 3. Контекст игрока (`buildPlayerContext`)

Собираем из БД:
```go
type PlayerContext struct {
    Username    string
    Score       int
    Rank        int
    Planets     []PlanetInfo      // название, тип, ресурсы, здания, production_factor
    Research    map[string]int    // технологии и уровни
    Fleets      []FleetInfo       // активные флоты, миссии
    Ships       []ShipInfo        // корабли на планетах
    Artefacts   []ArtefactInfo    // активные артефакты (с оставшимся временем)
    EPoints     int               // боевой опыт
    Battles     int               // кол-во боёв
}
```

Форматируем в текст для prompt:
```
Игрок: [username], место в рейтинге: #42, очки: 15000, боевой опыт: 500, боёв: 23

Планеты (2):
- "Главная" [1:5:3], тип: normaltempplanet, диаметр: 12500
  Металл: 50000/500000, Кремний: 30000/400000, Водород: 10000
  Здания: Шахта М(8, factor=80%), Шахта Si(6), Лаборатория(5), Верфь(4)
- "Колония" [2:7:8], металл 20000, кремний 15000

Исследования: Оружие(5), Щиты(4), Броня(3), Компьютер(6)

Флот на Главной: 10 Истребителей, 5 Крейсеров, 2 Транспорта
Активные миссии: Атака на [3:2:1] (прибытие через 12 мин)

Активные артефакты: Catalyst (+10% производство, осталось 2ч)
```

**Что не включать** (слишком дорого / бесполезно):
- Полную историю боёв (достаточно "23 боя, 500 опыта")
- Всю галактику (только флоты в полёте игрока)
- Архив сообщений чата
- Данные других игроков (только рейтинг если спрашивает про атаку)

### 4. System prompt — структура и источники знаний

System prompt делится на **две части** с разным TTL кэша:

#### Кэшируемая часть (статическая, из `configs/`)

Вкладывается один раз и кэшируется через [prompt caching](https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching) — 90% скидка на повторные запросы.

ИИ **не знает формулы наизусть** — они вшиваются из актуальных конфигов при старте сервера. Если изменится баланс в YAML — советник автоматически даёт правильные цифры.

**Из `configs/buildings.yml`:**
```
Шахта Металла ур.N: производство = 30 * N * 1.1^N * factor/100
Верфь ур.N: скорость постройки кораблей × (1 + N*0.1)
Лаборатория ур.N: разблокирует исследования до ур.N
```

**Из `configs/artefacts.yml`:**
```
Catalyst: +10% к производству всех ресурсов, длительность 24ч, стоит 5000М+2000Si
Aegis: +20% к щитам флота в бою
```

**Из `configs/units.yml`:**
```
Истребитель: атака 50, щит 10, броня 100, стоит 2000М/500Si
Крейсер: атака 400, щит 50, броня 700, стоит 20000М/7000Si/2000H
```

**Боевая механика (из oxsar-spec.txt §14):**
```
Бой: раунды, каждый раунд флоты атакуют друг друга
Урон = атака * AttackMul (от артефактов battle_bonus)
Поглощение = щит * ShieldMul
Прочность = броня * ShellMul
```

#### Динамическая часть (меняется каждый запрос)

Текущее состояние конкретного игрока — результат `buildPlayerContext()`.

#### Итоговая структура запроса

```
[КЭШИРУЕМАЯ — редко меняется, ~1500 токенов]
Ты — ИИ-советник в космической стратегии Oxsar.
Отвечай на русском языке, дружелюбно и кратко (2-5 предложений).
Давай конкретные советы, основанные на состоянии игрока.
Отвечай только на вопросы об игре. Если вопрос не по теме — вежливо откажи.

Правила игры, формулы и каталог (из configs/):
[buildStaticGameKnowledge() — вставляется при старте сервера]

[ДИНАМИЧЕСКАЯ — каждый запрос, ~500-800 токенов]
Текущее состояние игрока:
[buildPlayerContext() — вставляется перед каждым запросом]

[ВОПРОС ИГРОКА]
Что мне строить дальше?
```

### 5. Rate limiting

Таблица в БД:
```sql
CREATE TABLE ai_advisor_log (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id),
    question    TEXT NOT NULL,
    answer      TEXT NOT NULL,
    tokens_used INT,
    created_at  TIMESTAMPTZ DEFAULT now()
);
```

Проверка:
```go
func (s *Service) checkRateLimit(ctx context.Context, userID string) error {
    var count int
    s.db.Pool().QueryRow(ctx, `
        SELECT COUNT(*) FROM ai_advisor_log
        WHERE user_id = $1 AND created_at > now() - interval '1 day'
    `, userID).Scan(&count)
    if count >= s.maxPerDay {
        return ErrRateLimitExceeded
    }
    return nil
}
```

### 6. Backend API endpoint

```go
// POST /api/ai-advisor/ask
// Body: { "question": "Что мне строить?" }
// Response: { "answer": "...", "questions_left": 18 }
```

### 7. Frontend — AIAdvisorWidget

Небольшой виджет в UI (не отдельная страница, а floating кнопка или панель):

```tsx
// frontend/src/features/ai-advisor/AIAdvisorWidget.tsx

export function AIAdvisorWidget() {
  const [open, setOpen] = useState(false);
  const [question, setQuestion] = useState('');
  const [answer, setAnswer] = useState('');

  const mutation = useMutation({
    mutationFn: (q: string) =>
      api.post<{ answer: string; questions_left: number }>(
        '/api/ai-advisor/ask', { question: q }
      ),
    onSuccess: (data) => setAnswer(data.answer),
  });

  return (
    <>
      {/* Floating кнопка */}
      <button onClick={() => setOpen(true)}>🤖 Советник</button>

      {open && (
        <div className="ox-panel">
          <h3>🤖 AI Советник</h3>
          <p style={{ color: 'var(--ox-fg-dim)' }}>
            Задай вопрос об игре — отвечу с учётом твоего состояния.
          </p>
          <textarea
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            placeholder="Что мне строить дальше?"
          />
          <button onClick={() => mutation.mutate(question)}>
            {mutation.isPending ? 'Думаю...' : 'Спросить'}
          </button>
          {answer && <div className="ox-panel">{answer}</div>}
        </div>
      )}
    </>
  );
}
```

Добавить в `App.tsx` рядом с навигацией.

---

## Альтернативный backend: Ollama (локальная LLM)

Ollama — локальный сервер с открытыми моделями. Поднимается рядом с игровым сервером,
не требует внешнего API-ключа, для игроков бесплатно.

### Рекомендуемые модели (по качеству русского языка)

| Модель | Размер | RAM | Качество русского | Скорость |
|--------|--------|-----|-------------------|---------|
| `qwen2.5:3b` | 2 GB | 3 GB | ✅ хорошо | ~2-3 сек |
| `qwen2.5:7b` | 4 GB | 6 GB | ✅ отлично | ~5 сек |
| `llama3.2:3b` | 2 GB | 3 GB | ⚠️ средне | ~2 сек |
| `mistral:7b` | 4 GB | 5 GB | ⚠️ средне | ~5 сек |
| `gemma2:2b` | 2 GB | 3 GB | ❌ слабо | ~2 сек |

**Дефолт: `qwen2.5:3b`** — Qwen (Alibaba) обучался на китайском + русском, поэтому
среди малых моделей русский лучший. 3b-версия влезает даже на бюджетный VPS.

### Рекомендации по VPS

| VPS | RAM | Диск | Рекомендация | Конфиг |
|-----|-----|------|-------------|--------|
| **Бюджетный** (Hetzner CX21, DO Basic) | 4 GB | 40 GB | `qwen2.5:3b` | `OLLAMA_MODEL=qwen2.5:3b` |
| **Средний** (Hetzner CX31, DO General) | 8 GB | 80 GB | `qwen2.5:7b` | `OLLAMA_MODEL=qwen2.5:7b` |
| **Мощный** (Hetzner CX41+, выделенный) | 16+ GB | 160+ GB | `qwen2.5:7b` или Claude API | на усмотрение |
| **Нет Ollama** (любой, только Claude API) | любой | любой | Claude API | `OLLAMA_URL` не задан |

**Важно по дискам:** модели занимают место на диске (2-4 GB), плюс сам Ollama ~500 MB.
На VPS с 20 GB диском места хватит, но нужно следить.

**CPU vs GPU:** Ollama работает на CPU без GPU — скорость ~2-5 сек на ответ.
С GPU (RTX 3060+) — менее 1 сек, но на VPS GPU редко доступен.

**Нагрузка:** Ollama обрабатывает запросы последовательно. При одновременных
запросах от нескольких игроков — очередь. Для casual-игры (не более 5-10 запросов
в минуту) достаточно одного инстанса на CPU.

### Запуск Ollama

```bash
# Установка на Linux (Ubuntu/Debian)
curl -fsSL https://ollama.com/install.sh | sh

# Скачать модель (выбрать одну по RAM)
ollama pull qwen2.5:3b   # для VPS с 4 GB RAM
ollama pull qwen2.5:7b   # для VPS с 8+ GB RAM

# Проверка
curl http://localhost:11434/api/tags

# Ollama слушает на localhost:11434
# В docker-compose добавить как сервис или запускать systemd-сервисом
```

```yaml
# docker-compose.yml — добавить сервис (опционально):
ollama:
  image: ollama/ollama
  ports:
    - "11434:11434"
  volumes:
    - ollama_data:/root/.ollama
  deploy:
    resources:
      limits:
        memory: 6g   # подобрать под модель

volumes:
  ollama_data:
```

### Интеграция в Go — интерфейс LLMClient

Чтобы сервис мог переключаться между Claude и Ollama, вводим интерфейс:

```go
// backend/internal/aiadvisor/llm.go

type LLMClient interface {
    Complete(ctx context.Context, systemPrompt, userMessage string) (string, error)
}
```

**Реализация для Claude API** (`client.go` — уже описан выше).

**Реализация для Ollama:**

```go
// backend/internal/aiadvisor/ollama.go

type OllamaClient struct {
    baseURL string        // "http://localhost:11434"
    model   string        // "qwen2.5:3b" (дефолт) или "qwen2.5:7b"
    http    *http.Client
}

func (c *OllamaClient) Complete(ctx context.Context, system, user string) (string, error) {
    body, _ := json.Marshal(map[string]any{
        "model":  c.model,
        "stream": false,
        "messages": []map[string]string{
            {"role": "system", "content": system},
            {"role": "user", "content": user},
        },
    })

    req, _ := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+"/api/chat", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.http.Do(req)
    if err != nil {
        return "", fmt.Errorf("ollama: %w", err)
    }
    defer resp.Body.Close()

    var result struct {
        Message struct {
            Content string `json:"content"`
        } `json:"message"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("ollama decode: %w", err)
    }
    return result.Message.Content, nil
}
```

### Выбор backend через конфиг

```go
// В AIAdvisorConfig добавить:
type AIAdvisorConfig struct {
    APIKey      string  // ANTHROPIC_API_KEY
    ProxyURL    string  // ANTHROPIC_PROXY_URL
    OllamaURL   string  // OLLAMA_URL (напр. "http://localhost:11434")
    OllamaModel string  // OLLAMA_MODEL (напр. "qwen2.5:3b"), default "qwen2.5:3b"
    MaxPerDay   int
    MaxTokens   int
}
```

```go
// В main.go — выбор реализации:
var llm aiadvisor.LLMClient
if cfg.AIAdvisor.OllamaURL != "" {
    llm = aiadvisor.NewOllamaClient(cfg.AIAdvisor.OllamaURL, cfg.AIAdvisor.OllamaModel)
} else {
    llm = aiadvisor.NewAnthropicClient(cfg.AIAdvisor)
}
svc := aiadvisor.NewService(db, catalog, llm, cfg.AIAdvisor)
```

В `.env`:
```env
# Режим Claude (по умолчанию):
ANTHROPIC_API_KEY=sk-ant-...

# Режим Ollama (если задан OLLAMA_URL — Claude не используется):
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=qwen2.5:7b
```

### Стоимость для игрока при Ollama

Если сервер работает в режиме Ollama — все вопросы **бесплатны** (0 кредитов).
Выбор модели (Haiku/Sonnet/Opus) в UI скрывается, так как неактуален.

```go
func (s *Service) EstimateCost(model string, questionLen int) int {
    if s.isOllama {
        return 0  // локальный режим — бесплатно
    }
    // ... обычная логика кредитов
}
```

### Ограничения Ollama

- Требует 8-16 GB RAM на сервере (зависит от модели)
- Генерация занимает 3-10 секунд (Claude — 1-3 сек)
- Качество ответов на русском заметно ниже Claude Sonnet/Opus
- Один инстанс обслуживает запросы последовательно — при нагрузке очередь

---

## Стоимость и ограничения

| Параметр | Значение |
|----------|---------|
| Модель | claude-sonnet-4-6 |
| Токены на ответ | ~1000 (prompt) + ~300 (ответ) ≈ 1300 токенов |
| Цена за 1M токенов | ~$3 input / ~$15 output |
| Цена за 1 вопрос | ~$0.007 (0.7 цента) |
| Лимит | 20 вопросов/игрок/день |
| При 100 активных игроках | ~$14/день |

**Рекомендации по снижению стоимости:**
- Кэшировать system prompt через [prompt caching](https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching) (скидка 90% на повторные запросы)
- Сократить контекст игрока до минимально необходимого
- Rate limit (20 вопросов/день — разумный предел)

---

## Порядок разработки

1. **Backend (M6):**
   - [ ] Добавить Anthropic SDK (`go get`)
   - [ ] Создать `backend/internal/aiadvisor/service.go`
   - [ ] Миграция `ai_advisor_log` таблицы
   - [ ] Endpoint `POST /api/ai-advisor/ask`
   - [ ] Rate limiting
   - [ ] `ANTHROPIC_API_KEY` в `.env` и docker-compose

2. **Frontend (M6):**
   - [ ] Создать `AIAdvisorWidget.tsx`
   - [ ] Добавить floating кнопку в `App.tsx`
   - [ ] Показывать `questions_left` (лимит)
   - [ ] Loading + error states

3. **Тестирование:**
   - [ ] Проверить качество ответов на 10-15 тестовых вопросов
   - [ ] Проверить rate limiting
   - [ ] Проверить prompt caching (снижение стоимости)

---

## Файлы в nova

| Файл | Статус | Описание |
|------|--------|---------|
| `backend/internal/aiadvisor/llm.go` | 🆕 | Интерфейс `LLMClient` |
| `backend/internal/aiadvisor/client.go` | 🆕 | Реализация для Claude API (с прокси) |
| `backend/internal/aiadvisor/ollama.go` | 🆕 | Реализация для Ollama (локальная LLM) |
| `backend/internal/aiadvisor/service.go` | 🆕 | Основная логика Ask, EstimateCost, rate limit |
| `backend/internal/aiadvisor/handler.go` | 🆕 | HTTP endpoints (estimate + ask) |
| `backend/internal/aiadvisor/context.go` | 🆕 | buildPlayerContext + buildStaticGameKnowledge |
| `migrations/XXXX_ai_advisor_log.sql` | 🆕 | Rate limit таблица |
| `frontend/src/features/ai-advisor/AIAdvisorWidget.tsx` | 🆕 | UI виджет с выбором модели и превью стоимости |
| `backend/internal/config/config.go` | ✏️ | Добавить AIAdvisorConfig (APIKey, ProxyURL, OllamaURL, …) |
| `backend/cmd/server/main.go` | ✏️ | Выбор LLMClient (Claude vs Ollama), регистрация роута |
| `.env` / `docker-compose.yml` | ✏️ | ANTHROPIC_API_KEY, ANTHROPIC_PROXY_URL, OLLAMA_URL, OLLAMA_MODEL |
