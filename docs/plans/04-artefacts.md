# План D: Артефакты

---

## История (завершено)

### ✅ Базовая система артефактов (итерация ~25)
- `artefacts_user` таблица, состояния: held → active → expired
- `ArtefactSpec`: effect type, stackable, lifetime_seconds
- `computeChanges`: применение эффектов к полям планеты (factor_metal, factor_silicon и т.д.)
- `KindArtefactExpire` event — автоматическое истечение
- UI: ArtefactsScreen, список held/active с кнопками активации

### ✅ Рынок артефактов (итерация ~35)
- `artefact_market` таблица, CreateOffer / BuyOffer / CancelOffer
- UI: ArtmarketScreen с фильтром «мои/чужие», inline price input
- Idempotency key для BuyOffer

### ✅ Артефакты из экспедиций (итерация ~30, улучшено в план 22)
- `expArtefact` в `fleet/expedition.go`: случайный артефакт → state=held
- Вес artefact в weighted-random зависит от exp_power (план 22)

---

## Открытые задачи

### ✅ D.1 Лимит стеккинга

- `ArtefactSpec.MaxStacks int` в catalog.go
- `catalyst.max_stacks: 3`, `power_generator.max_stacks: 2` в artefacts.yml
- `ErrMaxStacksReached` в service.go; проверка при активации; handler → 400
- Тесты: `TestMaxStacks_SpecField`, `TestErrMaxStacksReached_IsSentinel`

---

### ✅ D.2 Battle bonus артефакты — верификация ID (план 01, итерация ~48)
ID верифицированы по `oxsar2/sql/data.sql` (не consts.php — в нём этих констант нет):
- battle_attack_power: id=318, battle_shell_power: id=316, battle_shield_power: id=317
- _10 версии: id=359-361
- artefacts.yml обновлён, A.2 реализована

---

### ✅ D.3 Реферальная система

- `migrations/0047_referral.sql`: `users.referred_by TEXT REFERENCES users(id)` + индекс
- `referral/service.go`: `ProcessRegistration` (referred_by + стартовые ресурсы 10/5/2 + 3000 очков рефереру),
  `ProcessPurchase` (20% кредитов рефереру)
- `auth/service.go`: `ReferralProcessor` интерфейс + `WithReferral()`; `RegisterInput.ReferredBy`
- `auth/handler.go`: `POST /api/auth/register?ref=<userID>` передаёт `ReferredBy`
- `cmd/server/main.go`: `referral.NewService(db)` подключён через `WithReferral`
- `ProcessPurchase` вызывается из payment webhook (план F.1)
- Тесты: константы, стартовые ресурсы, bonus calc
