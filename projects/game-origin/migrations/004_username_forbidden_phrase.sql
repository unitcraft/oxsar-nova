-- План 46/48 Шаг 0 (149-ФЗ): фраза для ошибки регистрации, когда
-- никнейм содержит запрещённое слово из UGC-blacklist
-- (projects/game-nova/configs/moderation/blacklist.yaml).
--
-- Намеренно не выдаём найденный корень обратно пользователю —
-- иначе атакующий через подбор узнает содержимое blacklist'а.
--
-- ID 11000+ зарезервированы под новые фразы плана 46+ (legacy-импорт
-- использовал ID до ~10000).

INSERT INTO `na_phrases`
  (`phraseid`, `languageid`, `phrasegroupid`, `title`, `content`, `translated`)
VALUES
  (11001, 1, 7, 'USERNAME_FORBIDDEN', 'Имя содержит запрещённое слово.', 1),
  (11002, 2, 7, 'USERNAME_FORBIDDEN', 'Username contains a forbidden word.', 0),
  (11003, 3, 7, 'USERNAME_FORBIDDEN', 'Der Name enthält ein verbotenes Wort.', 0),
  (11004, 4, 7, 'USERNAME_FORBIDDEN', 'Ім&#39;я містить заборонене слово.', 0),
  (11005, 5, 7, 'USERNAME_FORBIDDEN', '用户名包含禁用词。', 0),
  (11006, 6, 7, 'USERNAME_FORBIDDEN', 'Імя ўтрымлівае забароненае слова.', 0)
ON DUPLICATE KEY UPDATE `content` = VALUES(`content`);
