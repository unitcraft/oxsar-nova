#!/usr/bin/env bash
#
# update-baselines.sh — обновить эталоны (намеренно перезаписать существующие).
# План 73 Ф.1+Ф.2 §«Регламент».
#
# Используется только при намеренном изменении legacy-UI (новые фичи в legacy
# или сознательная правка стилей). Технические правки (refactor) НЕ должны
# менять baseline — diff в Ф.3 должен оставаться 0%.
#
# Скрипт делает то же что take-screenshots.sh, но с явным confirmation prompt
# и подсказкой описать изменение в коммит-сообщении.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
BASELINE_DIR="$REPO_ROOT/tests/e2e/origin-baseline"

cat <<EOF
ВНИМАНИЕ: вы собираетесь ПЕРЕПИСАТЬ baseline-скриншоты в:
  $BASELINE_DIR/screenshots/

Это допустимо только при НАМЕРЕННОМ изменении legacy-UI. Технические правки
(refactor, чистка) НЕ должны менять baseline.

В коммит-сообщении явно укажите:
  - Что именно изменилось в legacy-UI (страница, блок).
  - Почему (issue / план / решение).
  - Ссылку на release notes если применимо (см. план 73 §«Регламент»).

EOF

read -r -p "Продолжить? (yes/no): " ANSWER
if [[ "$ANSWER" != "yes" ]]; then
  echo "отмена"
  exit 1
fi

# Делегируем take-screenshots.sh с force-флагом (он перезаписывает PNG).
SMOKE="${SMOKE:-0}" bash "$BASELINE_DIR/take-screenshots.sh"

cat <<EOF

Готово. Дальше:
  1. git diff --stat $BASELINE_DIR/screenshots/   # убедиться какие PNG изменились
  2. git add только намеренно обновлённые
  3. git commit -m "feat(e2e): обновить baseline после <кратко>" \\
       -- $BASELINE_DIR/screenshots/<files>...
EOF
