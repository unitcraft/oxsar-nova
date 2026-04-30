FROM php:8.3-fpm-alpine

RUN docker-php-ext-install pdo pdo_mysql

# Memcached extension (new API, recommended for PHP 8.x).
# Legacy used old `class Memcache` API, but PECL memcache не поддерживается
# для PHP 8.3 — MemCacheHandler обновлён на новое `class Memcached` (37.5c.1).
RUN apk add --no-cache --virtual .build-deps $PHPIZE_DEPS libmemcached-dev zlib-dev \
    && apk add --no-cache libmemcached \
    && pecl install memcached \
    && docker-php-ext-enable memcached \
    && apk del .build-deps

# GD + FreeType для генерации preview-картинок артефактов
# (artImageUrl → public/artefact-image.php, см. план 37.5d.9).
RUN apk add --no-cache --virtual .build-deps-gd $PHPIZE_DEPS libpng-dev freetype-dev libjpeg-turbo-dev \
    && apk add --no-cache libpng freetype libjpeg-turbo \
    && docker-php-ext-configure gd --with-freetype --with-jpeg \
    && docker-php-ext-install -j$(nproc) gd \
    && apk del .build-deps-gd

# План 43 Ф.1: Composer для замены Recipe-фреймворка (GPL) на пакеты под
# PolyForm-совместимыми лицензиями. unzip нужен для composer install.
RUN apk add --no-cache git unzip \
    && curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer

# OpenJDK для запуска src/game/Assault.jar (порт oxsar2-java/Assault.java).
# Используется и реальными битвами (Assault.class.php), и страницей
# симулятора (Simulator.class.php), и event-loop (EventHandler.class.php).
# headless — без AWT/Swing, экономит ~50 МБ. Java 17 LTS, обратно
# совместима с bytecode major=50 (Java 6) у Assault.jar.
RUN apk add --no-cache openjdk17-jre-headless

# MySQL JDBC connector (mysql-connector-j 8.0.33) для Assault.jar.
# Сам Assault.jar собран без вшитого JDBC (Class-Path: пустой), а код зовёт
# `Class.forName("com.mysql.jdbc.Driver")` — старый legacy-класс. В
# connector/J 8.0.33 он сохранён как deprecated shim → выдаёт warning, но
# работает. PHP-вызовы exec'ат `java -cp Assault.jar:$MYSQL_CONNECTOR_JAR`.
# В Alpine community JDBC-драйверов под Java нет, поэтому качаем с Maven
# Central. Версия зафиксирована для воспроизводимости.
RUN curl -sSLo /opt/mysql-connector.jar \
    'https://repo1.maven.org/maven2/com/mysql/mysql-connector-j/8.0.33/mysql-connector-j-8.0.33.jar' \
    && echo "e2a3b2fc726a1ac64e998585db86b30fa8bf3f706195b78bb77c5f99bf877bd9  /opt/mysql-connector.jar" | sha256sum -c -

WORKDIR /var/www

COPY --chown=www-data:www-data . .

# composer install выполняется ПОСЛЕ COPY (composer.json должен быть на диске).
# --no-dev: продакшен-сборка без phpunit и пр. --no-interaction: CI-friendly.
# --optimize-autoloader: prebuild PSR-4 classmap для скорости.
# Если composer.json отсутствует (ранние стадии плана 43) — игнорируем,
# чтобы образ собирался для legacy-кода без vendor/.
RUN if [ -f composer.json ]; then \
        composer install --no-dev --no-interaction --optimize-autoloader --no-progress; \
    fi
