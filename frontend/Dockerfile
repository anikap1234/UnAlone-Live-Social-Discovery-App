FROM node:18-alpine AS builder
WORKDIR /app
ARG VITE_API_BASE
ARG VITE_MAPBOX_TOKEN
ENV VITE_API_BASE=${VITE_API_BASE}
ENV VITE_MAPBOX_TOKEN=${VITE_MAPBOX_TOKEN}
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY . .
RUN npm run build

FROM nginx:stable-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
