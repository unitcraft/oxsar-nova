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
