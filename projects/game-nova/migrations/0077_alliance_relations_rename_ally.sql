-- План 67 Ф.1: миграция данных 'ally' → 'friend' в alliance_relationships.
--
-- Отделена от 0076 потому что postgres не позволяет использовать
-- только что добавленные enum-значения в DML внутри той же
-- транзакции, в которой ALTER TYPE ADD VALUE был выполнен. Goose
-- даёт каждой миграции свою транзакцию — поэтому к 0077 значение
-- 'friend' уже доступно.
--
-- Семантика: 'ally' (nova-исходное) и 'friend' (origin-инспирированное)
-- означают одно и то же — союз; миграция уравнивает их. Сервис в Ф.2
-- будет писать только 'friend'; принимать на вход оба значения для
-- обратной совместимости (если в каких-то фикстурах сохранилось
-- 'ally').

-- +goose Up
UPDATE alliance_relationships SET relation = 'friend' WHERE relation = 'ally';

-- +goose Down
UPDATE alliance_relationships SET relation = 'ally' WHERE relation = 'friend';
