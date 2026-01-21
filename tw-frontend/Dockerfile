# Use Alpine for building to avoid Debian repo connection issues
FROM node:20-alpine AS builder

# Install build dependencies for canvas
RUN apk add --no-cache \
    build-base \
    cairo-dev \
    pango-dev \
    jpeg-dev \
    giflib-dev \
    librsvg-dev \
    pixman-dev \
    python3

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
