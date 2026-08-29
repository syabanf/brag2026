# syntax=docker/dockerfile:1

# ── build ─────────────────────────────────────────────────────────────────
FROM node:22-alpine AS build
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Vite inlines VITE_* at build time, so the API URL is a build argument rather
# than a runtime variable.
ARG VITE_API_URL=http://localhost:8080
ENV VITE_API_URL=$VITE_API_URL

# Quick sign-in advertises the seeded demo accounts on the login screen. Off
# unless asked for, so a real deployment never ships it.
ARG VITE_DEMO_SIGNIN=false
ENV VITE_DEMO_SIGNIN=$VITE_DEMO_SIGNIN

RUN npm run build


# ── runtime ───────────────────────────────────────────────────────────────
FROM nginx:1.27-alpine AS runtime

COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1

CMD ["nginx", "-g", "daemon off;"]
