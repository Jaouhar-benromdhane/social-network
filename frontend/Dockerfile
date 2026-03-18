FROM node:22-alpine AS builder

WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci

COPY frontend ./
RUN npm run build

FROM nginx:1.27-alpine
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

RUN mkdir -p /app/data

ENV APP_PORT=8080
ENV DB_PATH=/app/data/social_network.db
ENV MIGRATIONS_PATH=file:///app/pkg/db/migrations/sqlite

EXPOSE 8080
CMD ["./server"]
