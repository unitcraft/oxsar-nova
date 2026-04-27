-- План 60: удаление legacy реферальной системы из game-origin.
-- Новая реферальная программа (план 59) реализуется с нуля на portal-backend
-- и legacy-данные не переиспользует (другая семантика, антифрод, валюта).

-- Реферальная статистика и связи "кто кого пригласил".
DROP VIEW IF EXISTS na_referral_ext;
DROP TABLE IF EXISTS na_referral;

-- Локализованные строки UI реферальной системы.
-- Ключи в na_phrases.title: REFERRAL_CREDIT_MESSAGE, REFERRAL_CREDIT_SUBJECT, MENU_REFERAL.
-- Удаляем для всех языков (1=ru, 2=en, 3=de, 4=ua, 5=zh, 6=by).
DELETE FROM na_phrases WHERE title IN (
  'REFERRAL_CREDIT_MESSAGE',
  'REFERRAL_CREDIT_SUBJECT',
  'MENU_REFERAL'
);
