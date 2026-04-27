# Fixtures для UI-сравнения с legacy

**Не коммитятся в git** — содержат реальные emails/passwords/IP реальных
игроков (PII). См. `.gitignore`.

## Как сгенерировать

Требуется запущенный legacy oxsar2 (`d:\Sources\oxsar2`, контейнер
`oxsar2-mysql-1`):

```bash
bash projects/game-origin/tools/snapshot-legacy-user.sh
# или для другого юзера:
bash projects/game-origin/tools/snapshot-legacy-user.sh 2094833
```

Создаст `test-user-snapshot.sql` (~80-100 KB) с данными test-юзера +
связанные таблицы (alliance, чат, рейтинг).

## Как применить к нашей БД

```bash
bash projects/game-origin/tools/apply-test-user-fixture.sh
```

Скрипт сохранит `na_user.global_user_id` (наша надстройка для JWT),
очистит существующего dev-юзера, накатит snapshot, восстановит
`global_user_id`. После этого `dev-login.php` логинит как test-юзер
из legacy.

## Что внутри snapshot

- **na_user**: 1 строка test-юзера + ~50 топ-юзеров для рейтинга +
  alliance-members + chat-authors.
- **na_planet, na_galaxy**: 9 планет test-юзера (5 обычных + 4 луны).
- **na_building2planet**: ~65 построек.
- **na_research2user**: 19 исследований.
- **na_artefact2user**: 55 артефактов.
- **na_unit2shipyard**: 15 типов юнитов.
- **na_alliance + na_user2ally**: alliance "Tagi" (aid=42).
- **na_chat**: последние 100 сообщений.
- Прочее: officer, password, referral, user_experience, и т.п.

## Зачем

Без идентичных данных сравнение наших страниц с legacy бесполезно
(будут ложные различия от разного состояния БД). См. план 37.5d.
