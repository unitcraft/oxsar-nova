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

WORKDIR /var/www

COPY --chown=www-data:www-data . .
