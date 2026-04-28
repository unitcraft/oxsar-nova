<?php
/**
 * Moderation — UGC blacklist для game-origin (план 46/48 Шаг 0).
 *
 * Читает общий YAML `projects/game-nova/configs/moderation/blacklist.yaml`
 * (источник истины и для game-nova/auth-сервисов на Go) и проверяет
 * пользовательский ввод (никнейм, тег альянса, сообщение в чате).
 *
 * Логика match'а — паритет с Go-версией
 * (projects/game-nova/backend/internal/moderation/blacklist.go):
 *   - lowercase + удаление всего, что не буква (любая кириллица/латиница);
 *   - поиск подстроки.
 *
 * YAML-формат (наш) простой: верхний уровень — мап
 * `groupName: [str1, str2, ...]`. Группы носят информационный
 * характер; класс собирает всё в плоский список корней.
 *
 * Парсер YAML — собственный минимальный, чтобы не тащить в game-origin
 * symfony/yaml. Поддерживает только формат blacklist.yaml; не
 * предназначен для произвольного YAML.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if (!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Moderation
{
    /** @var string[] плоский список нормализованных корней */
    private static $roots = null;

    /** @var string|null путь, из которого реально загружен список */
    private static $loadedFrom = null;

    /**
     * Возвращает абсолютный путь к общему blacklist.yaml.
     * Можно переопределить через define('OXSAR_MODERATION_BLACKLIST', '/path').
     */
    public static function blacklistPath()
    {
        if (defined('OXSAR_MODERATION_BLACKLIST')) {
            return OXSAR_MODERATION_BLACKLIST;
        }
        // APP_ROOT_DIR = projects/game-legacy-php/src/
        // YAML живёт в projects/game-nova/configs/moderation/blacklist.yaml
        return APP_ROOT_DIR . '../../game-nova/configs/moderation/blacklist.yaml';
    }

    /**
     * Ленивая загрузка списка. Если файла нет — список пустой,
     * isForbidden() всегда вернёт false (фильтр отключён). Это
     * соответствует поведению Go-версий на старте без файла.
     */
    private static function ensureLoaded()
    {
        if (self::$roots !== null) { return; }
        $path = self::blacklistPath();
        if (!is_readable($path)) {
            self::$roots = array();
            self::$loadedFrom = null;
            return;
        }
        $raw = @file_get_contents($path);
        if ($raw === false) {
            self::$roots = array();
            return;
        }
        self::$roots = self::parseSimpleYaml($raw);
        self::$loadedFrom = $path;
    }

    /**
     * Возвращает ['allowed'|'forbidden', $matchedRoot]
     * (matched root — пустая строка для allowed).
     */
    public static function check($input)
    {
        self::ensureLoaded();
        if (empty(self::$roots)) {
            return array('allowed', '');
        }
        $n = self::normalize((string)$input);
        if ($n === '') {
            return array('allowed', '');
        }
        foreach (self::$roots as $root) {
            if ($root !== '' && strpos($n, $root) !== false) {
                return array('forbidden', $root);
            }
        }
        return array('allowed', '');
    }

    /**
     * Шорткат: true, если $input содержит запрещённый корень.
     */
    public static function isForbidden($input)
    {
        list($status, ) = self::check($input);
        return $status === 'forbidden';
    }

    /**
     * Маскирует запрещённые корни в строке. Если ничего не найдено —
     * возвращает $input без изменений.
     *
     * Семантика — паритет с MaskForbidden из Go-версии
     * (projects/game-nova/backend/internal/moderation/blacklist.go):
     * если найдено хоть одно запрещённое слово, то ВСЕ буквенные
     * символы в строке заменяются на «*». Цифры, пробелы, пунктуация,
     * HTML/BBCode-теги остаются как есть. Длина строки сохраняется.
     *
     * Это «крупная» стратегия: лучше перерезать невинный текст
     * (если в нём есть мат), чем пропустить мат через regex-обход.
     * Для серьёзной модерации — сторонний сервис (см. план 50 §4).
     */
    public static function mask($input)
    {
        $input = (string)$input;
        if (!self::isForbidden($input)) { return $input; }
        $out = '';
        $len = mb_strlen($input, 'UTF-8');
        for ($i = 0; $i < $len; $i++) {
            $ch = mb_substr($input, $i, 1, 'UTF-8');
            $out .= self::isLetter($ch) ? '*' : $ch;
        }
        return $out;
    }

    /**
     * Количество корней в списке (для логов при старте/в админке).
     */
    public static function size()
    {
        self::ensureLoaded();
        return count(self::$roots);
    }

    /**
     * Откуда реально загружен список (или null, если файл недоступен).
     */
    public static function loadedFrom()
    {
        self::ensureLoaded();
        return self::$loadedFrom;
    }

    /**
     * Сброс кэша (для тестов).
     */
    public static function reset()
    {
        self::$roots = null;
        self::$loadedFrom = null;
    }

    /**
     * Нормализация: lowercase + удаление всего, что не буква.
     * Паритет с Go-isLetter: a-z, а-я, ё, A-Z, А-Я, Ё.
     */
    public static function normalize($s)
    {
        $s = mb_strtolower(trim((string)$s), 'UTF-8');
        $out = '';
        $len = mb_strlen($s, 'UTF-8');
        for ($i = 0; $i < $len; $i++) {
            $ch = mb_substr($s, $i, 1, 'UTF-8');
            if (self::isLetter($ch)) { $out .= $ch; }
        }
        return $out;
    }

    private static function isLetter($ch)
    {
        $code = self::ord($ch);
        if ($code >= 0x61 && $code <= 0x7A) { return true; } // a-z
        if ($code >= 0x41 && $code <= 0x5A) { return true; } // A-Z
        if ($code >= 0x0430 && $code <= 0x044F) { return true; } // а-я
        if ($code >= 0x0410 && $code <= 0x042F) { return true; } // А-Я
        if ($code === 0x0451 || $code === 0x0401) { return true; } // ё / Ё
        return false;
    }

    /**
     * Кодпоинт первого UTF-8 символа в строке.
     */
    private static function ord($ch)
    {
        if ($ch === '') { return 0; }
        $c = mb_convert_encoding($ch, 'UCS-4BE', 'UTF-8');
        $u = unpack('N', $c);
        return $u[1];
    }

    /**
     * Минимальный YAML-парсер для нашего формата:
     *   group_name:
     *     - word1
     *     - word2
     *     - "quoted word"
     *
     * Возвращает плоский список нормализованных непустых корней.
     * Не поддерживает: nested, anchors, multiline strings, flow-style.
     * Этого хватает для blacklist.yaml; если формат изменится —
     * меняем здесь и в Go-loader'е одновременно.
     */
    private static function parseSimpleYaml($raw)
    {
        $roots = array();
        $lines = preg_split('/\r\n|\n|\r/', $raw);
        foreach ($lines as $line) {
            $trim = trim($line);
            if ($trim === '' || $trim[0] === '#') { continue; }
            // Заголовок группы — игнорируем (нам нужны только значения).
            if (preg_match('/^[A-Za-z0-9_]+:\s*(?:#.*)?$/', $trim)) { continue; }
            // Строка списка: "  - word" или "  - 'word'" или "  - \"word\"".
            if (preg_match('/^-\s+(.*)$/', $trim, $m)) {
                $val = self::stripYamlValue($m[1]);
                if ($val !== '') {
                    $n = self::normalize($val);
                    if ($n !== '') { $roots[] = $n; }
                }
            }
        }
        return $roots;
    }

    private static function stripYamlValue($v)
    {
        // Удаляем хвостовой #-комментарий.
        $v = preg_replace('/\s+#.*$/', '', $v);
        $v = trim($v);
        // Снимаем кавычки.
        $len = strlen($v);
        if ($len >= 2 && (($v[0] === '"' && $v[$len-1] === '"') || ($v[0] === "'" && $v[$len-1] === "'"))) {
            $v = substr($v, 1, $len - 2);
        }
        return $v;
    }
}
