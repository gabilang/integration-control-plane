# ==========================================
# Stage 1: Build the React application
# ==========================================
FROM node:25-alpine AS builder

RUN npm install -g pnpm@10

WORKDIR /app

COPY package.json ./
RUN pnpm install

COPY . .
RUN pnpm build

# ==========================================
# Stage 2: Serve with nginx
# ==========================================
FROM nginx:1.27-alpine

USER root

RUN addgroup -g 10001 -S choreouser && \
    adduser -u 10001 -S choreouser -G choreouser

COPY --chown=10001:10001 nginx.conf /etc/nginx/nginx.conf
COPY --from=builder --chown=10001:10001 /app/dist /usr/share/nginx/html

COPY --chown=10001:10001 docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

RUN chown -R 10001:10001 /var/cache/nginx && \
    chown -R 10001:10001 /var/log/nginx && \
    chown -R 10001:10001 /etc/nginx && \
    chown -R 10001:10001 /usr/share/nginx/html && \
    touch /var/run/nginx.pid && \
    chown -R 10001:10001 /var/run/nginx.pid

USER 10001

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ || exit 1

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
