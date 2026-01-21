# Use Alpine for building to avoid Debian repo connection issues
# Network issues persist, removing native dependencies to bypass apk add
FROM node:20-alpine AS builder

# Skip native dependencies for canvas - hoping for pure JS fallback or optional dependency
ENV SVELTE_SKIP_CANVAS_BUILD=true

WORKDIR /app
COPY package.json package-lock.json ./
# Using npm install instead of npm ci to allow lock file updates when new packages are added
RUN npm install
COPY . .
RUN npm run build
RUN npm prune --production

# Use Alpine for smaller runtime image
FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/build build/
COPY --from=builder /app/node_modules node_modules/
COPY package.json .

EXPOSE 3000
ENV NODE_ENV=production
CMD [ "node", "build" ]
