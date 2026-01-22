# Use Debian (slim) to verify if prebuilt binaries work (glibc vs musl)
FROM node:20-slim AS builder

ENV SVELTE_SKIP_CANVAS_BUILD=true

WORKDIR /app
COPY package.json package-lock.json ./

# Clean install dependencies with SSL verification disabled for build environment
RUN npm config set strict-ssl false
RUN npm ci

COPY . .
RUN npm run build
RUN npm prune --production

# Use Debian (slim) for runtime validation to match builder
FROM node:20-slim
WORKDIR /app
COPY --from=builder /app/build build/
COPY --from=builder /app/node_modules node_modules/
COPY package.json .

EXPOSE 3000
ENV NODE_ENV=production
CMD [ "node", "build" ]
