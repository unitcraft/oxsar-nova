#!/bin/bash
# План 37.5d.3: сравнение всех страниц game-origin с legacy oxsar2.
#
# Pre-condition: snapshot test-юзера применён к нашей БД через
# apply-test-user-fixture.sh (37.5d.2). После этого dev-login → как test-юзер.
#
# Что делает:
# 1. Логин в legacy (test/quoYaMe1wHo4xaci) и у нас (dev-login.php).
# 2. Для каждой страницы из списка — curl у нас + curl в legacy.
# 3. Нормализация (убрать таймстампы, ID, числа) → diff.
# 4. Отчёт: список страниц с ненулевым diff + размер расхождения.
#
# Артефакт-каталог: tools/compare-output/
#   - ours/<page>.html        — наш HTML
#   - legacy/<page>.html      — legacy HTML
#   - normalized-ours/<page>  — после нормализации
#   - normalized-legacy/<page>— после нормализации
#   - diff/<page>.diff        — diff нормализованных версий
#   - report.md               — итоговый отчёт
#
# Использование:
#   bash projects/game-legacy-php/tools/compare-with-legacy.sh

set -eu

OUR_BASE="http://localhost:8092"
LEG_BASE="http://localhost:8080"
OUR_LOGIN="${OUR_BASE}/dev-login.php"
LEG_LOGIN="${LEG_BASE}/login.php"
LEG_USER="test"
LEG_PASS="quoYaMe1wHo4xaci"

OUTDIR="$(dirname "$0")/compare-output"
mkdir -p "$OUTDIR"/{ours,legacy,normalized-ours,normalized-legacy,diff}

OUR_COOKIES="$OUTDIR/ours.cookies"
LEG_COOKIES="$OUTDIR/legacy.cookies"
REPORT="$OUTDIR/report.md"

# Список страниц для сравнения (см. план 37.5d).
# Исключены: Page (абстрактный), Construction (абстрактный), FleetAjax,
# Logout, Officer, Moderator, RocketAttack, MonitorPlanet, ChatPro,
# ChatAlly, ArtefactMarketOld, EditConstruction, EditUnit, StockNew,
# TestAlienAI, ResTransferStats, Payment (соцплатежи отключены).
PAGES=(
  Main Resource Constructions Research Shipyard Defense Mission
  Galaxy Empire Stock ExchangeOpts Repair Disassemble
  Chat MSG Notepad Alliance Friends Search
  Artefacts ArtefactMarket Market Achievements Profession Tutorial
  Ranking Records Battlestats Techtree
  AdvTechCalculator Simulator BuildingInfo UnitInfo ArtefactInfo
  Preferences UserAgreement Support Widgets Changelog
  Exchange
)

# === Login ===
echo "=== Login ==="
echo "Login to legacy ($LEG_USER)..."
curl -sL -c "$LEG_COOKIES" -X POST "$LEG_LOGIN" \
  -d "username=${LEG_USER}&password=${LEG_PASS}&login=OK" \
  -o /dev/null -w "  HTTP %{http_code}\n"

echo "Login to ours (dev-login.php)..."
curl -sL -c "$OUR_COOKIES" "$OUR_LOGIN" \
  -o /dev/null -w "  HTTP %{http_code}\n"

# === Normalization ===
# Убираем таймстампы, числа, ID — оставляем структуру для diff.
normalize() {
  # 1. Даты dd.mm.yyyy hh:mm:ss → DATE
  # 2. id="xxx_NNNN" → id="xxx_N"
  # 3. value="123" → value="N"
  # 4. sid=abc123 → sid=SID
  # 5. Числа в текстовом контенте >= 2 цифр → N (но не в src/href)
  # 6. SQL queries в HTML-комментах → удалить
  sed -e 's|[0-9]\{2\}\.[0-9]\{2\}\.[0-9]\{4\}\s*[0-9]*:[0-9]*\(:[0-9]*\)\?|DATE|g' \
      -e 's|sid=[a-zA-Z0-9]*|sid=SID|g' \
      -e 's|id="\([a-zA-Z_]*\)[0-9]\+"|id="\1N"|g' \
      -e 's|value="[0-9]\+"|value="N"|g' \
      -e 's|>\s*[0-9]\{2,\}\s*<|>N<|g' \
      -e 's|>\s*[0-9]\{1,\}[.,][0-9]\+\s*<|>N<|g' \
      "$1"
}

# === Comparison loop ===
echo ""
echo "=== Comparison ==="

# Header в отчёт
{
  echo "# UI-сравнение game-origin vs legacy"
  echo ""
  echo "**Дата**: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "**Юзер**: $LEG_USER (userid=1)"
  echo ""
  echo "| Страница | Наш size | Legacy size | Diff lines | Статус |"
  echo "|---|---:|---:|---:|---|"
} > "$REPORT"

TOTAL=0
DIFF_COUNT=0
ERROR_COUNT=0

for page in "${PAGES[@]}"; do
  TOTAL=$((TOTAL+1))

  # Curl наш
  our_size=$(curl -sL -b "$OUR_COOKIES" -c "$OUR_COOKIES" \
    -o "$OUTDIR/ours/$page.html" \
    -w "%{size_download}" \
    "${OUR_BASE}/?go=$page" || echo "0")

  # Curl legacy
  leg_size=$(curl -sL -b "$LEG_COOKIES" -c "$LEG_COOKIES" \
    -o "$OUTDIR/legacy/$page.html" \
    -w "%{size_download}" \
    "${LEG_BASE}/game.php/$page" || echo "0")

  # Если хотя бы один не вернул HTML — пометить ошибкой
  if [ "$our_size" = "0" ] || [ "$leg_size" = "0" ]; then
    ERROR_COUNT=$((ERROR_COUNT+1))
    printf "%-25s our=%6s leg=%6s ERROR\n" "$page" "$our_size" "$leg_size"
    echo "| $page | $our_size | $leg_size | - | ❌ ERROR |" >> "$REPORT"
    continue
  fi

  # Нормализация
  normalize "$OUTDIR/ours/$page.html"   > "$OUTDIR/normalized-ours/$page"
  normalize "$OUTDIR/legacy/$page.html" > "$OUTDIR/normalized-legacy/$page"

  # Diff
  diff_lines=$(diff "$OUTDIR/normalized-ours/$page" "$OUTDIR/normalized-legacy/$page" 2>&1 \
    | tee "$OUTDIR/diff/$page.diff" | wc -l)

  if [ "$diff_lines" = "0" ]; then
    status="✅ identical"
  elif [ "$diff_lines" -lt "20" ]; then
    status="🟢 minor ($diff_lines lines)"
  elif [ "$diff_lines" -lt "100" ]; then
    status="🟡 moderate ($diff_lines lines)"
    DIFF_COUNT=$((DIFF_COUNT+1))
  else
    status="🔴 major ($diff_lines lines)"
    DIFF_COUNT=$((DIFF_COUNT+1))
  fi

  printf "%-25s our=%6s leg=%6s diff=%6s\n" "$page" "$our_size" "$leg_size" "$diff_lines"
  echo "| $page | $our_size | $leg_size | $diff_lines | $status |" >> "$REPORT"
done

# Footer отчёта
{
  echo ""
  echo "## Итог"
  echo ""
  echo "- Всего страниц: $TOTAL"
  echo "- С существенными расхождениями (≥20 diff lines): $DIFF_COUNT"
  echo "- Ошибок (HTTP 0/404/500): $ERROR_COUNT"
  echo ""
  echo "## Что делать дальше"
  echo ""
  echo "1. Открыть страницы с 🔴 major diff (>100 lines) — там либо"
  echo "   серьёзная регрессия, либо страница использует данные которых"
  echo "   нет в snapshot (например, MSG требует na_message)."
  echo "2. Для 🟡 moderate — диффы открыть в редакторе, классифицировать:"
  echo "   - Тип A (структурный): чинить в шаблоне/CSS."
  echo "   - Тип B (контентный, разные данные): не баг, расширить snapshot."
  echo "   - Тип C (шаблонный, не подставлено): чинить в .tpl."
  echo "3. Для 🟢 minor — обычно ОК, проверить по содержанию diff."
} >> "$REPORT"

echo ""
echo "=== Report: $REPORT ==="
cat "$REPORT"
