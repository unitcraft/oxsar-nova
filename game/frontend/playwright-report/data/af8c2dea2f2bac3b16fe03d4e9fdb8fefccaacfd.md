# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: critical/auth.spec.ts >> auth: session >> me endpoint returns session user
- Location: e2e/critical/auth.spec.ts:47:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.ox-header-right').getByText('bob')
Expected: visible
Error: strict mode violation: locator('.ox-header-right').getByText('bob') resolved to 2 elements:
    1) <span>Bob-Home</span> aka getByRole('banner').getByText('Bob-Home')
    2) <span>bob</span> aka getByText('bob', { exact: true })

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('.ox-header-right').getByText('bob')

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - banner [ref=e4]:
    - generic [ref=e5]: ✦ OXSAR
    - generic [ref=e6]:
      - generic [ref=e7]:
        - generic [ref=e8]: 🟠
        - generic [ref=e9]: Мет
        - generic [ref=e10]:
          - generic [ref=e11]: 25.0K
          - generic [ref=e12]: 25k
      - generic [ref=e13]:
        - generic [ref=e14]: 💎
        - generic [ref=e15]: Крем
        - generic [ref=e16]:
          - generic [ref=e17]: 25.0K
          - generic [ref=e18]: 25k
      - generic [ref=e19]:
        - generic [ref=e20]: 💧
        - generic [ref=e21]: Водор
        - generic [ref=e22]:
          - generic [ref=e23]: 25.0K
          - generic [ref=e24]: 25k
      - generic [ref=e25]:
        - generic [ref=e26]: ⚡
        - generic [ref=e27]:
          - generic [ref=e28]: "2314"
          - generic [ref=e29]: "-1059"
    - generic [ref=e30]:
      - generic "Кредиты" [ref=e31]:
        - generic [ref=e32]: 💳
        - generic [ref=e33]: "5010"
      - generic [ref=e34]: 10:06:34
      - button "🔍 Ctrl+K" [ref=e35] [cursor=pointer]
      - generic [ref=e37]:
        - generic [ref=e38]: 🪐
        - generic [ref=e39]: Bob-Home
        - generic [ref=e40]: 🏠
        - generic [ref=e41]: "[1:1:7]"
      - generic [ref=e42]: bob
      - button "Выйти" [ref=e43] [cursor=pointer]
  - generic [ref=e44]:
    - navigation [ref=e45]:
      - generic [ref=e46]: Планета
      - link "🏠 Обзор" [ref=e47] [cursor=pointer]:
        - /url: "#overview"
        - generic [ref=e48]: 🏠
        - generic [ref=e49]: Обзор
      - link "⚙️ Сырьё" [ref=e50] [cursor=pointer]:
        - /url: "#resource"
        - generic [ref=e51]: ⚙️
        - generic [ref=e52]: Сырьё
      - link "🏗 Постройки" [ref=e53] [cursor=pointer]:
        - /url: "#buildings"
        - generic [ref=e54]: 🏗
        - generic [ref=e55]: Постройки
      - link "🔬 Исследования" [ref=e56] [cursor=pointer]:
        - /url: "#research"
        - generic [ref=e57]: 🔬
        - generic [ref=e58]: Исследования
      - link "🚀 Верфь" [ref=e59] [cursor=pointer]:
        - /url: "#shipyard"
        - generic [ref=e60]: 🚀
        - generic [ref=e61]: Верфь
      - link "🔧 Ремонт" [ref=e62] [cursor=pointer]:
        - /url: "#repair"
        - generic [ref=e63]: 🔧
        - generic [ref=e64]: Ремонт
      - link "🎖 Профессия" [ref=e65] [cursor=pointer]:
        - /url: "#profession"
        - generic [ref=e66]: 🎖
        - generic [ref=e67]: Профессия
      - link "🌐 Империя" [ref=e68] [cursor=pointer]:
        - /url: "#empire"
        - generic [ref=e69]: 🌐
        - generic [ref=e70]: Империя
      - link "🌳 Техдерево" [ref=e71] [cursor=pointer]:
        - /url: "#techtree"
        - generic [ref=e72]: 🌳
        - generic [ref=e73]: Техдерево
      - link "⚙️ Настройки" [ref=e74] [cursor=pointer]:
        - /url: "#settings"
        - generic [ref=e75]: ⚙️
        - generic [ref=e76]: Настройки
      - generic [ref=e78]: Космос
      - link "🌌 Галактика" [ref=e79] [cursor=pointer]:
        - /url: "#galaxy"
        - generic [ref=e80]: 🌌
        - generic [ref=e81]: Галактика
      - link "🛸 Флот" [ref=e82] [cursor=pointer]:
        - /url: "#fleet"
        - generic [ref=e83]: 🛸
        - generic [ref=e84]: Флот
      - link "💥 Ракеты" [ref=e85] [cursor=pointer]:
        - /url: "#rockets"
        - generic [ref=e86]: 💥
        - generic [ref=e87]: Ракеты
      - generic [ref=e89]: Общение
      - link "📨 Сообщения" [ref=e90] [cursor=pointer]:
        - /url: "#messages"
        - generic [ref=e91]: 📨
        - generic [ref=e92]: Сообщения
      - link "💬 Чат" [ref=e93] [cursor=pointer]:
        - /url: "#chat"
        - generic [ref=e94]: 💬
        - generic [ref=e95]: Чат
      - link "🤝 Альянс" [ref=e96] [cursor=pointer]:
        - /url: "#alliance"
        - generic [ref=e97]: 🤝
        - generic [ref=e98]: Альянс
      - link "📝 Блокнот" [ref=e99] [cursor=pointer]:
        - /url: "#notepad"
        - generic [ref=e100]: 📝
        - generic [ref=e101]: Блокнот
      - link "🎁 Рефералы" [ref=e102] [cursor=pointer]:
        - /url: "#referral"
        - generic [ref=e103]: 🎁
        - generic [ref=e104]: Рефералы
      - link "⭐ Друзья" [ref=e105] [cursor=pointer]:
        - /url: "#friends"
        - generic [ref=e106]: ⭐
        - generic [ref=e107]: Друзья
      - generic [ref=e109]: Торговля
      - link "💱 Рынок" [ref=e110] [cursor=pointer]:
        - /url: "#market"
        - generic [ref=e111]: 💱
        - generic [ref=e112]: Рынок
      - link "💎 Артефакты" [ref=e113] [cursor=pointer]:
        - /url: "#artefacts"
        - generic [ref=e114]: 💎
        - generic [ref=e115]: Артефакты
      - link "🏪 Рынок артефактов" [ref=e116] [cursor=pointer]:
        - /url: "#art-market"
        - generic [ref=e117]: 🏪
        - generic [ref=e118]: Рынок артефактов
      - link "⭐ Офицеры" [ref=e119] [cursor=pointer]:
        - /url: "#officers"
        - generic [ref=e120]: ⭐
        - generic [ref=e121]: Офицеры
      - generic [ref=e123]: Статистика
      - link "🏆 Рейтинг" [ref=e124] [cursor=pointer]:
        - /url: "#score"
        - generic [ref=e125]: 🏆
        - generic [ref=e126]: Рейтинг
      - link "🥇 Достижения" [ref=e127] [cursor=pointer]:
        - /url: "#achievements"
        - generic [ref=e128]: 🥇
        - generic [ref=e129]: Достижения
      - link "⚔ История боёв" [ref=e130] [cursor=pointer]:
        - /url: "#battlestats"
        - generic [ref=e131]: ⚔
        - generic [ref=e132]: История боёв
      - link "🏅 Рекорды" [ref=e133] [cursor=pointer]:
        - /url: "#records"
        - generic [ref=e134]: 🏅
        - generic [ref=e135]: Рекорды
      - link "⚔️ Симулятор боя" [ref=e136] [cursor=pointer]:
        - /url: "#sim"
        - generic [ref=e137]: ⚔️
        - generic [ref=e138]: Симулятор боя
    - main [ref=e139]:
      - generic [ref=e140]:
        - generic [ref=e141]:
          - generic [ref=e142]:
            - generic [ref=e143]: Очки
            - generic [ref=e144]: "0"
          - generic [ref=e145]:
            - generic [ref=e146]: Место в рейтинге
            - generic [ref=e147]: "#1"
          - generic [ref=e149]:
            - generic [ref=e150]: Сейчас играют
            - generic [ref=e151]: "5"
          - generic [ref=e152]:
            - generic [ref=e153]: За 24 часа
            - generic [ref=e154]: "5"
        - generic [ref=e155]:
          - generic [ref=e156]: "🔗 Реферальная ссылка:"
          - code [ref=e157]: http://frontend:5173/?ref=00000000-0000-0000-0000-000000000003
          - button "Скопировать" [ref=e158] [cursor=pointer]
        - generic [ref=e159]:
          - generic [ref=e160]:
            - generic [ref=e161]:
              - generic [ref=e162]: Bob-Home
              - generic [ref=e163]: "[1:1:7]"
            - button "⚙️" [ref=e164] [cursor=pointer]
          - generic [ref=e165]:
            - generic [ref=e166]:
              - generic [ref=e167]: 📐
              - generic [ref=e168]: "Диаметр:"
              - generic [ref=e169]: 18 800 км
            - generic [ref=e170]:
              - generic [ref=e171]: 🔲
              - generic [ref=e172]: "Поля:"
              - generic [ref=e173]: 0 / 90
            - generic [ref=e174]:
              - generic [ref=e175]: 🌡️
              - generic [ref=e176]: "Температура:"
              - generic [ref=e177]: "-20°C … 40°C"
          - generic [ref=e178]:
            - generic [ref=e179]:
              - generic [ref=e180]: 🟠
              - generic [ref=e181]:
                - generic [ref=e182]: Металл
                - text: 25.0K
                - generic [ref=e183]: 2 769/ч
            - generic [ref=e184]:
              - generic [ref=e185]: 💎
              - generic [ref=e186]:
                - generic [ref=e187]: Кремний
                - text: 25.0K
                - generic [ref=e188]: 1 373/ч
            - generic [ref=e189]:
              - generic [ref=e190]: 💧
              - generic [ref=e191]:
                - generic [ref=e192]: Водород
                - text: 25.0K
                - generic [ref=e193]: 541/ч
  - contentinfo [ref=e194]: oxsar-nova v0.1.0 — dev preview
```

# Test source

```ts
  1  | // Ф.1.1: авторизация. Регистрация, логин, неверный пароль, logout.
  2  | // Refresh-токен не тестируем напрямую (он автоматический) — это часть
  3  | // smoke-спека: токен подставляется через fixture и работает.
  4  | 
  5  | import { test, expect } from '@playwright/test';
  6  | import { loginAs, TEST_PASSWORD } from '../fixtures/auth';
  7  | 
  8  | test.describe('auth: login screen', () => {
  9  |   test('shows login form by default', async ({ page }) => {
  10 |     await page.goto('/');
  11 |     await expect(page.getByText('OXSAR').first()).toBeVisible();
  12 |     await expect(page.getByRole('button', { name: 'Войти', exact: true })).toBeVisible();
  13 |   });
  14 | 
  15 |   test('login with valid credentials lands on Overview', async ({ page }) => {
  16 |     await page.goto('/');
  17 |     await page.getByRole('button', { name: 'Войти', exact: true }).first().click();
  18 | 
  19 |     // Поле email принимает username тоже
  20 |     await page.getByLabel(/e-?mail|логин/i).first().fill('alice');
  21 |     await page.getByLabel('Пароль').fill(TEST_PASSWORD);
  22 |     await page.getByRole('button', { name: 'Войти', exact: true }).last().click();
  23 | 
  24 |     // После логина появляется шапка с логотипом в sidebar-layout.
  25 |     await expect(page.locator('.ox-header-logo')).toBeVisible({ timeout: 15_000 });
  26 |   });
  27 | 
  28 |   test('wrong password shows error and stays on login', async ({ page }) => {
  29 |     await page.goto('/');
  30 |     await page.getByLabel(/e-?mail|логин/i).first().fill('alice');
  31 |     await page.getByLabel('Пароль').fill('wrong-password-999');
  32 |     await page.getByRole('button', { name: 'Войти', exact: true }).last().click();
  33 | 
  34 |     await expect(page.locator('.ox-error')).toBeVisible({ timeout: 10_000 });
  35 |     // Остались на login-экране
  36 |     await expect(page.locator('.ox-header-logo')).not.toBeVisible();
  37 |   });
  38 | });
  39 | 
  40 | test.describe('auth: session', () => {
  41 |   test('logout returns to login screen', async ({ page }) => {
  42 |     await loginAs(page, 'alice');
  43 |     await page.getByRole('button', { name: 'Выйти' }).click();
  44 |     await expect(page.getByRole('button', { name: 'Войти', exact: true })).toBeVisible({ timeout: 10_000 });
  45 |   });
  46 | 
  47 |   test('me endpoint returns session user', async ({ page }) => {
  48 |     await loginAs(page, 'bob');
  49 |     // Имя пользователя отображается в header
> 50 |     await expect(page.locator('.ox-header-right').getByText('bob')).toBeVisible();
     |                                                                     ^ Error: expect(locator).toBeVisible() failed
  51 |   });
  52 | });
  53 | 
```