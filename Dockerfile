# syntax=docker/dockerfile:1

# ── deps ──────────────────────────────────────────────────────────────────
FROM node:22-alpine AS deps
WORKDIR /app

# PGlite ships a WASM build that expects a glibc-compatible loader.
RUN apk add --no-cache libc6-compat

COPY package.json package-lock.json ./
RUN npm ci


# ── builder ───────────────────────────────────────────────────────────────
FROM node:22-alpine AS builder
WORKDIR /app

RUN apk add --no-cache libc6-compat

COPY --from=deps /app/node_modules ./node_modules
COPY . .

ENV NEXT_TELEMETRY_DISABLED=1

# Demo availability is baked in here because NEXT_PUBLIC_* is inlined at build
# time. Override with --build-arg NEXT_PUBLIC_DEMO_MODE=false for production.
ARG NEXT_PUBLIC_DEMO_MODE=true
ENV NEXT_PUBLIC_DEMO_MODE=$NEXT_PUBLIC_DEMO_MODE

RUN npm run build


# ── runner ────────────────────────────────────────────────────────────────
FROM node:22-alpine AS runner
WORKDIR /app

RUN apk add --no-cache libc6-compat

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

RUN addgroup --system --gid 1001 nodejs \
 && adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# Demo mode boots PGlite and applies these at runtime by reading them from
# disk, so they must ship with the image rather than being bundled.
COPY --from=builder --chown=nextjs:nodejs /app/db ./db

# .env carries the demo defaults; real environment variables still win over it.
COPY --from=builder --chown=nextjs:nodejs /app/.env ./.env

USER nextjs
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=40s --retries=3 \
  CMD node -e "fetch('http://127.0.0.1:3000/api/health').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"

CMD ["node", "server.js"]
