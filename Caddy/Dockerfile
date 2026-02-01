FROM caddy:2.10.2-builder AS builder
RUN go env -w GOPROXY=https://goproxy.cn,direct \
    && xcaddy build --with github.com/mholt/caddy-ratelimit
FROM caddy:2.10.2-alpine
COPY --from=builder /usr/bin/caddy /usr/bin/caddy