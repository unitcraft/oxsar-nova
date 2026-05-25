# Git hooks для oxsar-nova

Кросс-разработческие git-hooks. Активируются разово на каждой машине:

```bash
git config core.hooksPath scripts/git-hooks
```

После этого Git будет использовать hook'и из этой папки вместо
стандартных `.git/hooks/`.

## Хуки

### `pre-commit`

Защита от случайного захвата чужих файлов в commit при параллельных
Claude Code сессиях.

**Зачем:** 4 раза за 2 дня (2026-04-27/28) у Claude Code агентов
случайно попадали в commit чужие staged-файлы — потому что между
`git add path` и `git commit` другая параллельная сессия успевала
что-то staged'ить, а commit без двойного-тире подбирает всё из
индекса. См. memory `feedback_parallel_session_check.md`.

**Логика:** если переменная среды `CC_AGENT_PATHS` установлена
(агент сам выставляет её при старте сессии), hook проверяет что
в commit идут только эти пути. Если попадается чужой файл — commit
блокируется с понятным сообщением.

**Использование агентом:**

```bash
export CC_AGENT_PATHS="internal/billing/client/ pkg/idempotency/ docs/plans/77-..."
git add internal/billing/client/...
git commit -m "..." -- internal/billing/client/ pkg/idempotency/ docs/plans/77-...
```

**Backwards-compat:** если `CC_AGENT_PATHS` не задана — hook просто
проходит, не блокирует. Это для ручных коммитов людей и legacy-сессий.

**Связанные документы:**
- memory `feedback_parallel_session_check.md` — правило 3:
  ВСЕГДА `git commit -m "..." -- path1 path2`.
- `docs/active-sessions.md` — лайв-документ для координации между
  параллельными агентами.

### `commit-msg`

Срезает из коммит-сообщений автоматические AI-атрибуционные
trailer'ы. Реальные `Co-Authored-By: <человек>@<домен>` для живых
соавторов сохраняются.

**Зачем:** политика репо — коммиты oxsar-nova не несут vendor- и
AI-меток. Hook — единая страховка на уровне Git, работающая
независимо от состояния AI-сессии и её настроек.

**Связанные документы:**
- `docs/origin-rights.md` §6 — политика атрибуции в коммитах.
- `docs/ops/claude-code-attribution.md` — настройка для AI-сессий.
- memory `feedback_commits.md` — правило для AI-сессий.

## Проверка

После коммита:

```bash
git log -1 --format=%B
```

В сообщении не должно быть AI/vendor-trailer'ов. Если что-то такое
осталось — hook не активирован (`git config --get core.hooksPath`
должно вернуть `scripts/git-hooks`).
