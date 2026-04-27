# Git hooks для oxsar-nova

Кросс-разработческие git-hooks. Активируются разово на каждой машине:

```bash
git config core.hooksPath scripts/git-hooks
```

После этого Git будет использовать hook'и из этой папки вместо
стандартных `.git/hooks/`.

## Хуки

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
