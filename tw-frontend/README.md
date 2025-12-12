# Thousand Worlds - Frontend

A SvelteKit 2.0 Progressive Web App (PWA) for the Thousand Worlds MUD Platform, featuring real-time WebSocket communication, offline support, and mobile-optimized gameplay.

## 🛠️ Technology Stack

| Component | Technology |
|-----------|------------|
| **Framework** | SvelteKit 2.0 |
| **Language** | TypeScript 5.0 |
| **Styling** | Tailwind CSS 3.4 |
| **Build Tool** | Vite 5.0 |
| **PWA** | vite-plugin-pwa 0.17 |
| **Real-time** | WebSocket, nats.ws |
| **Unit Testing** | Vitest 3.2 |
| **E2E Testing** | Playwright 1.57 |

---

## 📁 Project Structure

```
tw-frontend/
├── src/
│   ├── app.html                 # HTML template
│   ├── app.css                  # Global styles
│   ├── app.d.ts                 # TypeScript declarations
│   ├── routes/                  # SvelteKit pages
│   │   ├── +layout.svelte       # Root layout
│   │   ├── +page.svelte         # Landing/login page
│   │   └── game/                # Game interface
│   │       └── +page.svelte     # Main game page
│   └── lib/
│       ├── components/          # UI components
│       │   ├── Character/       # Character display components
│       │   ├── Drift/           # Behavioral drift monitor
│       │   ├── Input/           # Command input components
│       │   ├── Inventory/       # Inventory management
│       │   ├── Layout/          # Layout components
│       │   ├── Map/             # Map visualization
│       │   ├── Output/          # Game output/messages
│       │   ├── PWA/             # PWA install prompts
│       │   └── WorldEntry.svelte
│       ├── services/            # API and services
│       ├── stores/              # Svelte stores (state management)
│       ├── types/               # TypeScript type definitions
│       ├── network/             # Network utilities
│       └── utils/               # Utility functions
├── static/                      # Static assets (icons, images)
├── tests/
│   └── e2e/                     # Playwright E2E tests
├── vite.config.ts               # Vite + PWA configuration
├── vitest.config.ts             # Vitest configuration
├── playwright.config.ts         # Playwright configuration
├── svelte.config.js             # SvelteKit configuration
├── tailwind.config.js           # Tailwind CSS configuration
└── Dockerfile                   # Production container
```

---

## 🎨 Components

### Character Components
| Component | Description |
|-----------|-------------|
| `Character/` | Character stats, attributes, and character sheet display |

### Drift Components
| Component | Description |
|-----------|-------------|
| `Drift/` | Behavioral drift monitor showing personality changes when inhabiting NPCs |

### Input Components
| Component | Description |
|-----------|-------------|
| `Input/` | Command input, command parser, autocomplete, and command history |

### Inventory Components
| Component | Description |
|-----------|-------------|
| `Inventory/` | Item management, equipment slots, weight tracking |

### Layout Components
| Component | Description |
|-----------|-------------|
| `Layout/` | Page layouts, navigation, responsive containers |

### Map Components
| Component | Description |
|-----------|-------------|
| `Map/` | 2D top-down map visualization, fog of war, entity markers |

### Output Components
| Component | Description |
|-----------|-------------|
| `Output/` | Game message log, formatted text output, color-coded messages |

### PWA Components
| Component | Description |
|-----------|-------------|
| `PWA/` | Install prompts, offline status indicators |

---

## 🌐 PWA Configuration

The app is configured as a Progressive Web App with the following features:

### Service Worker
- **Auto-update**: Automatically updates when new versions are available
- **Offline fallback**: Graceful degradation when offline

### Caching Strategies
| Resource Type | Strategy | Cache Duration |
|---------------|----------|----------------|
| API calls | NetworkFirst | 5 minutes |
| Images | CacheFirst | 30 days |
| JS/CSS | StaleWhileRevalidate | Until updated |
| Fonts | CacheFirst | 30 days |

### Manifest
- Name: "Thousand Worlds MUD Client"
- Display: Standalone (fullscreen app experience)
- Theme color: #16213e
- Background color: #1a1a2e

---

## 🚀 Quick Start

### Prerequisites
- Node.js 18+
- npm 9+

### Install Dependencies
```bash
npm install
```

### Start Development Server
```bash
npm run dev
```

Access at: **http://localhost:5173**

The dev server proxies API calls to `http://localhost:8080` (backend).

---

## 📜 npm Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server with hot reload |
| `npm run build` | Build for production |
| `npm run preview` | Preview production build locally |
| `npm run check` | TypeScript type checking |
| `npm run test` | Run unit tests with Vitest |
| `npm run test:ui` | Run tests with Vitest UI |
| `npm run test:coverage` | Run tests with coverage report |
| `npm run test:e2e` | Run E2E tests with Playwright |
| `npm run test:e2e:ui` | Run E2E tests with Playwright UI |

---

## 🧪 Testing

### Unit Tests (Vitest)
```bash
# Run all unit tests
npm run test

# Run with UI
npm run test:ui

# Run with coverage
npm run test:coverage
```

### E2E Tests (Playwright)
```bash
# Run E2E tests
npm run test:e2e

# Run with UI (debug mode)
npm run test:e2e:ui
```

E2E tests are located in `tests/e2e/` and cover critical user flows.

---

## 🏗️ Build & Deploy

### Production Build
```bash
npm run build
```

Output is generated in `.svelte-kit/output/`.

### Docker Build
```bash
docker build -t thousand-worlds/frontend:latest .
```

### Preview Production Build
```bash
npm run preview
```

---

## ⚙️ Configuration

### Vite Configuration (`vite.config.ts`)
- PWA manifest and service worker configuration
- Dev server proxy to backend API
- Runtime caching strategies

### Development Proxy
The dev server is configured to proxy `/api/*` requests to the backend:
```typescript
proxy: {
  '/api': {
    target: 'http://localhost:8080',
    changeOrigin: true,
    ws: true  // WebSocket support
  }
}
```

### Environment
- Dev server runs on port 5173
- Backend API expected at port 8080
- WebSocket connections proxied automatically

---

## 📱 Mobile Optimization

- **Responsive Design**: Tailwind breakpoints for all screen sizes
- **Touch-Friendly**: Large tap targets, swipe gestures
- **PWA Install**: "Add to Home Screen" on iOS/Android
- **Offline Mode**: Read-only access when disconnected
- **Network Binding**: Server binds to `0.0.0.0` for mobile device access

---

## 📚 Related Documentation

- [../README.MD](../README.MD) - Project overview
- [../tw-backend/README.md](../tw-backend/README.md) - Backend documentation
- [../features.md](../features.md) - Feature specifications
