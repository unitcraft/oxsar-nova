FROM node:22-alpine
WORKDIR /app

COPY portal/frontend/package.json ./
COPY portal/frontend/package-lock.json* ./
RUN npm install --no-audit --no-fund

COPY portal/frontend ./

EXPOSE 5174

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
