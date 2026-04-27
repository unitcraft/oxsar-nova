<?php
/**
 * File — clean-room rewrite (план 43 Ф.2). Заменяет одноимённый класс
 * фреймворка Recipe (GPL). Только методы, реально вызываемые в проекте.
 *
 * Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 */

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class File
{
    /**
     * Преобразует число байт в человекочитаемую строку: «1.5 MB», «12 KB».
     * Используется в MSG для размера папки сообщений.
     */
    public static function bytesToString($bytes)
    {
        $bytes = (float)$bytes;
        if($bytes <= 0) { return '0 B'; }
        $units = array('B', 'KB', 'MB', 'GB', 'TB');
        $idx = 0;
        while($bytes >= 1024 && $idx < count($units) - 1)
        {
            $bytes /= 1024;
            $idx++;
        }
        $precision = $idx === 0 ? 0 : ($bytes >= 100 ? 0 : 1);
        return number_format($bytes, $precision, '.', '').' '.$units[$idx];
    }

    /**
     * Возвращает расширение файла в lowercase. Без точки. Пустая строка
     * если расширения нет.
     */
    public static function getFileExtension($filename)
    {
        if(!is_string($filename) || $filename === '') { return ''; }
        $ext = pathinfo($filename, PATHINFO_EXTENSION);
        return strtolower((string)$ext);
    }

    /**
     * Удаляет файл. Не бросает на отсутствующем файле (legacy-семантика —
     * для cache/cleanup-операций).
     */
    public static function rmFile($path)
    {
        if(!is_string($path) || $path === '') { return false; }
        if(!is_file($path)) { return false; }
        return @unlink($path);
    }

    /**
     * Удаляет всё содержимое директории (файлы и поддиректории). Сама
     * директория остаётся. Используется для очистки сессий/кэша.
     */
    public static function rmDirectoryContent($dir)
    {
        if(!is_string($dir) || $dir === '' || !is_dir($dir)) { return false; }
        $items = @scandir($dir);
        if($items === false) { return false; }
        foreach($items as $item)
        {
            if($item === '.' || $item === '..') { continue; }
            $path = rtrim($dir, '/\\').DIRECTORY_SEPARATOR.$item;
            if(is_dir($path) && !is_link($path))
            {
                self::rmDirectoryContent($path);
                @rmdir($path);
            }
            else
            {
                @unlink($path);
            }
        }
        return true;
    }
}
