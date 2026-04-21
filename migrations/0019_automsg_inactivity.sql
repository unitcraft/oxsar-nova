-- +goose Up
-- Снимаем FK с automsg_sent.key чтобы можно было использовать
-- составные ключи (INACTIVITY_REMINDER_2026W17) без шаблона в defs.
ALTER TABLE automsg_sent DROP CONSTRAINT IF EXISTS automsg_sent_key_fkey;

-- Шаблон уведомления о неактивности.
INSERT INTO automsg_defs (key, title, body_template, folder) VALUES
    ('INACTIVITY_REMINDER',
     'Давно не виделись, {{username}}!',
     'Прошло несколько дней с вашего последнего визита. '
     || 'Не забывайте — ваши шахты продолжают добычу, хранилища могут переполниться. '
     || 'Возвращайтесь и проверьте статус ваших планет.',
     2);

-- +goose Down
INSERT INTO automsg_defs (key, title, body_template, folder)
    SELECT 'INACTIVITY_REMINDER', '', '', 2
    WHERE NOT EXISTS (SELECT 1 FROM automsg_defs WHERE key='INACTIVITY_REMINDER');

ALTER TABLE automsg_sent
    ADD CONSTRAINT automsg_sent_key_fkey
    FOREIGN KEY (key) REFERENCES automsg_defs(key) ON DELETE CASCADE;

DELETE FROM automsg_defs WHERE key = 'INACTIVITY_REMINDER';
