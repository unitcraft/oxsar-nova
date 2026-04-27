# Атрибуция Claude Code в коммитах oxsar-nova

Если вы используете Claude Code (или другой агент Anthropic) для работы
с этим репозиторием, настройте проектный `.claude/settings.json` так,
чтобы коммиты получали trailer `Generated-with: Claude Code` вместо
дефолтного `Co-Authored-By: Claude <noreply@anthropic.com>`.

## Зачем

Это согласуется с юридической позицией проекта (см.
[docs/origin-rights.md](../origin-rights.md) §6, план 41): AI-ассистент —
**технический инструмент**, не соавтор. Trailer `Co-Authored-By` —
Git-стандарт соавторства, GitHub отображает Claude как контрибьютора и
засчитывает в статистику. Произвольный `Generated-with:` Git и GitHub
не интерпретируют как соавторство.

## Как настроить

### Минимум: атрибуция через `attribution.commit`

В файле `.claude/settings.json` (он в `.gitignore`, личный для каждого
разработчика):

```json
{
  "permissions": {
    "allow": ["Bash(*)", "Edit(**)", "Write(**)", "Read(**)"]
  },
  "attribution": {
    "commit": "Generated-with: Claude Code",
    "pr": ""
  }
}
```

Поле `permissions` — на ваше усмотрение, оно к атрибуции отношения не
имеет; пример выше — широкие разрешения для удобства, можно сужать.

**Известная проблема (CC 2.1.x, 2026-04):** настройка `attribution.commit`
иногда не подхватывается автоматически — коммиты идут без AI-trailer'а
(приемлемо по правилу выше) или с `Co-Authored-By: Claude` (нарушение,
если AI вписал руками). Поэтому ниже рекомендуется второй слой защиты —
`SessionStart`-hook + git-hook.

### Полная автоматизация: SessionStart-hook + git-hook

Расширенный `.claude/settings.json` для каждой сессии автоматически
активирует расшаренный git-hook `scripts/git-hooks/commit-msg`, который
**гарантированно** срезает `Co-Authored-By: Claude` из коммит-сообщения
независимо от того, что написано в HEREDOC:

```json
{
  "permissions": {
    "allow": ["Bash(*)", "Edit(**)", "Write(**)", "Read(**)"]
  },
  "attribution": {
    "commit": "Generated-with: Claude Code",
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

После этого:
- `attribution.commit` пытается вписать `Generated-with: Claude Code` —
  идеальный исход;
- если CLI его не подхватил, и AI вписал `Co-Authored-By: Claude`
  руками в HEREDOC — git-hook это срежет автоматически на
  `git commit`;
- если AI ничего не вписал — коммит идёт чисто, без AI-trailer
  (тоже приемлемо по правилу плана 41 §6).

### Альтернатива без settings.json

Если разработчик не хочет создавать `.claude/settings.json`, можно
активировать только git-hook вручную одной командой:

```bash
git config core.hooksPath scripts/git-hooks
```

Это инструкция для всех новых сессий и разработчиков, описана в
[CLAUDE.md](../../CLAUDE.md) §«Onboarding: одноразовые настройки».

## Проверка

После настройки сделайте тестовый коммит и посмотрите:

```bash
git log -1 --format=%B
```

В сообщении должен быть trailer `Generated-with: Claude Code`. Если
вместо него стоит `Co-Authored-By: Claude` — настройка не подхвачена,
проверьте версию Claude Code и путь к проектному settings.json.

## Прошлая история коммитов

В коммитах до плана 41 (2026-04-26) встречается trailer
`Co-Authored-By: Claude <noreply@anthropic.com>`. Это **техническая
метка** процесса использования AI, не юридическое заявление о
соавторстве — см. [docs/origin-rights.md](../origin-rights.md) §6.
Историю не переписываем; разъяснение в `docs/origin-rights.md` имеет
ту же юридическую силу, что и удаление трейлера.
