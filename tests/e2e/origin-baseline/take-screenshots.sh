#!/usr/bin/env bash
#
# take-screenshots.sh — снять baseline-скриншоты origin-экранов из legacy-php.
# План 73 Ф.1+Ф.2.
#
# Что делает:
#   1. Поднимает legacy-php docker-стек (mysql + memcached + php + nginx + worker).
#   2. Ждёт пока nginx ответит 200 на /dev-login.php (max 60s).
#   3. Запускает Playwright spec — он логинится через /dev-login.php и снимает
#      PNG в tests/e2e/origin-baseline/screenshots/.
#   4. Останавливает (НЕ down) docker-стек, чтобы повторный запуск был быстрым.
#
# Окружение:
#   SMOKE=1 (default) — снять только smoke-набор (7 экранов).
#   SMOKE=0           — снять все 22 экрана (Spring 1+2).
#   LEGACY_URL        — URL legacy-php (default http://localhost:8092).
#   KEEP_RUNNING=1    — не останавливать docker-стек после снятия.
#
# Использование:
#   bash tests/e2e/origin-baseline/take-screenshots.sh
#   SMOKE=0 bash tests/e2e/origin-baseline/take-screenshots.sh   # all 22

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
BASELINE_DIR="$REPO_ROOT/tests/e2e/origin-baseline"
LEGACY_DOCKER_DIR="$REPO_ROOT/projects/game-legacy-php/docker"
LEGACY_URL="${LEGACY_URL:-http://localhost:8092}"

cd "$BASELINE_DIR"

echo "==> 1/4: подъём legacy-php docker-стека ($LEGACY_DOCKER_DIR)"
( cd "$LEGACY_DOCKER_DIR" && docker compose up -d )

echo "==> 2/4: ожидание health (legacy-php /dev-login.php → 2xx/3xx)"
ATTEMPTS=0
MAX_ATTEMPTS=30  # 30 × 2s = 60s
until code=$(curl -s -o /dev/null -w '%{http_code}' "$LEGACY_URL/dev-login.php" 2>/dev/null || echo 000); \
      [[ "$code" =~ ^[23] ]]; do
  ATTEMPTS=$((ATTEMPTS+1))
  if (( ATTEMPTS >= MAX_ATTEMPTS )); then
    echo "ОШИБКА: legacy-php не ответил за 60s (last code=$code)" >&2
    echo "  Логи: cd $LEGACY_DOCKER_DIR && docker compose logs --tail=100" >&2
    exit 1
  fi
  sleep 2
done
echo "    legacy-php готов (HTTP $code)"

echo "==> 3/4: установка Playwright (если нужно) и запуск спека"
if [[ ! -d node_modules ]]; then
  npm install --no-audit --no-fund
fi
# Браузер Chromium — устанавливаем при первом запуске.
npx playwright install --with-deps chromium 2>/dev/null || npx playwright install chromium

mkdir -p screenshots
SMOKE="${SMOKE:-1}" LEGACY_URL="$LEGACY_URL" npx playwright test --project=chromium

echo "==> 4/4: подсчёт PNG и cleanup"
ls -1 screenshots/*.png 2>/dev/null | wc -l | xargs -I{} echo "    создано PNG: {}"

if [[ "${KEEP_RUNNING:-0}" != "1" ]]; then
  echo "    останавливаем docker-стек (KEEP_RUNNING=1 чтобы оставить)"
  ( cd "$LEGACY_DOCKER_DIR" && docker compose stop )
fi

echo "ГОТОВО. Эталоны в $BASELINE_DIR/screenshots/"
