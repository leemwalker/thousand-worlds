# Project: Local LLM MUD Frontend
# Tech Stack: SvelteKit, TypeScript, TailwindCSS, WebSockets.

## UX Guidelines
- **Terminal Aesthetic:** Dark mode default. Monospaced fonts.
- **Mobile-First:** The input bar must remain visible on mobile keyboards and text should be easy to read.
- **Offline-Capable:** The "Shell" must load without internet.
- **Connection:** Auto-reconnect.

## Technical Rules
1. **WebSockets:** Maintain a persistent connection. Reconnect automatically on drop with exponential backoff.
2. **State Management:** Use Svelte Stores for `messages[]` and `playerStats`.
3. **Input Handling:**
   - Local Echo: Immediately show what the user typed.
   - Command History: Up/Down arrow keys to cycle previous commands.
   - Sanitize inputs before sending.

## PWA
- Ensure `manifest.json` is configured for `display: standalone`.
- Service Worker must cache static assets (fonts, images).

## 🚨 Time Dilation Handling
- The Server sends `TimeMultiplier` in the handshake.
- **Animations:** CSS animations and JS timers must multiply speeds by `TimeMultiplier`.
  - If `TimeMultiplier = 10`, a "Walk" animation plays 10x faster.
  - If `TimeMultiplier = 0.1`, the world moves in slow motion.
