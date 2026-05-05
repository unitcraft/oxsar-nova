package aiadvisor

// Файл содержит типы и константы rule-based автопилота (план 06.1).
//
// Автопилот — детерминированный советник, который видит всё состояние игрока
// и все формулы экономики/боя. На основе этого он предлагает топ-3 ходов
// (стратегия выбирается игроком). Игрок подтверждает один — и backend
// создаёт игровое событие через стандартный сервисный метод
// (building.Enqueue / research.Enqueue / fleet.Send / ...).
//
// Расчёт топ-3 — потенциально тяжёлая операция (snapshot всего состояния +
// scoring всех кандидатов), поэтому выполняется в воркере по событию
// event.KindAutopilotAdvise. Результат пишется в ai_advisor_log
// (status='ready', action_params=JSONB с топ-3) и доставляется на фронт
// либо WS-пушем (Ф.6), либо polling-ом GET /result/{jobID}.

// Strategy — целевая функция оптимизации. Хранится в payload event'а
// и в ai_advisor_log.strategy.
type Strategy string

const (
	StrategyEconomy   Strategy = "economy"   // Максимум прироста производства ресурсов
	StrategyMilitary  Strategy = "military"  // Максимум боевой мощи флота
	StrategyDefense   Strategy = "defense"   // Максимум оборонных установок и защитных технологий
	StrategyExpansion Strategy = "expansion" // Скорейшая колонизация
	StrategyAuto      Strategy = "auto"      // Эвристика фазы развития
)

// Validate возвращает true если значение — известная стратегия.
func (s Strategy) Validate() bool {
	switch s {
	case StrategyEconomy, StrategyMilitary, StrategyDefense, StrategyExpansion, StrategyAuto:
		return true
	}
	return false
}

// AutopilotStatus — жизненный цикл записи в ai_advisor_log для autopilot.
const (
	AutopilotStatusPending  = "pending"  // event поставлен, воркер ещё не считал
	AutopilotStatusReady    = "ready"    // топ-3 рассчитан, ожидает выбора игрока
	AutopilotStatusExecuted = "executed" // игрок выбрал и подтвердил, событие создано
	AutopilotStatusExpired  = "expired"  // TTL истёк (1 час с ready)
)

// AutopilotSource — значение колонки ai_advisor_log.source для записей автопилота.
const AutopilotSource = "autopilot"

// AdviseInput — payload event'а KindAutopilotAdvise.
// Сериализуется в events.payload (JSONB).
type AdviseInput struct {
	UserID   string   `json:"user_id"`
	Strategy Strategy `json:"strategy"`
	LogID    string   `json:"log_id"` // id записи в ai_advisor_log
}

// Recommendation — одна карточка топ-3 в action_params.
//
// Поля Description и Benefit — человекочитаемые строки для UI;
// все локализованные тексты собираются на сервере (строгое i18n не
// нужно, текст формируется из YAML-каталога юнитов).
type Recommendation struct {
	ID          string         `json:"id"`          // uuid рекомендации, передаётся в Execute
	Category    string         `json:"category"`    // "building" | "research" | "mission" | "defense" | ...
	ActionType  string         `json:"action_type"` // конкретный подтип, например "build_metalmine"
	PlanetID    string         `json:"planet_id,omitempty"`
	UnitID      int            `json:"unit_id,omitempty"`
	Params      map[string]any `json:"params,omitempty"`
	Score       float64        `json:"score"`       // числовая оценка выгоды (внутренняя)
	Description string         `json:"description"` // что предлагается ("Построить Металлошахту ур.5")
	Benefit     string         `json:"benefit"`     // ожидаемая выгода ("+120 очков за 4ч 20мин")
}

// AutopilotResult — то, что лежит в action_params (JSONB) и возвращается
// фронту в ответе на /advise polling и WS-пуше.
type AutopilotResult struct {
	JobID    string           `json:"job_id"`
	Status   string           `json:"status"`
	Strategy Strategy         `json:"strategy"`
	Recs     []Recommendation `json:"recs"`
}

// EnqueueResult — ответ POST /advise (job_id для последующего polling).
type EnqueueResult struct {
	JobID string `json:"job_id"`
}

// ExecuteResult — ответ POST /execute (id созданного игрового события).
type ExecuteResult struct {
	EventID string `json:"event_id"`
}
