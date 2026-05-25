# Настройка Claude Code для работы с oxsar-nova

Политика репозитория: коммит-сообщения не содержат vendor-атрибуции
AI-инструментов. Принудительно: git-hook
[scripts/git-hooks/commit-msg](../../scripts/git-hooks/commit-msg)
срезает соответствующие trailer'ы перед записью коммита. Реальные
`Co-Authored-By` для живых соавторов сохраняются.

Обоснование — [docs/origin-rights.md](../origin-rights.md) §6.
Лицензионные условия Claude Code / Anthropic API не требуют указания
атрибуции вендора в выходных артефактах, поэтому такая политика
полностью соответствует условиям использования.

## Как настроить Claude Code

В файле `.claude/settings.json` (он в `.gitignore`, личный для каждого
разработчика):

```json
{
  "permissions": {
    "allow": ["Bash(*)", "Edit(**)", "Write(**)", "Read(**)"]
  },
  "attribution": {
    "commit": "",
    "pr": ""
  },
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "git config core.hooksPath scripts/git-hooks"
          }
        ]
      }
    ]
  }
}
```

Поле `permissions` — на ваше усмотрение, к атрибуции отношения не
имеет. Пустые `attribution.commit` и `attribution.pr` отключают
автоматическое добавление AI-меток. `SessionStart`-hook автоматически
активирует git-hook (`scripts/git-hooks/commit-msg`), который —
гарантированная страховка: даже если будущая версия CLI начнёт
вписывать что-то своё или если AI впишет trailer руками в HEREDOC,
hook это срежет.

## Альтернатива без settings.json

Активировать только git-hook вручную (одной командой, идемпотентно):

```bash
git config core.hooksPath scripts/git-hooks
```

Эта команда описана в [CLAUDE.md](../../CLAUDE.md) §«Onboarding:
одноразовые настройки».

## Проверка

После любого коммита:

```bash
git log -1 --format=%B
```

В сообщении не должно быть AI/vendor-trailer'ов. Если что-то такое
осталось — hook не активирован (`git config --get core.hooksPath`
должно вернуть `scripts/git-hooks`).
