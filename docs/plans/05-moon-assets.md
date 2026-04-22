# План: Обновление иконок и иллюстраций луны

## Контекст

Переименование и обновление ассетов луны в проекте:
- `mond.jpg` (легаси название, немецкий) → `moon.jpg` (английское, понятное)
- Создана новая SVG иконка `moon-icon.svg` (32×32px, с кратерами)

---

## Шаги реализации

### 1. Обновить ссылки в коде (уже сделано локально)

Переименованы файлы:
```
frontend/public/images/planets/mond.jpg → moon.jpg ✅
frontend/public/images/planets/moon-icon.svg (новый) ✅
```

### 2. Найти и обновить все ссылки на `mond.jpg` в коде

**Поиск:** `grep -r "mond\.jpg" frontend/`

Обновить в следующих файлах (если найдутся):
- `frontend/src/ui/PlanetImage.tsx` или похожие компоненты
- `frontend/src/features/*/` (все экраны планет)
- CSS/конфиги с путями к изображениям

**Замена:** `mond.jpg` → `moon.jpg`

### 3. Добавить SVG иконку в компоненты

Создать React компонент для SVG иконки луны:

```tsx
// frontend/src/ui/MoonIcon.tsx
export function MoonIcon({ size = 20 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 32 32" fill="none">
      {/* SVG content из moon-icon.svg */}
    </svg>
  );
}
```

Или просто импортировать SVG:
```tsx
import moonIcon from '@/images/planets/moon-icon.svg';

// Использование:
<img src={moonIcon} width="20" height="20" alt="луна" />
```

### 4. Обновить использование в App.tsx

**Текущий код** (строка 123):
```typescript
{f.dst_is_moon ? ' 🌑' : ''}
```

**Вариант 1 (emoji):** оставить как есть

**Вариант 2 (SVG иконка):**
```typescript
{f.dst_is_moon ? <MoonIcon size={16} /> : ''}
```

**Вариант 3 (изображение):**
```typescript
{f.dst_is_moon ? <img src="/images/planets/moon-icon.svg" width="16" height="16" alt="луна" /> : ''}
```

### 5. Обновить PlanetCarousel (если используется mond.jpg)

Проверить, где показывается лупа вместе с планетой в карусели:
- Заменить `mond.jpg` на `moon.jpg`
- Убедиться, что размер и расположение оверлея правильное

### 6. Тестирование

- [x] Проверить, что луна отображается везде корректно (OverviewScreen, карусель, входящие атаки)
- [x] Убедиться, что новая SVG иконка видна при малых размерах (20px, 32px) — MoonIcon компонент
- [x] Проверить в браузере (http://localhost:5173) — работает корректно

### 7. Git коммит

```bash
git add frontend/public/images/planets/moon.jpg
git add frontend/public/images/planets/moon-icon.svg
git commit -m "refactor(ui): rename moon asset mond.jpg → moon.jpg, add moon-icon.svg"
```

---

## Файлы

| Файл | Статус | Описание |
|------|--------|---------|
| `frontend/public/images/planets/moon.jpg` | ✅ | JPG фотография луны (3.2 KB) |
| `frontend/public/images/planets/moon-icon.svg` | ✅ | SVG иконка луны с кратерами (32×32) |
| `frontend/public/images/planets/mond.jpg` | ❌ | Удалено (переименовано) |

---

## Статус реализации

✅ **ЗАВЕРШЕНА** — 2026-04-23

Все шаги выполнены:
1. ✅ Файлы переименованы (mond.jpg → moon.jpg)
2. ✅ Добавлена moon-icon.svg иконка
3. ✅ Создан MoonIcon React компонент (frontend/src/ui/MoonIcon.tsx)
4. ✅ Ссылки обновлены в OverviewScreen.tsx
5. ✅ Луна отображается корректно везде
6. ✅ Тестировано в браузере

## Замечания

- Старый файл `mond.jpg` больше не нужен (удалён при переименовании)
- SVG иконка масштабируется без потери качества
- Названия на английском удобнее для интернационализации (i18n)
- По цвету SVG иконка близка к JPG (~#b8b8b8 серый)
- MoonIcon компонент готов к использованию в любом месте UI
