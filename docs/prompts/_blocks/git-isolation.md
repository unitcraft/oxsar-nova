# Блок: GIT-ИЗОЛЯЦИЯ от параллельных сессий (включи в промпт)

**Источник:** memory `feedback_parallel_session_check.md` + 4 прецедента
захвата чужих файлов за 2026-04-27/28.

**Применение в промпте:** скопируй секцию ниже целиком в раздел
«GIT-ИЗОЛЯЦИЯ» continuation/initial-промпта. Адаптируй список путей
под конкретный план.

---

```
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ (4 прецедента граблей в memory):

Параллельно с тобой могут работать 2-4 других Claude Code агента
на этом же репозитории. Без явной изоляции — захват чужих файлов
в твой commit гарантирован.

ОБЯЗАТЕЛЬНЫЙ АЛГОРИТМ:

1. ПРИ СТАРТЕ:
   - git status --short — посмотри текущее состояние.
   - cat docs/active-sessions.md — посмотри какие агенты сейчас
     активны и какие файлы трогают.
   - Если в active-sessions.md свободный slot — добавь свой:
     | C | План X Ф.Y | path1/ path2/ | <дата-время> | feat(...): ... |

2. ПЕРЕД git add:
   - git status --short — повторно, может изменилось.
   - Сверь со своим списком путей. Если чужие файлы пересекаются
     с твоими — стоп, спроси пользователя.

3. git add ТОЛЬКО ПОИМЁННО:
   - НИКОГДА git add . / git add -A / git add -u.
   - ВСЕГДА git add path1 path2 path3 — явный список.

4. ПЕРЕД git commit:
   - git diff --cached --name-only — финальная проверка.
   - Если в staged видишь файлы которые не твои — git reset HEAD --
     <чужой-путь>.

5. git commit ВСЕГДА С ДВОЙНЫМ ТИРЕ:
   git commit -m "..." -- path1 path2 path3

   ЭТО ЖЁСТКОЕ ПРАВИЛО. Без `--` git подберёт всё из индекса,
   включая чужие staged-файлы которые могла добавить параллельная
   сессия между твоими git status и git commit.

6. ОПЦИОНАЛЬНАЯ ДОПОЛНИТЕЛЬНАЯ ЗАЩИТА (pre-commit hook):
   export CC_AGENT_PATHS="path1 path2 path3"
   git add $CC_AGENT_PATHS
   git commit -m "..." -- $CC_AGENT_PATHS

   Если установлен pre-commit hook (scripts/git-hooks/pre-commit) —
   он заблокирует commit если в staged попало что-то вне CC_AGENT_PATHS.

7. ПОСЛЕ commit:
   - Удали свою строку из docs/active-sessions.md.
   - Это сигнал другим агентам что slot свободен.

ЕСЛИ СЛУЧАЙНО ЗАХВАТИЛ ЧУЖОЙ КОММИТ:
- git revert HEAD --no-edit — создать revert-коммит.
- git checkout HASH-захвата -- chuжой-путь1 chuжой-путь2 ... —
  восстановить чужие файлы из удалённого коммита.
- git reset HEAD -- chuжой-путь1 ... — вернуть в untracked, как
  было до твоего fuckup'а.
- Затем сделай чистый повторный коммит ТОЛЬКО своих файлов.

См. прецедент 2026-04-28 #4 в memory feedback_parallel_session_check.md.
```
