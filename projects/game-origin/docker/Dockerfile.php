FROM php:8.3-fpm-alpine

RUN docker-php-ext-install pdo pdo_mysql

# Memcache extension (legacy uses old `class Memcache` API, see MemCacheHandler.class.php).
# pecl memcache 8.2 supports PHP 8.x.
RUN apk add --no-cache --virtual .build-deps $PHPIZE_DEPS zlib-dev \
    && pecl install memcache-8.2 \
    && docker-php-ext-enable memcache \
    && apk del .build-deps

WORKDIR /var/www

COPY --chown=www-data:www-data . .
