FROM node:22-alpine
WORKDIR /app

COPY projects/portal/frontend/package.json ./
COPY projects/portal/frontend/package-lock.json* ./
RUN npm install --no-audit --no-fund

COPY projects/portal/frontend ./

EXPOSE 5174

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
