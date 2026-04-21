# Frontend dev-режим в docker. Нужен, потому что HMR и live-reload
# важнее, чем prod-оптимальный bundle; а локально node может быть
# не установлен.
#
# Образ: node:22-alpine. Внутри — `npm install` + `npm run dev`.
# vite слушает :5173 на 0.0.0.0 (см. --host в compose command).
# Прокси на backend настраивается через VITE_PROXY_TARGET (см.
# vite.config.ts).

FROM node:22-alpine
WORKDIR /app

# Кешируем слой с npm install: package.json/package-lock.json меняется
# редко, исходники — часто.
COPY frontend/package.json ./
# package-lock.json может отсутствовать на первой сборке — не падаем.
COPY frontend/package-lock.json* ./
RUN npm install --no-audit --no-fund

# Копируем исходники; в dev compose-окружении эта директория
# перекроется bind-mount'ом из хоста — код редактируется без ребилда
# образа.
COPY frontend ./

EXPOSE 5173

# --host 0.0.0.0 чтобы порт был доступен из других контейнеров и
# с хоста (иначе vite слушает только 127.0.0.1 внутри контейнера).
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
