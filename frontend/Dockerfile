FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm install

COPY . .

RUN npm run build

FROM nginx:1.28-alpine

WORKDIR /usr/share/nginx/html

COPY --from=builder /app/dist ./

EXPOSE 3000

