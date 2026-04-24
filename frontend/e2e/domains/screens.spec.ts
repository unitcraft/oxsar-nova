// Ф.2: доменные экраны (каждый — smoke+минимальная проверка содержимого).
// Для более глубоких сценариев (отправка сообщения, покупка лота) нужны
// отдельные спеки с подготовкой БД — добавляются по мере необходимости.

import { test, expect } from '@playwright/test';
import { loginAs, type TestUserName } from '../fixtures/auth';
import { goToTab } from '../helpers/nav';

function screenOpens(
  tab: string,
  user: TestUserName,
  mustContain?: RegExp,
): void {
  test(`${tab} opens for ${user}`, async ({ page }) => {
    await loginAs(page, user);
    await goToTab(page, tab);
    await expect(page.locator('main.ox-content')).toBeVisible();
    if (mustContain) {
      await expect(page.locator('main.ox-content')).toContainText(mustContain);
    }
  });
}

test.describe('Ф.2.1 repair', () => {
  screenOpens('repair', 'bob', /ремонт|repair|disassemble|разобр/i);
});

test.describe('Ф.2.2 market (exchange + lots)', () => {
  screenOpens('market', 'bob', /рынок|market|обмен|курс/i);
});

test.describe('Ф.2.3 rockets', () => {
  screenOpens('rockets', 'bob', /ракет|rocket|impact|запуск/i);
});

test.describe('Ф.2.4 artefacts', () => {
  screenOpens('artefacts', 'bob', /артефакт|artefact|нет/i);
});

test.describe('Ф.2.5 art-market', () => {
  screenOpens('art-market', 'bob', /арт|лот|market|offer|кредит/i);
});

test.describe('Ф.2.6 officers', () => {
  screenOpens('officers', 'bob', /офицер|officer|адмир|инженер|геолог|меркур/i);
});

test.describe('Ф.2.8 alliance', () => {
  // bob — member альянса [UT]
  screenOpens('alliance', 'bob', /альянс|alliance|UT|Testers/i);
});

test.describe('Ф.2.9 chat', () => {
  screenOpens('chat', 'bob', /чат|chat|сообщ/i);
});

test.describe('Ф.2.10 score', () => {
  screenOpens('score', 'bob', /рейтинг|score|highscore|очк/i);
});

test.describe('Ф.2.11 achievements', () => {
  screenOpens('achievements', 'bob', /достижен|achievement/i);
});

test.describe('Ф.2.12 tutorial / profession', () => {
  screenOpens('profession', 'bob', /профессия|profession|туториал|tutorial/i);
});

test.describe('Ф.2.13 battle sim', () => {
  screenOpens('sim', 'bob', /симуля|battle|sim|истреб/i);
});
