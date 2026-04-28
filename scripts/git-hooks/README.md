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

Удаляет из коммит-сообщений строки вида
`Co-Authored-By: Claude <noreply@anthropic.com>`.

**Зачем:** план 41 §6 разделил юридическую и техническую плоскости
упоминания AI. Trailer `Co-Authored-By` — Git-стандарт соавторства,
GitHub отображает Claude как контрибьютора. Для стратегии «AI как
инструмент» нужна нейтральная метка (`Generated-with: Claude Code`)
или отсутствие AI-trailer'а вообще.

AI-сессии иногда вписывают `Co-Authored-By: Claude` руками в HEREDOC
коммит-сообщения, обходя настройку `attribution.commit` в
`.claude/settings.json`. Этот hook — единая страховка на уровне Git,
работающая независимо от состояния AI-сессии.

**Что НЕ удаляется:**
- `Co-Authored-By: <человек>@<домен>` для реальных соавторов;
- `Generated-with: Claude Code` (правильный AI-trailer);
- любые другие trailer'ы.

**Удаляется только:**
- `Co-Authored-By: Claude <noreply@anthropic.com>`
- `Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>`

**Связанные документы:**
- `docs/plans/41-origin-rights.md` §6 — обоснование разделения trailer'ов.
- `docs/origin-rights.md` §6 — публичное объяснение для коммитов до
  разделительного `1bae5ee3fd`.
- `docs/ops/claude-code-attribution.md` — как настроить
  `attribution.commit` в проектном `.claude/settings.json`.
- memory `feedback_commits.md` — правило для AI-сессий.

## Проверка

После коммита:

```bash
git log -1 --format=%B
```

В выводе должен быть **либо** `Generated-with: Claude Code` (идеально),
**либо** отсутствовать AI-trailer вообще (приемлемо). Не должно быть
`Co-Authored-By: Claude` — если есть, значит hook не активирован
(`git config --get core.hooksPath` должно вернуть `scripts/git-hooks`).
