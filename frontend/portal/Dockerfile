FROM node:22-alpine
WORKDIR /app

COPY frontend/portal/package.json ./
COPY frontend/portal/package-lock.json* ./
RUN npm install --no-audit --no-fund

COPY frontend/portal ./

EXPOSE 5174

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
