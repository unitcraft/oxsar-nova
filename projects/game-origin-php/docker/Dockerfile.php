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
