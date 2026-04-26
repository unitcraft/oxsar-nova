FROM php:8.3-fpm-alpine

RUN docker-php-ext-install pdo pdo_mysql

WORKDIR /var/www

COPY --chown=www-data:www-data . .
