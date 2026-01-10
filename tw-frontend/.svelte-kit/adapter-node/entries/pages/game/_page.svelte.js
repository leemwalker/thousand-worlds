import { c as create_ssr_component, d as each, e as escape, f as createEventDispatcher, b as add_attribute, o as onDestroy, a as subscribe, v as validate_component, n as null_to_empty } from "../../../chunks/ssr.js";
import "@sveltejs/kit/internal";
import "../../../chunks/exports.js";
import "../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../chunks/state.svelte.js";
import { g as gameStore, v as validateServerMessage, m as mapStore } from "../../../chunks/schemas.js";
import { Engine } from "@babylonjs/core/Engines/engine.js";
import { Scene } from "@babylonjs/core/scene.js";
import { ArcRotateCamera } from "@babylonjs/core/Cameras/arcRotateCamera.js";
import { PointLight } from "@babylonjs/core/Lights/pointLight.js";
import { HemisphericLight } from "@babylonjs/core/Lights/hemisphericLight.js";
import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder.js";
import "@babylonjs/core/Meshes/transformNode.js";
import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial.js";
import { Texture } from "@babylonjs/core/Materials/Textures/texture.js";
import "@babylonjs/core/Materials/Textures/rawTexture.js";
import { Vector3 } from "@babylonjs/core/Maths/math.vector.js";
import { Color4, Color3 } from "@babylonjs/core/Maths/math.color.js";
import "@babylonjs/core/Layers/glowLayer.js";
import { Effect } from "@babylonjs/core/Materials/effect.js";
import "@babylonjs/core/Materials/shaderMaterial.js";
import "@babylonjs/core/Meshes/mesh.js";
import "@babylonjs/core/Animations/animation.js";
import "@babylonjs/core/Animations/easing.js";
import "@babylonjs/core/PostProcesses/postProcess.js";
import { w as writable, d as derived } from "../../../chunks/index.js";
import "@babylonjs/core/Materials/Textures/cubeTexture.js";
import { ParticleSystem } from "@babylonjs/core/Particles/particleSystem.js";
import { UniversalCamera } from "@babylonjs/core/Cameras/universalCamera.js";
import "@babylonjs/core/Collisions/collisionCoordinator.js";
import "@babylonjs/core/Cameras/Inputs/freeCameraKeyboardMoveInput.js";
import "@babylonjs/core/Cameras/Inputs/freeCameraMouseInput.js";
const DB_NAME = "mud-command-queue";
const DB_VERSION = 1;
const STORE_NAME = "commands";
const MAX_RETRIES = 3;
class CommandQueue {
  db = null;
  isProcessing = false;
  async init() {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(DB_NAME, DB_VERSION);
      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        this.db = request.result;
        resolve();
      };
      request.onupgradeneeded = (event) => {
        const db = event.target.result;
        if (!db.objectStoreNames.contains(STORE_NAME)) {
          const store = db.createObjectStore(STORE_NAME, { keyPath: "id" });
          store.createIndex("status", "status", { unique: false });
          store.createIndex("timestamp", "timestamp", { unique: false });
        }
      };
    });
  }
  // Enqueue command for later sending (O(1) IndexedDB write)
  async enqueue(text) {
    if (!this.db) throw new Error("Database not initialized");
    const queuedCommand = {
      id: crypto.randomUUID(),
      text,
      timestamp: Date.now(),
      retryCount: 0,
      status: "pending"
    };
    return new Promise((resolve, reject) => {
      const transaction = this.db.transaction([STORE_NAME], "readwrite");
      const store = transaction.objectStore(STORE_NAME);
      const request = store.add(queuedCommand);
      request.onsuccess = () => resolve(queuedCommand.id);
      request.onerror = () => reject(request.error);
    });
  }
  // Get all pending commands (for UI display)
  async getPending() {
    if (!this.db) return [];
    return new Promise((resolve, reject) => {
      const transaction = this.db.transaction([STORE_NAME], "readonly");
      const store = transaction.objectStore(STORE_NAME);
      const index = store.index("status");
      const request = index.getAll("pending");
      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
  }
  // Get count of pending commands
  async getPendingCount() {
    if (!this.db) return 0;
    return new Promise((resolve, reject) => {
      const transaction = this.db.transaction([STORE_NAME], "readonly");
      const store = transaction.objectStore(STORE_NAME);
      const index = store.index("status");
      const request = index.count("pending");
      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
  }
  // Process queue when connection restored
  async processQueue(sendFn) {
    if (!this.db || this.isProcessing) return;
    this.isProcessing = true;
    try {
      const pending = await this.getPending();
      for (const queuedCmd of pending) {
        await this.updateStatus(queuedCmd.id, "processing");
        try {
          await sendFn(queuedCmd.text);
          await this.remove(queuedCmd.id);
        } catch (error) {
          console.error("Failed to send queued command:", error);
          if (queuedCmd.retryCount < MAX_RETRIES) {
            await this.incrementRetry(queuedCmd.id);
          } else {
            await this.updateStatus(queuedCmd.id, "failed");
          }
        }
      }
    } finally {
      this.isProcessing = false;
    }
  }
  // Clear all failed commands
  async clearFailed() {
    if (!this.db) return;
    return new Promise((resolve, reject) => {
      const transaction = this.db.transaction([STORE_NAME], "readwrite");
      const store = transaction.objectStore(STORE_NAME);
      const index = store.index("status");
      const request = index.openCursor(IDBKeyRange.only("failed"));
      request.onsuccess = (event) => {
        const cursor = event.target.result;
        if (cursor) {
          cursor.delete();
          cursor.continue();
        } else {
          resolve();
        }
      };
      request.onerror = () => reject(request.error);
    });
  }
  // Private helpers
  async updateStatus(id, status) {
    if (!this.db) return;
    return new Promise((resolve, reject) => {
      const transaction = this.db.transaction([STORE_NAME], "readwrite");
      const store = transaction.objectStore(STORE_NAME);
      const getRequest = store.get(id);
      getRequest.onsuccess = () => {
        const cmd = getRequest.result;
        if (cmd) {
          cmd.status = status;
          const putRequest = store.put(cmd);
          putRequest.onsuccess = () => resolve();
          putRequest.onerror = () => reject(putRequest.error);
        } else {
          resolve();
        }
      };
      getRequest.onerror = () => reject(getRequest.error);
    });
  }
  async incrementRetry(id) {
    if (!this.db) return;
    return new Promise((resolve, reject) => {
      const transaction = this.db.transaction([STORE_NAME], "readwrite");
      const store = transaction.objectStore(STORE_NAME);
      const getRequest = store.get(id);
      getRequest.onsuccess = () => {
        const cmd = getRequest.result;
        if (cmd) {
          cmd.retryCount++;
          cmd.status = "pending";
          const putRequest = store.put(cmd);
          putRequest.onsuccess = () => resolve();
          putRequest.onerror = () => reject(putRequest.error);
        } else {
          resolve();
        }
      };
      getRequest.onerror = () => reject(getRequest.error);
    });
  }
  async remove(id) {
    if (!this.db) return;
    return new Promise((resolve, reject) => {
      const transaction = this.db.transaction([STORE_NAME], "readwrite");
      const store = transaction.objectStore(STORE_NAME);
      const request = store.delete(id);
      request.onsuccess = () => resolve();
      request.onerror = () => reject(request.error);
    });
  }
  // Cleanup on close
  close() {
    if (this.db) {
      this.db.close();
      this.db = null;
    }
  }
}
const commandQueue = new CommandQueue();
class GameWebSocket {
  ws = null;
  reconnectAttempts = 0;
  maxReconnectAttempts = 5;
  reconnectDelay = 1e3;
  currentCharacterId;
  isIntentionalDisconnect = false;
  // Store for connection status
  connected = writable(false);
  // Store for pending command count
  pendingCommands = writable(0);
  // Message handler
  messageHandlers = /* @__PURE__ */ new Set();
  // Reconnection callbacks - fire when connection is re-established after disconnect
  reconnectCallbacks = /* @__PURE__ */ new Set();
  wasConnectedBefore = false;
  connect(characterId) {
    console.log("[WebSocket] Attempting to connect...", { characterId });
    this.isIntentionalDisconnect = false;
    if (characterId) {
      this.currentCharacterId = characterId;
    }
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsHost = window.location.host;
    let wsUrl = `${protocol}//${wsHost}/api/game/ws`;
    if (this.currentCharacterId) {
      wsUrl += `?character_id=${encodeURIComponent(this.currentCharacterId)}`;
    }
    console.log("[WebSocket] URL:", wsUrl);
    try {
      this.ws = new WebSocket(wsUrl);
      this.ws.onopen = () => {
        console.log("[WebSocket] Connection opened!");
        this.connected.set(true);
        const isReconnect = this.reconnectAttempts > 0 || this.wasConnectedBefore;
        this.reconnectAttempts = 0;
        this.wasConnectedBefore = true;
        gameStore.setLoading(false);
        this.processQueuedCommands();
        if (isReconnect) {
          console.log("[WebSocket] Reconnection successful, firing callbacks");
          this.reconnectCallbacks.forEach((cb) => {
            try {
              cb();
            } catch (e) {
              console.error("[WebSocket] Reconnect callback error:", e);
            }
          });
        }
      };
      this.ws.binaryType = "arraybuffer";
      this.ws.onmessage = (event) => {
        try {
          if (event.data instanceof ArrayBuffer) {
            this.handleBinaryMessage(event.data);
            return;
          }
          const rawData = event.data.toString();
          const parts = rawData.split("\n").filter((p) => p.trim() !== "");
          for (const part of parts) {
            try {
              const parsed = JSON.parse(part);
              const validation = validateServerMessage(parsed);
              if (!validation.valid) {
                console.warn("[WebSocket] Message failed validation, processing anyway");
              }
              const message = parsed;
              this.handleMessage(message);
            } catch (e) {
              console.error("[WebSocket] Failed to parse message part:", e);
            }
          }
        } catch (error) {
          console.error("[WebSocket] Failed to process message:", error);
        }
      };
      this.ws.onerror = (error) => {
        console.error("[WebSocket] Error:", error);
        gameStore.setError("Connection error");
      };
      this.ws.onclose = () => {
        console.log("[WebSocket] Connection closed");
        this.connected.set(false);
        if (!this.isIntentionalDisconnect) {
          this.attemptReconnect();
        }
      };
    } catch (error) {
      console.error("[WebSocket] Failed to create WebSocket:", error);
      gameStore.setError("Failed to create connection");
    }
  }
  disconnect() {
    this.isIntentionalDisconnect = true;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.connected.set(false);
  }
  sendRawCommand(text, payload) {
    console.log("[WebSocket] sendRawCommand called:", text, payload);
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error("[WebSocket] Not connected, readyState:", this.ws?.readyState);
      return;
    }
    const message = {
      type: "command",
      data: { text, payload }
    };
    console.log("[WebSocket] Sending command:", JSON.stringify(message));
    this.ws.send(JSON.stringify(message));
    console.log("[WebSocket] Command sent successfully");
  }
  async sendCommandWithQueue(text) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      await commandQueue.enqueue(text);
      await this.updatePendingCount();
      console.log("Command queued for later sending");
      return;
    }
    this.sendRawCommand(text);
  }
  async processQueuedCommands() {
    try {
      await commandQueue.processQueue(async (text) => {
        return new Promise((resolve, reject) => {
          if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            reject(new Error("WebSocket not connected"));
            return;
          }
          const message = {
            type: "command",
            data: { text }
          };
          this.ws.send(JSON.stringify(message));
          resolve();
        });
      });
      await this.updatePendingCount();
    } catch (error) {
      console.error("Error processing queued commands:", error);
    }
  }
  async updatePendingCount() {
    const count = await commandQueue.getPendingCount();
    this.pendingCommands.set(count);
  }
  onMessage(handler) {
    this.messageHandlers.add(handler);
    return () => {
      this.messageHandlers.delete(handler);
    };
  }
  /**
   * Register a callback to be called when WebSocket reconnects after a disconnect.
   * Useful for refreshing world map or other state after connection is restored.
   * Returns an unsubscribe function.
   */
  onReconnect(handler) {
    this.reconnectCallbacks.add(handler);
    return () => {
      this.reconnectCallbacks.delete(handler);
    };
  }
  handleMessage(message) {
    if (message.type === "map_update") {
      const mappedData = this.mapBackendToFrontend(message.data);
      mapStore.setMapData(mappedData);
    } else if (message.type === "game_message") {
      const gameMsg = message;
      if (gameMsg.data?.metadata && gameMsg.data.type === "map_update") {
        const mappedData = this.mapBackendToFrontend(gameMsg.data.metadata);
        mapStore.setMapData(mappedData);
      }
      gameStore.addMessage({
        type: "game_message",
        timestamp: message.timestamp || Date.now(),
        content: gameMsg.data.content,
        sender: gameMsg.data.sender,
        channel: gameMsg.data.channel
      });
    } else if (message.type === "state_update") {
      gameStore.updateStats(message.data.stats || {});
      if (message.data.inventory) {
        gameStore.setInventory(message.data.inventory);
      }
    } else if (message.type === "world_map_image_response") {
      const d = message.data;
      gameStore.updateWorld({
        textureBlob: d.imageBlob,
        heightmapBlob: d.heightmapBlob,
        materialBlob: d.materialBlob,
        iceBlob: d.iceBlob,
        geo: {
          seaLevel: d.sea_level || 0,
          maxElevation: d.max_elevation || 0,
          minElevation: d.min_elevation || 0
        }
        // Maintain existing satellites if not provided
      });
    } else if (message.type === "world_reset") {
      console.log("[WS] world_reset received - switching to molten planet view");
      gameStore.resetWorld();
    }
    this.messageHandlers.forEach((handler) => {
      try {
        handler(message);
      } catch (error) {
        console.error("Message handler error:", error);
      }
    });
  }
  mapBackendToFrontend(data) {
    if (data && Array.isArray(data.tiles)) {
      return data;
    }
    if (data && Array.isArray(data.cells)) {
      return {
        ...data,
        tiles: data.cells.map((cell) => ({
          x: cell.q,
          y: cell.r,
          biome: cell.biome_type || "Default",
          elevation: cell.elevation || 0,
          // Map other fields if needed, or pass through extra props usually ignored
          occluded: cell.occluded,
          is_player: cell.is_player,
          entities: cell.entities,
          portal: cell.portal
        }))
      };
    }
    return data;
  }
  handleBinaryMessage(buffer) {
    const view = new DataView(buffer);
    let offset = 0;
    const msgType = view.getUint8(offset);
    offset += 1;
    if (msgType === 1) {
      const jsonLen = view.getUint32(offset, false);
      offset += 4;
      const jsonBytes = new Uint8Array(buffer, offset, jsonLen);
      const jsonStr = new TextDecoder().decode(jsonBytes);
      const jsonData = JSON.parse(jsonStr);
      offset += jsonLen;
      view.getUint32(offset, false);
      offset += 4;
      const imageLen = view.getUint32(offset, false);
      offset += 4;
      const imageBytes = new Uint8Array(buffer, offset, imageLen);
      const imageBlob = new Blob([imageBytes], { type: "image/webp" });
      offset += imageLen;
      let gridData = null;
      if (offset + 4 <= buffer.byteLength) {
        const gridLen = view.getUint32(offset, false);
        offset += 4;
        if (gridLen > 0 && offset + gridLen <= buffer.byteLength) {
          gridData = buffer.slice(offset, offset + gridLen);
          offset += gridLen;
          console.log(`[WebSocket] Parsed grid data: ${gridLen} bytes`);
        }
      }
      let heightmapBlob = null;
      if (offset + 4 <= buffer.byteLength) {
        const heightmapLen = view.getUint32(offset, false);
        offset += 4;
        if (heightmapLen > 0 && offset + heightmapLen <= buffer.byteLength) {
          const heightmapBytes = new Uint8Array(buffer, offset, heightmapLen);
          heightmapBlob = new Blob([heightmapBytes], { type: "image/png" });
          offset += heightmapLen;
          console.log(`[WebSocket] Parsed heightmap data: ${heightmapLen} bytes`);
        }
      }
      let materialBlob = null;
      if (offset + 4 <= buffer.byteLength) {
        const materialLen = view.getUint32(offset, false);
        offset += 4;
        if (materialLen > 0 && offset + materialLen <= buffer.byteLength) {
          const materialBytes = new Uint8Array(buffer, offset, materialLen);
          materialBlob = new Blob([materialBytes], { type: "image/png" });
          offset += materialLen;
          console.log(`[WebSocket] Parsed material data: ${materialLen} bytes`);
        }
      }
      let iceBlob = null;
      if (offset + 4 <= buffer.byteLength) {
        const iceLen = view.getUint32(offset, false);
        offset += 4;
        if (iceLen > 0 && offset + iceLen <= buffer.byteLength) {
          const iceBytes = new Uint8Array(buffer, offset, iceLen);
          iceBlob = new Blob([iceBytes], { type: "image/png" });
          offset += iceLen;
          console.log(`[WebSocket] Parsed ice data: ${iceLen} bytes`);
        }
      }
      let normalMapBlob = null;
      if (offset + 4 <= buffer.byteLength) {
        const normalMapLen = view.getUint32(offset, false);
        offset += 4;
        if (normalMapLen > 0 && offset + normalMapLen <= buffer.byteLength) {
          const normalMapBytes = new Uint8Array(buffer, offset, normalMapLen);
          normalMapBlob = new Blob([normalMapBytes], { type: "image/png" });
          offset += normalMapLen;
          console.log(`[WebSocket] Parsed normal map data: ${normalMapLen} bytes`);
        }
      }
      const message = {
        type: "world_map_image_response",
        data: {
          ...jsonData,
          imageBlob,
          // WebP image blob
          gridData,
          // Binary grid data (or null)
          heightmapBlob,
          // PNG heightmap blob (or null)
          materialBlob,
          // PNG material data (or null)
          iceBlob,
          // PNG ice sheet data (or null)
          normalMapBlob
          // PNG normal map data (or null)
        },
        timestamp: Date.now()
      };
      this.handleMessage(message);
    } else if (msgType === 2) {
      const jsonLen = view.getUint32(offset, false);
      offset += 4;
      const jsonBytes = new Uint8Array(buffer, offset, jsonLen);
      const jsonStr = new TextDecoder().decode(jsonBytes);
      const jsonData = JSON.parse(jsonStr);
      offset += jsonLen;
      view.getUint32(offset, false);
      offset += 4;
      const imageSize = jsonData.imageSize || 0;
      const heightmapSize = jsonData.heightmapSize || 0;
      const imageBytes = new Uint8Array(buffer, offset, imageSize);
      offset += imageSize;
      const heightmapBytes = new Uint8Array(buffer, offset, heightmapSize);
      offset += heightmapSize;
      console.log(`[WebSocket] Tile ${jsonData.face}_${jsonData.level}_${jsonData.x}_${jsonData.y}: image=${imageSize} heightmap=${heightmapSize}`);
      const message = {
        type: "world_tile_response",
        data: {
          ...jsonData,
          imageBytes,
          heightmapBytes
        },
        timestamp: Date.now()
      };
      this.handleMessage(message);
    } else {
      console.warn("[WebSocket] Unknown binary message type:", msgType);
    }
  }
  attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error("[WebSocket] Max reconnect attempts reached");
      gameStore.setError("Connection lost. Please refresh.");
      return;
    }
    this.reconnectAttempts++;
    const baseDelay = this.reconnectDelay;
    const exponentialDelay = Math.min(
      baseDelay * Math.pow(2, this.reconnectAttempts - 1),
      3e4
      // Cap at 30 seconds
    );
    const jitter = exponentialDelay * 0.2 * (Math.random() * 2 - 1);
    const delay = Math.round(exponentialDelay + jitter);
    console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
    setTimeout(() => {
      this.connect(this.currentCharacterId);
    }, delay);
  }
  /**
   * Request world map refresh after reconnection.
   * Called automatically when connection is re-established.
   */
  requestWorldMapRefresh() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.log("[WebSocket] Requesting world map refresh after reconnection");
      this.sendRawCommand("world_map_image", {});
    }
  }
}
const gameWebSocket = new GameWebSocket();
const hexToRgb = (hex) => {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result ? {
    r: parseInt(result[1], 16),
    g: parseInt(result[2], 16),
    b: parseInt(result[3], 16)
  } : { r: 0, g: 0, b: 0 };
};
({
  // Biomes
  Ocean: hexToRgb("#4682B4"),
  Rainforest: hexToRgb("#228B22"),
  DeciduousForest: hexToRgb("#3C783C"),
  Taiga: hexToRgb("#32503C"),
  Grassland: hexToRgb("#90EE90"),
  Desert: hexToRgb("#F4A460"),
  Tundra: hexToRgb("#E0FFFF"),
  Alpine: hexToRgb("#808080"),
  // Special
  Lobby: hexToRgb("#505064"),
  Void: hexToRgb("#14141E"),
  Default: hexToRgb("#333333")
});
[
  // Bathymetry (Deepest to Sea Level)
  { el: -6e3, color: hexToRgb("#050d1a") },
  // Deep Ocean Trenches (very dark blue/purple)
  { el: -4e3, color: hexToRgb("#0a1929") },
  // Abyssal Plain (dark blue)
  { el: -2e3, color: hexToRgb("#0d3a5c") },
  // Lower Continental Slope (medium-dark blue)
  { el: -1e3, color: hexToRgb("#115c8c") },
  // Upper Continental Slope
  { el: -200, color: hexToRgb("#1976d2") },
  // Continental Shelf (transition to medium blue)
  { el: 0, color: hexToRgb("#4fc3f7") },
  // Coastline (light cyan)
  // Hypsometric (Sea Level to Peaks)
  { el: 100, color: hexToRgb("#2e7d32") },
  // Coastal Lowlands (darker green)
  { el: 200, color: hexToRgb("#66bb6a") },
  // Continental Lowlands (lighter green/yellow-green)
  { el: 500, color: hexToRgb("#c5e1a5") },
  // Foothills/Uplands (yellow/beige)
  { el: 1e3, color: hexToRgb("#d7ccc8") },
  // Lower Mountain ranges (light brown)
  { el: 2e3, color: hexToRgb("#a1887f") },
  // Mid-Mountain ranges (medium brown)
  { el: 3e3, color: hexToRgb("#8d6e63") },
  // High Mountain ranges (dark brown/reddish brown)
  { el: 5e3, color: hexToRgb("#9e9e9e") },
  // Very High Peaks (grey/white start)
  { el: 8848, color: hexToRgb("#fafafa") }
  // Maximum (Everest height, pure white)
];
const WorldMapLegend = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let showTopography;
  let showFeatures;
  let showTectonics;
  let showTemp;
  let showMoisture;
  let showBiomes;
  const terrainColors = [
    {
      label: "Summit",
      color: "rgb(250, 250, 250)",
      desc: "> 3000m"
    },
    {
      label: "Peak",
      color: "rgb(179, 179, 179)",
      desc: "2000-3000m"
    },
    {
      label: "High Mtn",
      color: "rgb(115, 102, 102)",
      desc: "1000-2000m"
    },
    {
      label: "Mountain",
      color: "rgb(128, 115, 115)",
      desc: "500-1000m"
    },
    {
      label: "Foothill",
      color: "rgb(140, 179, 128)",
      desc: "200-500m"
    },
    {
      label: "Plain",
      color: "rgb(102, 186, 107)",
      desc: "100-200m"
    },
    {
      label: "Lowland",
      color: "rgb(46, 125, 51)",
      desc: "0-100m"
    },
    {
      label: "Coast",
      color: "rgb(79, 195, 247)",
      desc: "Sea Level"
    }
  ];
  const oceanColors = [
    {
      label: "Shallow",
      color: "rgb(0, 153, 204)",
      desc: "0 to -500m"
    },
    // vec3(0.0, 0.6, 0.8) turquoise
    {
      label: "Mid Ocean",
      color: "rgb(0, 77, 128)",
      desc: "-500 to -2000m"
    },
    // vec3(0.0, 0.3, 0.5) ocean blue
    {
      label: "Deep Ocean",
      color: "rgb(0, 26, 51)",
      desc: "< -2000m"
    }
  ];
  const biomeColors = [
    {
      label: "Ice / Tundra",
      color: "rgb(204, 217, 230)"
    },
    // 0.8, 0.85, 0.9
    {
      label: "Taiga",
      color: "rgb(77, 128, 102)"
    },
    // 0.3, 0.5, 0.4
    {
      label: "Alpine",
      color: "rgb(153, 140, 128)"
    },
    // 0.6, 0.55, 0.5
    {
      label: "Deciduous",
      color: "rgb(128, 153, 102)"
    },
    // 0.5, 0.6, 0.4
    {
      label: "Grassland",
      color: "rgb(102, 153, 77)"
    },
    // 0.4, 0.6, 0.3
    {
      label: "Savanna",
      color: "rgb(179, 166, 89)"
    },
    // vec3(0.7, 0.65, 0.35) golden
    {
      label: "Rainforest",
      color: "rgb(51, 128, 77)"
    },
    // 0.2, 0.5, 0.3
    {
      label: "Desert",
      color: "rgb(230, 204, 128)"
    }
  ];
  const lobbyColors = [
    {
      label: "Wall",
      color: "rgb(89, 89, 102)",
      desc: "Structure"
    },
    // vec3(0.35, 0.35, 0.4) dark grey
    {
      label: "Floor",
      color: "rgb(217, 209, 199)",
      desc: "Walkable"
    },
    // vec3(0.85, 0.82, 0.78) marble
    {
      label: "Portal",
      color: "rgb(204, 51, 204)",
      desc: "Teleport"
    }
  ];
  const tectonicLegend = [
    {
      label: "Oceanic Plate",
      color: "rgba(59, 130, 246, 0.4)",
      desc: "Dense crust"
    },
    {
      label: "Continental Plate",
      color: "rgba(16, 185, 129, 0.4)",
      desc: "Buoyant crust"
    }
  ];
  let { mode = "terrain" } = $$props;
  let { activeLayers = /* @__PURE__ */ new Set() } = $$props;
  let { isLobby = false } = $$props;
  let { expanded = false } = $$props;
  if ($$props.mode === void 0 && $$bindings.mode && mode !== void 0) $$bindings.mode(mode);
  if ($$props.activeLayers === void 0 && $$bindings.activeLayers && activeLayers !== void 0) $$bindings.activeLayers(activeLayers);
  if ($$props.isLobby === void 0 && $$bindings.isLobby && isLobby !== void 0) $$bindings.isLobby(isLobby);
  if ($$props.expanded === void 0 && $$bindings.expanded && expanded !== void 0) $$bindings.expanded(expanded);
  showTopography = activeLayers.has("elevation") || activeLayers.size === 0 && mode === "terrain";
  showFeatures = activeLayers.has("features") || mode === "features";
  showTectonics = activeLayers.has("tectonics") || mode === "tectonics";
  showTemp = activeLayers.has("temp") || mode === "temp";
  showMoisture = activeLayers.has("moisture") || mode === "moisture";
  showBiomes = activeLayers.has("biome") || mode === "biome";
  return `<div class="${[
    "bg-gray-900/90 backdrop-blur border border-gray-700 rounded-lg shadow-xl text-xs overflow-hidden transition-all duration-300",
    (expanded ? "w-48" : "") + " " + (!expanded ? "w-10" : "")
  ].join(" ").trim()}"><button class="w-full p-2 flex items-center justify-between hover:bg-gray-800 transition-colors" title="Toggle Legend"><div class="flex items-center gap-2"><svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 01-.447-.894L15 7m0 13V7"></path></svg> ${expanded ? `<span class="font-bold text-gray-200" data-svelte-h="svelte-1s58muz">Map Legend</span>` : ``}</div> ${expanded ? `<div class="text-gray-500 flex gap-1">${isLobby ? `🏛️` : `${showTopography ? `⛰️` : ``} ${showBiomes ? `🌿` : ``} ${showTemp ? `🌡️` : ``} ${showMoisture ? `💧` : ``} ${showFeatures ? `📍` : ``} ${showTectonics ? `📐` : ``}`}</div>` : ``}</button> ${expanded ? `<div class="p-3 border-t border-gray-800 space-y-4 max-h-[60vh] overflow-y-auto">${isLobby ? ` <div><h4 class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider" data-svelte-h="svelte-1y3woe6">Lobby</h4> <div class="space-y-1">${each(lobbyColors, (item) => {
    return `<div class="flex items-center gap-2"><div class="w-4 h-4 rounded shadow-sm border border-black/20" style="${"background-color: " + escape(item.color, true)}"></div> <div class="flex-1"><div class="text-gray-300">${escape(item.label)}</div> <div class="text-gray-600 text-[10px]">${escape(item.desc)} </div></div> </div>`;
  })}</div></div>` : ` ${showFeatures ? `<div data-svelte-h="svelte-1fulvxx"><h4 class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider">Features</h4> <div class="space-y-2"><div class="flex items-center gap-2"><span class="text-lg">🌋</span> <span class="text-gray-300">Volcano</span></div> <div class="flex items-center gap-2"><span class="text-lg">🏔️</span> <span class="text-gray-300">Peak</span></div> <div class="flex items-center gap-2"><span class="text-lg">🕳️</span> <span class="text-gray-300">Trench</span></div></div></div>` : ``}  ${showTemp ? `<div data-svelte-h="svelte-y7qnqu"><h4 class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider">Temperature</h4> <div class="space-y-2"><div class="flex items-center gap-2"><div class="w-4 h-4 rounded bg-red-500/60"></div> <span class="text-gray-300">Hot (Equator)</span></div> <div class="flex items-center gap-2"><div class="w-4 h-4 rounded bg-green-500/60"></div> <span class="text-gray-300">Temperate</span></div> <div class="flex items-center gap-2"><div class="w-4 h-4 rounded bg-blue-500/60"></div> <span class="text-gray-300">Cold (Poles)</span></div></div></div>` : ``}  ${showMoisture ? `<div data-svelte-h="svelte-tqxgxw"><h4 class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider">Moisture</h4> <div class="space-y-2"><div class="flex items-center gap-2"><div class="w-4 h-4 rounded bg-blue-600/80"></div> <span class="text-gray-300">Wet / Rainforest</span></div> <div class="flex items-center gap-2"><div class="w-4 h-4 rounded bg-blue-400/50"></div> <span class="text-gray-300">Seasonal</span></div> <div class="flex items-center gap-2"><div class="w-4 h-4 rounded bg-white/20"></div> <span class="text-gray-300">Arid / Desert</span></div></div></div>` : ``}  ${showTectonics ? `<div><h4 class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider" data-svelte-h="svelte-11jn3s4">Tectonics</h4> <div class="space-y-1"><div class="text-gray-500 italic mb-2" data-svelte-h="svelte-3vjcfn">Plates rely on auto-generated colors.</div> ${each(tectonicLegend, (item) => {
    return `<div class="flex items-center gap-2"><div class="w-4 h-4 rounded shadow-sm border border-black/20" style="${"background-color: " + escape(item.color, true)}"></div> <div class="flex-1"><div class="text-gray-300">${escape(item.label)}</div> <div class="text-gray-600 text-[10px]">${escape(item.desc)} </div></div> </div>`;
  })}</div></div>` : ``}  ${showBiomes ? `<div><h4 class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider" data-svelte-h="svelte-em46jn">Biomes</h4> <div class="space-y-1">${each(biomeColors, (item) => {
    return `<div class="flex items-center gap-2"><div class="w-4 h-4 rounded shadow-sm border border-black/20" style="${"background-color: " + escape(item.color, true)}"></div> <div class="text-gray-300">${escape(item.label)}</div> </div>`;
  })}</div></div>` : ``}  ${showTopography ? `<div><h4 class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider" data-svelte-h="svelte-henxth">Topography</h4> <div class="space-y-1">${each(terrainColors, (item) => {
    return `<div class="flex items-center gap-2"><div class="w-4 h-4 rounded shadow-sm border border-black/20" style="${"background-color: " + escape(item.color, true)}"></div> <div class="flex-1"><div class="text-gray-300">${escape(item.label)}</div> <div class="text-gray-600 text-[10px]">${escape(item.desc)} </div></div> </div>`;
  })}</div></div>  <div class="mt-4"><h4 class="font-bold text-gray-400 mb-2 uppercase text-[10px] tracking-wider" data-svelte-h="svelte-vni9ow">Ocean</h4> <div class="space-y-1">${each(oceanColors, (item) => {
    return `<div class="flex items-center gap-2"><div class="w-4 h-4 rounded shadow-sm border border-black/20" style="${"background-color: " + escape(item.color, true)}"></div> <div class="flex-1"><div class="text-gray-300">${escape(item.label)}</div> <div class="text-gray-600 text-[10px]">${escape(item.desc)} </div></div> </div>`;
  })}</div></div>` : ``}`}</div>` : ``}</div>`;
});
const MapOverlayCanvas = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  createEventDispatcher();
  let { width } = $$props;
  let { height } = $$props;
  let { gridWidth } = $$props;
  let { gridHeight } = $$props;
  let { activeLayers = /* @__PURE__ */ new Set() } = $$props;
  let { overlayData = {} } = $$props;
  let { tectonicsData = null } = $$props;
  let { plateInfo = [] } = $$props;
  let { mineralsData = null } = $$props;
  let { showTectonics = false } = $$props;
  let { showMinerals = false } = $$props;
  let { cameraX = 0.5 } = $$props;
  let { cameraY = 0.5 } = $$props;
  let { zoom = 1 } = $$props;
  let canvas;
  if ($$props.width === void 0 && $$bindings.width && width !== void 0) $$bindings.width(width);
  if ($$props.height === void 0 && $$bindings.height && height !== void 0) $$bindings.height(height);
  if ($$props.gridWidth === void 0 && $$bindings.gridWidth && gridWidth !== void 0) $$bindings.gridWidth(gridWidth);
  if ($$props.gridHeight === void 0 && $$bindings.gridHeight && gridHeight !== void 0) $$bindings.gridHeight(gridHeight);
  if ($$props.activeLayers === void 0 && $$bindings.activeLayers && activeLayers !== void 0) $$bindings.activeLayers(activeLayers);
  if ($$props.overlayData === void 0 && $$bindings.overlayData && overlayData !== void 0) $$bindings.overlayData(overlayData);
  if ($$props.tectonicsData === void 0 && $$bindings.tectonicsData && tectonicsData !== void 0) $$bindings.tectonicsData(tectonicsData);
  if ($$props.plateInfo === void 0 && $$bindings.plateInfo && plateInfo !== void 0) $$bindings.plateInfo(plateInfo);
  if ($$props.mineralsData === void 0 && $$bindings.mineralsData && mineralsData !== void 0) $$bindings.mineralsData(mineralsData);
  if ($$props.showTectonics === void 0 && $$bindings.showTectonics && showTectonics !== void 0) $$bindings.showTectonics(showTectonics);
  if ($$props.showMinerals === void 0 && $$bindings.showMinerals && showMinerals !== void 0) $$bindings.showMinerals(showMinerals);
  if ($$props.cameraX === void 0 && $$bindings.cameraX && cameraX !== void 0) $$bindings.cameraX(cameraX);
  if ($$props.cameraY === void 0 && $$bindings.cameraY && cameraY !== void 0) $$bindings.cameraY(cameraY);
  if ($$props.zoom === void 0 && $$bindings.zoom && zoom !== void 0) $$bindings.zoom(zoom);
  return `<canvas${add_attribute("width", width, 0)}${add_attribute("height", height, 0)} class="absolute inset-0 pointer-events-auto"${add_attribute("this", canvas, 0)}></canvas>`;
});
Effect.ShadersStore["displacementVertexShader"] = `
    precision highp float;

    // Attributes
    attribute vec3 position;
    attribute vec3 normal;
    attribute vec2 uv;

    // Uniforms
    uniform mat4 world;
    uniform mat4 viewProjection;
    uniform sampler2D heightmap;
    uniform float scale;

    // Varyings
    varying float vHeight;
    varying vec3 vNormal;
    varying vec3 vPosition;
    varying vec2 vUV; // We still pass UV for material textures that might use it (though we should strictly use calc'd UV)

    void main(void) {
        vNormal = normal;
        
        // We will calculate UVs in fragment shader for the diffuse, 
        // BUT we need UVs here for sampling the heightmap if we want displacement.
        // HOWEVER, standard UVs pinch at poles.
        // Ideally, heightmap would also be sampled via 3D position, but vertex texture fetch isn't always cheap/easy with gradients.
        // For now, we will stick to using the Mesh UVs for vertex displacement (geometry is dense there anyway so pinching is less visible in height than in texture),
        // or we could try to compute spherical UVs here too.
        // For Icosphere, 'uv' attribute might not be what we want.
        // Let's compute a basic spherical UV here for the vertex displacement fetch.
        
        vec3 p_norm = normalize(position);
        float long_u = atan(p_norm.z, p_norm.x) / (2.0 * 3.14159265359) + 0.5;
        float lat_v = asin(p_norm.y) / 3.14159265359 + 0.5;
        vec2 sphericalUV = vec2(long_u, lat_v);
        vUV = sphericalUV; // Use this for height fetch

        // Sample height
        // Assuming heightmap is normalized 0.0-1.0
        float h = texture2D(heightmap, sphericalUV).r;
        vHeight = h;
        
        // Displace along normal
        // scale determines max height in world units
        vec3 p = position + normal * (h * scale);
        vPosition = p;
        
        gl_Position = viewProjection * world * vec4(p, 1.0);
    }
`;
Effect.ShadersStore["displacementFragmentShader"] = `
    #extension GL_OES_standard_derivatives : enable
    precision highp float;

    // Varyings
    varying float vHeight;
    varying vec3 vNormal;
    varying vec3 vPosition;

    // Uniforms
    uniform vec3 color; // Base color (fallback)
    uniform vec3 lightDirection; // Dynamic sun direction (normalized)
    uniform float seaLevel; // Normalized sea level (0-1 in heightmap space)
    uniform float minElevation; // Minimum elevation in meters
    uniform float maxElevation; // Maximum elevation in meters
    
    // Data textures
    uniform sampler2D diffuseTex;  // The actual color map (replacing procedural rock)
    uniform sampler2D specularTex; // Specular map (water shininess)
    uniform sampler2D materialTex; // R=hardness, G=continental, B=sediment
    uniform sampler2D iceTex;      // R=ice thickness
    uniform sampler2D normalTex;   // Normal map for 3D shadows
    
    uniform bool hasDiffuseTex;
    uniform bool hasSpecularTex;
    uniform bool hasMaterialTex;
    uniform bool hasIceTex;
    uniform bool hasNormalTex;

    const float PI = 3.14159265359;

    // Calculate Spherical UV from 3D Position
    vec2 calculateUV(vec3 p) {
        vec3 v = normalize(p);
        float u = (atan(v.z, v.x) / (2.0 * PI)) + 0.5;
        float v_coord = (asin(v.y) / PI) + 0.5;
        return vec2(u, v_coord);
    }

    // Color palette for rock types based on hardness
    vec3 getRockColor(float hardness, bool isContinental) {
        if (!isContinental) {
            // Oceanic crust: basalt (dark grey)
            return vec3(0.235, 0.235, 0.255);
        }
        
        // Continental crust: sedimentary -> granite based on hardness
        vec3 sandstone = vec3(0.706, 0.549, 0.392);  // Soft sedimentary
        vec3 granite = vec3(0.627, 0.588, 0.569);    // Medium metamorphic
        vec3 hardRock = vec3(0.4, 0.4, 0.42);        // Hard crystalline
        
        if (hardness < 0.5) {
            return mix(sandstone, granite, hardness * 2.0);
        }
        return mix(granite, hardRock, (hardness - 0.5) * 2.0);
    }

    // Sediment overlay color
    vec3 getSedimentColor(float sediment) {
        vec3 sedimentTan = vec3(0.706, 0.627, 0.471);
        return sedimentTan;
    }

    // Ice/snow color based on thickness
    vec3 getIceColor(float thickness) {
        vec3 frost = vec3(0.9, 0.91, 0.93);       // Light frost
        vec3 snow = vec3(0.95, 0.95, 0.97);       // Snow white
        vec3 glacier = vec3(0.75, 0.85, 0.92);    // Blue-white glacier
        
        if (thickness < 0.3) {
            return mix(frost, snow, thickness / 0.3);
        }
        return mix(snow, glacier, (thickness - 0.3) / 0.7);
    }

    // Satellite-style bathymetric (underwater terrain visible through water)
    vec3 getBathymetricColor(vec3 terrainColor, float depthFactor) {
        vec3 shallowWater = vec3(0.314, 0.706, 0.863);  // Turquoise
        vec3 deepWater = vec3(0.02, 0.08, 0.2);         // Deep navy
        
        // Water tint based on depth
        vec3 waterTint = mix(shallowWater, deepWater, depthFactor);
        
        // Blend terrain with water - shallow shows terrain, deep obscures
        float visibility = 1.0 - min(depthFactor * 1.5, 0.85);
        return mix(waterTint, terrainColor * 0.7, visibility);
    }

    void main(void) {
        // Calculate UVs from 3D position
        vec2 uv = calculateUV(vPosition);

        // --- SEAM FIX ---
        // Calculate analytic derivatives
        vec2 uv_dx = dFdx(uv);
        vec2 uv_dy = dFdy(uv);
        
        // Check for wrapping in U (longitude)
        // If the derivative is too large, it means we wrapped across the Date Line
        if (abs(uv_dx.x) > 0.5) {
            uv_dx.x = -sign(uv_dx.x) * (1.0 - abs(uv_dx.x));
        }
        if (abs(uv_dy.x) > 0.5) {
            uv_dy.x = -sign(uv_dy.x) * (1.0 - abs(uv_dy.x));
        }
        // ----------------

        // Calculate perturbed normal
        vec3 normal = normalize(vNormal);
        
        if (hasNormalTex) {
            // Sample normal map (tangent space)
            // Use textureGrad for seam-safe sampling
            vec3 mapN;
            #ifdef GL_OES_standard_derivatives
                mapN = texture2DGradEXT(normalTex, uv, uv_dx, uv_dy).rgb * 2.0 - 1.0;
            #else
                mapN = texture2D(normalTex, uv).rgb * 2.0 - 1.0;
            #endif
            
            // Calculate TBN matrix
            // Tangent (East-West)
            vec3 T = normalize(cross(vec3(0.0, 1.0, 0.0), normal));
            if (length(T) < 0.001) T = vec3(1.0, 0.0, 0.0); // Pole fallback
            
            // Bitangent (North-South)
            vec3 B = normalize(cross(normal, T));
            
            // Transform to World Space
            normal = normalize(T * mapN.x + B * mapN.y + normal * mapN.z);
        }

        // Lighting
        float ndotl = max(0.0, dot(normal, lightDirection));
        vec3 viewDir = normalize(vec3(0.0, 0.0, 0.0) - vPosition); // Simplified View Dir approximation or pass camera pos
        vec3 halfVector = normalize(lightDirection + viewDir);
        float NdotH = max(0.0, dot(normal, halfVector));
        
        // --- BASE SURFACE COLOR ---
        vec3 surfaceColor = vec3(0.5); // Default grey
        
        if (hasDiffuseTex) {
            // If we have a diffuse texture (real map), use it!
            #ifdef GL_OES_standard_derivatives
                surfaceColor = texture2DGradEXT(diffuseTex, uv, uv_dx, uv_dy).rgb;
            #else
                surfaceColor = texture2D(diffuseTex, uv).rgb;
            #endif
        } else {
            // Procedural Coloring (Fallback or Data View)
            // Sample material data if available
            float hardness = 0.5;
            bool isContinental = true;
            float sediment = 0.0;
            
            if (hasMaterialTex) {
                vec4 matData;
                #ifdef GL_OES_standard_derivatives
                    matData = texture2DGradEXT(materialTex, uv, uv_dx, uv_dy);
                #else
                    matData = texture2D(materialTex, uv);
                #endif
                hardness = matData.r;
                isContinental = matData.g > 0.5;
                sediment = matData.b;
            }
            
            // Get base rock color from material data
            surfaceColor = getRockColor(hardness, isContinental);
            
            // Apply sediment overlay
            if (sediment > 0.1) {
                vec3 sedimentCol = getSedimentColor(sediment);
                surfaceColor = mix(surfaceColor, sedimentCol, min(sediment * 1.5, 0.7));
            }
            
            // Check if underwater
            if (vHeight < seaLevel) {
                // Underwater - satellite-style bathymetry showing underwater terrain
                float depthFactor = (seaLevel - vHeight) / max(seaLevel, 0.001);
                surfaceColor = getBathymetricColor(surfaceColor, depthFactor);
            }
        }
        
        // --- OVERLAYS (Ice, etc.) ---
        // Apply ice overlay
        if (hasIceTex) {
            float ice;
            #ifdef GL_OES_standard_derivatives
                ice = texture2DGradEXT(iceTex, uv, uv_dx, uv_dy).r;
            #else
                ice = texture2D(iceTex, uv).r;
            #endif
            
            if (ice > 0.05) {
                vec3 iceCol = getIceColor(ice);
                surfaceColor = mix(surfaceColor, iceCol, min(ice * 2.0, 0.95));
            }
        }
        
        // SPECULAR HIGHLIGHT
        float specularPower = 30.0;
        float specularIntensity = 0.0;
        if (hasSpecularTex) {
             #ifdef GL_OES_standard_derivatives
                specularIntensity = texture2DGradEXT(specularTex, uv, uv_dx, uv_dy).r;
            #else
                specularIntensity = texture2D(specularTex, uv).r;
            #endif
        }
        vec3 specular = vec3(pow(NdotH, specularPower)) * specularIntensity;
        
        vec3 finalColor = surfaceColor * (0.25 + 0.75 * ndotl) + specular; // Ambient + Diffuse + Specular

        gl_FragColor = vec4(finalColor, 1.0);
    }
`;
Effect.ShadersStore["moltenPlanetVertexShader"] = `
    precision highp float;
    
    // Attributes
    attribute vec3 position;
    attribute vec3 normal;
    attribute vec2 uv;
    
    // Uniforms
    uniform mat4 worldViewProjection;
    uniform mat4 world;
    uniform float time;
    
    // Varying
    varying vec2 vUV;
    varying vec3 vPosition;
    varying vec3 vNormal;
    
    void main(void) {
        vec3 positionUpdated = position;
        
        // Slight undulation for "breathing" planet effect
        float undulation = sin(time * 0.5 + position.y * 2.0) * 0.01;
        positionUpdated += normal * undulation;
        
        gl_Position = worldViewProjection * vec4(positionUpdated, 1.0);
        
        vUV = uv;
        vPosition = position;
        vNormal = normalize(vec3(world * vec4(normal, 0.0)));
    }
`;
Effect.ShadersStore["moltenPlanetFragmentShader"] = `
    precision highp float;
    
    // Varying
    varying vec2 vUV;
    varying vec3 vPosition;
    varying vec3 vNormal;
    
    // Uniforms
    uniform float time;
    
    // Noise function
    vec3 mod289(vec3 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
    vec4 mod289(vec4 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
    vec4 permute(vec4 x) { return mod289(((x*34.0)+1.0)*x); }
    vec4 taylorInvSqrt(vec4 r) { return 1.79284291400159 - 0.85373472095314 * r; }
    
    float snoise(vec3 v) {
        const vec2  C = vec2(1.0/6.0, 1.0/3.0) ;
        const vec4  D = vec4(0.0, 0.5, 1.0, 2.0);
        
        // First corner
        vec3 i  = floor(v + dot(v, C.yyy) );
        vec3 x0 = v - i + dot(i, C.xxx) ;
        
        // Other corners
        vec3 g = step(x0.yzx, x0.xyz);
        vec3 l = 1.0 - g;
        vec3 i1 = min( g.xyz, l.zxy );
        vec3 i2 = max( g.xyz, l.zxy );
        
        //   x0 = x0 - 0.0 + 0.0 * C.xxx;
        //   x1 = x0 - i1  + 1.0 * C.xxx;
        //   x2 = x0 - i2  + 2.0 * C.xxx;
        //   x3 = x0 - 1.0 + 3.0 * C.xxx;
        vec3 x1 = x0 - i1 + C.xxx;
        vec3 x2 = x0 - i2 + C.yyy; // 2.0*C.x = 1/3 = C.y
        vec3 x3 = x0 - D.yyy;      // -1.0+3.0*C.x = -0.5 = -D.y
        
        // Permutations
        i = mod289(i);
        vec4 p = permute( permute( permute(
                  i.z + vec4(0.0, i1.z, i2.z, 1.0 ))
                + i.y + vec4(0.0, i1.y, i2.y, 1.0 ))
                + i.x + vec4(0.0, i1.x, i2.x, 1.0 ));
                
        // Gradients: 7x7 points over a square, mapped onto an octahedron.
        // The ring size 17*17 = 289 is close to a multiple of 49 (49*6 = 294)
        float n_ = 0.142857142857; // 1.0/7.0
        vec3  ns = n_ * D.wyz - D.xzx;
        
        vec4 j = p - 49.0 * floor(p * ns.z * ns.z);  //  mod(p,7*7)
        
        vec4 x_ = floor(j * ns.z);
        vec4 y_ = floor(j - 7.0 * x_ );    // mod(j,N)
        
        vec4 x = x_ *ns.x + ns.yyyy;
        vec4 y = y_ *ns.x + ns.yyyy;
        vec4 h = 1.0 - abs(x) - abs(y);
        
        vec4 b0 = vec4( x.xy, y.xy );
        vec4 b1 = vec4( x.zw, y.zw );
        
        //vec4 s0 = vec4(lessThan(b0,0.0))*2.0 - 1.0;
        //vec4 s1 = vec4(lessThan(b1,0.0))*2.0 - 1.0;
        vec4 s0 = floor(b0)*2.0 + 1.0;
        vec4 s1 = floor(b1)*2.0 + 1.0;
        vec4 sh = -step(h, vec4(0.0));
        
        vec4 a0 = b0.xzyw + s0.xzyw*sh.xxyy ;
        vec4 a1 = b1.xzyw + s1.xzyw*sh.zzww ;
        
        vec3 p0 = vec3(a0.xy,h.x);
        vec3 p1 = vec3(a0.zw,h.y);
        vec3 p2 = vec3(a1.xy,h.z);
        vec3 p3 = vec3(a1.zw,h.w);
        
        //Normalise gradients
        vec4 norm = taylorInvSqrt(vec4(dot(p0,p0), dot(p1,p1), dot(p2, p2), dot(p3,p3)));
        p0 *= norm.x;
        p1 *= norm.y;
        p2 *= norm.z;
        p3 *= norm.w;
        
        // Mix final noise value
        vec4 m = max(0.6 - vec4(dot(x0,x0), dot(x1,x1), dot(x2,x2), dot(x3,x3)), 0.0);
        m = m * m;
        return 42.0 * dot( m*m, vec4( dot(p0,x0), dot(p1,x1),
                                      dot(p2,x2), dot(p3,x3) ) );
    }
    
    // Fractal Brownian Motion
    float fbm(vec3 p) {
        float total = 0.0;
        float amp = 0.5;
        float freq = 1.0;
        for(int i = 0; i < 4; i++) {
            total += snoise(p * freq) * amp;
            freq *= 2.0;
            amp *= 0.5;
        }
        return total;
    }

    void main(void) {
        // Base coordinate with rotation over time
        vec3 coord = vPosition * 2.0;
        coord.x += time * 0.1; 
        
        // Generate noise layers
        float n1 = fbm(coord);
        float n2 = fbm(coord * 2.0 + vec3(time * 0.2));
        float noise = n1 * 0.6 + n2 * 0.4; // Combined noise
        
        // Lava colors
        vec3 darkMagma = vec3(0.1, 0.0, 0.0);
        vec3 redMagma = vec3(0.5, 0.0, 0.0);
        vec3 orangeLava = vec3(1.0, 0.3, 0.0);
        vec3 brightYellow = vec3(1.0, 0.9, 0.2);
        
        // Color mixing based on noise
        vec3 color;
        if (noise < 0.2) {
            color = mix(darkMagma, redMagma, smoothstep(-0.5, 0.2, noise));
        } else if (noise < 0.5) {
            color = mix(redMagma, orangeLava, smoothstep(0.2, 0.5, noise));
        } else {
            color = mix(orangeLava, brightYellow, smoothstep(0.5, 1.0, noise));
        }
        
        // Add "crust" effect
        float crustMap = smoothstep(0.4, 0.45, fbm(vPosition * 4.0 - vec3(time * 0.05)));
        color = mix(color, darkMagma * 0.5, crustMap);
        
        // Glow/Emission
        float brightness = dot(color, vec3(0.299, 0.587, 0.114));
        vec3 emission = color * (brightness * 2.0); // Brighter areas glow more
        
        gl_FragColor = vec4(color + emission * 0.5, 1.0);
    }
`;
Effect.ShadersStore["atmosphereVertexShader"] = `
    precision highp float;

    attribute vec3 position;
    
    uniform mat4 world;
    uniform mat4 viewProjection;

    varying vec3 vPosition;

    void main(void) {
        vPosition = position;
        gl_Position = viewProjection * world * vec4(position, 1.0);
    }
`;
Effect.ShadersStore["atmosphereFragmentShader"] = `
    precision highp float;

    varying vec3 vPosition;

    uniform vec3 sunDirection;
    uniform vec3 zenithColor;      // Color at top of sky
    uniform vec3 horizonColor;     // Color at horizon
    uniform vec3 groundColor;      // Color below horizon
    uniform vec3 sunColor;
    uniform float sunSize;
    uniform float atmosphereHeight;
    uniform float sunIntensity;

    void main(void) {
        vec3 viewDir = normalize(vPosition);
        
        // Calculate altitude angle (-1 = down, 0 = horizon, 1 = up)
        float altitude = viewDir.y;
        
        // Sky gradient based on altitude
        vec3 skyColor;
        
        if (altitude > 0.0) {
            // Above horizon - blend zenith to horizon
            float t = pow(altitude, 0.5); // Non-linear curve for more horizon color
            skyColor = mix(horizonColor, zenithColor, t);
        } else {
            // Below horizon - blend to ground color
            float t = clamp(-altitude * 3.0, 0.0, 1.0);
            skyColor = mix(horizonColor, groundColor, t);
        }
        
        // Sun disc
        float sunAngle = dot(viewDir, sunDirection);
        float sunDisc = smoothstep(1.0 - sunSize * 0.01, 1.0, sunAngle);
        
        // Sun glow (halo around sun)
        float sunGlow = pow(max(0.0, sunAngle), 8.0) * 0.5;
        
        // Atmospheric scattering near horizon (brighter near sun)
        float horizonGlow = 1.0 - abs(altitude);
        horizonGlow = pow(horizonGlow, 3.0);
        float nearSun = max(0.0, sunAngle);
        vec3 horizonScatter = horizonColor * horizonGlow * (1.0 + nearSun * 2.0);
        
        // Combine
        vec3 finalColor = skyColor + horizonScatter * 0.3;
        finalColor += sunColor * (sunDisc + sunGlow) * sunIntensity;
        
        gl_FragColor = vec4(finalColor, 1.0);
    }
`;
Effect.ShadersStore["waterSurfaceVertexShader"] = `
    precision highp float;

    attribute vec3 position;
    attribute vec3 normal;
    attribute vec2 uv;

    uniform mat4 world;
    uniform mat4 viewProjection;
    uniform float time;
    uniform float waveHeight;
    uniform float waveFrequency;

    varying vec2 vUV;
    varying vec3 vNormal;
    varying vec3 vWorldPos;
    varying float vWaveOffset;

    void main(void) {
        vUV = uv;
        vNormal = normal;
        
        // Simple wave animation
        float wave = sin(uv.x * waveFrequency + time) * cos(uv.y * waveFrequency + time * 0.7);
        vWaveOffset = wave * waveHeight;
        
        vec3 displaced = position + normal * vWaveOffset;
        vWorldPos = (world * vec4(displaced, 1.0)).xyz;
        
        gl_Position = viewProjection * world * vec4(displaced, 1.0);
    }
`;
Effect.ShadersStore["waterSurfaceFragmentShader"] = `
    precision highp float;

    varying vec2 vUV;
    varying vec3 vNormal;
    varying vec3 vWorldPos;
    varying float vWaveOffset;

    uniform vec3 lightDirection;
    uniform vec3 waterColor;
    uniform vec3 deepWaterColor;
    uniform float transparency;
    uniform float specularPower;
    uniform vec3 cameraPosition;

    void main(void) {
        // Fresnel effect for reflection vs refraction
        vec3 viewDir = normalize(cameraPosition - vWorldPos);
        float fresnel = pow(1.0 - max(0.0, dot(viewDir, normalize(vNormal))), 2.0);
        
        // Wave-perturbed normal for specular
        vec3 perturbedNormal = normalize(vNormal + vec3(vWaveOffset * 0.5, 0.0, vWaveOffset * 0.3));
        
        // Specular highlight (sun glint)
        float spec = pow(max(0.0, dot(reflect(-lightDirection, perturbedNormal), viewDir)), specularPower);
        
        // Base water color with depth variation
        vec3 baseColor = mix(waterColor, deepWaterColor, 0.3);
        
        // Add specular
        vec3 finalColor = baseColor + vec3(spec);
        
        // Increase opacity at edges (fresnel)
        float alpha = mix(transparency, 1.0, fresnel);
        
        gl_FragColor = vec4(finalColor, alpha);
    }
`;
Effect.ShadersStore["underwaterPostProcessFragmentShader"] = `
    precision highp float;

    varying vec2 vUV;
    
    uniform sampler2D textureSampler;
    uniform float time;
    uniform float depth;
    uniform vec3 waterColor;
    uniform float distortionStrength;
    uniform float causticsStrength;

    void main(void) {
        // Distortion effect
        vec2 distortedUV = vUV;
        distortedUV.x += sin(vUV.y * 20.0 + time * 2.0) * distortionStrength;
        distortedUV.y += cos(vUV.x * 20.0 + time * 1.5) * distortionStrength;
        
        vec4 color = texture2D(textureSampler, distortedUV);
        
        // Caustics pattern
        float caustics = sin(vUV.x * 30.0 + time) * sin(vUV.y * 30.0 + time * 0.8);
        caustics = caustics * caustics * causticsStrength;
        
        // Apply water color tint based on depth
        float depthFactor = clamp(depth * 0.5, 0.0, 0.8);
        vec3 tintedColor = mix(color.rgb, waterColor, depthFactor);
        
        // Add caustics
        tintedColor += vec3(caustics);
        
        // Reduce contrast underwater
        tintedColor = mix(tintedColor, vec3(0.5), depthFactor * 0.3);
        
        gl_FragColor = vec4(tintedColor, 1.0);
    }
`;
const MAX_MESSAGES = 1e3;
class CircularBuffer {
  buffer;
  maxSize;
  startIndex = 0;
  size = 0;
  constructor(maxSize = MAX_MESSAGES) {
    this.maxSize = maxSize;
    this.buffer = new Array(maxSize);
  }
  // Add message to buffer (O(1))
  push(item) {
    const index = (this.startIndex + this.size) % this.maxSize;
    this.buffer[index] = item;
    if (this.size < this.maxSize) {
      this.size++;
    } else {
      this.startIndex = (this.startIndex + 1) % this.maxSize;
    }
  }
  // Get all messages in order (O(N) where N = min(size, maxSize))
  getAll() {
    const result = [];
    for (let i = 0; i < this.size; i++) {
      const index = (this.startIndex + i) % this.maxSize;
      result.push(this.buffer[index]);
    }
    return result;
  }
  // Get recent N messages (O(N))
  getRecent(count) {
    const actualCount = Math.min(count, this.size);
    const result = [];
    for (let i = this.size - actualCount; i < this.size; i++) {
      const index = (this.startIndex + i) % this.maxSize;
      result.push(this.buffer[index]);
    }
    return result;
  }
  // Get message by ID (O(N) - for searching)
  findById(id) {
    for (let i = 0; i < this.size; i++) {
      const index = (this.startIndex + i) % this.maxSize;
      if (this.buffer[index].id === id) {
        return this.buffer[index];
      }
    }
    return void 0;
  }
  // Clear all messages (O(1))
  clear() {
    this.startIndex = 0;
    this.size = 0;
  }
  // Get current size
  getSize() {
    return this.size;
  }
  // Get maximum capacity
  getCapacity() {
    return this.maxSize;
  }
  // Check if buffer is full
  isFull() {
    return this.size === this.maxSize;
  }
}
function detectDeviceType() {
  if (typeof window === "undefined") return "desktop";
  const hasTouch = "ontouchstart" in window || navigator.maxTouchPoints > 0;
  const isSmallScreen = window.innerWidth < 769;
  const isMobileUA = /Android|iPhone|iPad|iPod/i.test(navigator.userAgent);
  if (isSmallScreen && (hasTouch || isMobileUA)) return "mobile";
  if (/iPad|Tablet/i.test(navigator.userAgent)) return "tablet";
  return "desktop";
}
function getDefaultInterfaceMode() {
  const device = detectDeviceType();
  return device === "mobile" ? "TEXT" : "VISUAL";
}
function loadPreferredMode() {
  if (typeof localStorage === "undefined") return "auto";
  const saved = localStorage.getItem("tw-interface-mode");
  if (saved === "TEXT" || saved === "VISUAL") return saved;
  return "auto";
}
function getInitialInterfaceMode() {
  const preferred = loadPreferredMode();
  if (preferred !== "auto") return preferred;
  return getDefaultInterfaceMode();
}
const initialUIState = {
  layoutMode: "desktop",
  screenWidth: 1024,
  interfaceMode: getInitialInterfaceMode(),
  preferredMode: loadPreferredMode(),
  activePanel: "none",
  isSidebarOpen: false
};
const uiState = writable(initialUIState);
const gameOutputBuffer = new CircularBuffer(1e3);
const gameOutput = writable([]);
const isMobile = derived(uiState, ($ui) => $ui.layoutMode === "mobile");
const interfaceMode = derived(uiState, ($ui) => $ui.interfaceMode);
const isTextMode = derived(uiState, ($ui) => $ui.interfaceMode === "TEXT");
derived(uiState, ($ui) => $ui.interfaceMode === "VISUAL");
function addGameMessage(message) {
  gameOutputBuffer.push(message);
  gameOutput.set(gameOutputBuffer.getAll());
}
const WorldController = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let { scene } = $$props;
  let { globeTextureBlob = null } = $$props;
  let { globeHeightmapBlob = null } = $$props;
  let { materialBlob = null } = $$props;
  let { iceBlob = null } = $$props;
  let { normalMapBlob = null } = $$props;
  let { seaLevel = 0 } = $$props;
  let { maxElevation = 8848 } = $$props;
  let { minElevation = -11e3 } = $$props;
  let { satellites = [] } = $$props;
  let { simulationSpeed = 1 } = $$props;
  let { onSendCommand = null } = $$props;
  let moonMeshes = [];
  let moonOrbitNodes = [];
  let isPaused = false;
  let unsubscribeTextMode = null;
  function pauseRendering() {
    if (isPaused) return;
    console.log("[WorldController] Pausing updates (TEXT mode)");
    isPaused = true;
  }
  function resumeRendering() {
    if (!isPaused) return;
    console.log("[WorldController] Resuming updates (VISUAL mode)");
    isPaused = false;
    performance.now();
  }
  onDestroy(() => {
    moonMeshes.forEach((m) => m.dispose());
    moonOrbitNodes.forEach((n) => n.dispose());
    moonMeshes = [];
    moonOrbitNodes = [];
  });
  if ($$props.scene === void 0 && $$bindings.scene && scene !== void 0) $$bindings.scene(scene);
  if ($$props.globeTextureBlob === void 0 && $$bindings.globeTextureBlob && globeTextureBlob !== void 0) $$bindings.globeTextureBlob(globeTextureBlob);
  if ($$props.globeHeightmapBlob === void 0 && $$bindings.globeHeightmapBlob && globeHeightmapBlob !== void 0) $$bindings.globeHeightmapBlob(globeHeightmapBlob);
  if ($$props.materialBlob === void 0 && $$bindings.materialBlob && materialBlob !== void 0) $$bindings.materialBlob(materialBlob);
  if ($$props.iceBlob === void 0 && $$bindings.iceBlob && iceBlob !== void 0) $$bindings.iceBlob(iceBlob);
  if ($$props.normalMapBlob === void 0 && $$bindings.normalMapBlob && normalMapBlob !== void 0) $$bindings.normalMapBlob(normalMapBlob);
  if ($$props.seaLevel === void 0 && $$bindings.seaLevel && seaLevel !== void 0) $$bindings.seaLevel(seaLevel);
  if ($$props.maxElevation === void 0 && $$bindings.maxElevation && maxElevation !== void 0) $$bindings.maxElevation(maxElevation);
  if ($$props.minElevation === void 0 && $$bindings.minElevation && minElevation !== void 0) $$bindings.minElevation(minElevation);
  if ($$props.satellites === void 0 && $$bindings.satellites && satellites !== void 0) $$bindings.satellites(satellites);
  if ($$props.simulationSpeed === void 0 && $$bindings.simulationSpeed && simulationSpeed !== void 0) $$bindings.simulationSpeed(simulationSpeed);
  if ($$props.onSendCommand === void 0 && $$bindings.onSendCommand && onSendCommand !== void 0) $$bindings.onSendCommand(onSendCommand);
  {
    if (typeof window !== "undefined") {
      unsubscribeTextMode?.();
      unsubscribeTextMode = isTextMode.subscribe((inTextMode) => {
        if (inTextMode && !isPaused) {
          pauseRendering();
        } else if (!inTextMode && isPaused) {
          resumeRendering();
        }
      });
    }
  }
  return `${slots.default ? slots.default({}) : ``}`;
});
const css$5 = {
  code: ".scene-canvas-container.svelte-vqxrxo{position:relative;overflow:hidden}.scene-canvas.svelte-vqxrxo{display:block;width:100%;height:100%;outline:none;touch-action:none}",
  map: '{"version":3,"file":"SceneCanvas.svelte","sources":["SceneCanvas.svelte"],"sourcesContent":["<script lang=\\"ts\\">import { onMount, onDestroy, createEventDispatcher } from \\"svelte\\";\\nexport let width = \\"100%\\";\\nexport let height = \\"100%\\";\\nconst dispatch = createEventDispatcher();\\nlet canvas;\\nlet container;\\nlet resizeObserver = null;\\nonMount(() => {\\n  if (!canvas) return;\\n  dispatch(\\"canvasReady\\", canvas);\\n  resizeObserver = new ResizeObserver((entries) => {\\n    for (const entry of entries) {\\n      const { width: width2, height: height2 } = entry.contentRect;\\n      canvas.width = width2;\\n      canvas.height = height2;\\n      dispatch(\\"resize\\", { width: width2, height: height2 });\\n    }\\n  });\\n  resizeObserver.observe(container);\\n});\\nonDestroy(() => {\\n  resizeObserver?.disconnect();\\n});\\nexport function getCanvas() {\\n  return canvas;\\n}\\n<\/script>\\n\\n/** * SceneCanvas.svelte * Owns the canvas element and provides it to\\nSceneManager. * Part of the Engine Hoist refactor - separates canvas from scene\\nlogic. */\\n<div\\n    bind:this={container}\\n    class=\\"scene-canvas-container\\"\\n    style=\\"width: {typeof width === \'number\'\\n        ? `${width}px`\\n        : width}; height: {typeof height === \'number\'\\n        ? `${height}px`\\n        : height};\\"\\n>\\n    <canvas bind:this={canvas} class=\\"scene-canvas\\" tabindex=\\"0\\"></canvas>\\n</div>\\n\\n<style>\\n    .scene-canvas-container {\\n        position: relative;\\n        overflow: hidden;\\n    }\\n\\n    .scene-canvas {\\n        display: block;\\n        width: 100%;\\n        height: 100%;\\n        outline: none;\\n        touch-action: none; /* Prevent browser zoom/pan on touch */\\n    }\\n</style>\\n"],"names":[],"mappings":"AA4CI,qCAAwB,CACpB,QAAQ,CAAE,QAAQ,CAClB,QAAQ,CAAE,MACd,CAEA,2BAAc,CACV,OAAO,CAAE,KAAK,CACd,KAAK,CAAE,IAAI,CACX,MAAM,CAAE,IAAI,CACZ,OAAO,CAAE,IAAI,CACb,YAAY,CAAE,IAClB"}'
};
const SceneCanvas = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let { width = "100%" } = $$props;
  let { height = "100%" } = $$props;
  createEventDispatcher();
  let canvas;
  let container;
  onDestroy(() => {
  });
  function getCanvas() {
    return canvas;
  }
  if ($$props.width === void 0 && $$bindings.width && width !== void 0) $$bindings.width(width);
  if ($$props.height === void 0 && $$bindings.height && height !== void 0) $$bindings.height(height);
  if ($$props.getCanvas === void 0 && $$bindings.getCanvas && getCanvas !== void 0) $$bindings.getCanvas(getCanvas);
  $$result.css.add(css$5);
  return `/** * SceneCanvas.svelte * Owns the canvas element and provides it to
SceneManager. * Part of the Engine Hoist refactor - separates canvas from scene
logic. */
<div class="scene-canvas-container svelte-vqxrxo" style="${"width: " + escape(typeof width === "number" ? `${width}px` : width, true) + "; height: " + escape(typeof height === "number" ? `${height}px` : height, true) + ";"}"${add_attribute("this", container, 0)}><canvas class="scene-canvas svelte-vqxrxo" tabindex="0"${add_attribute("this", canvas, 0)}></canvas> </div>`;
});
class SceneManager {
  engine = null;
  canvas = null;
  scenes = /* @__PURE__ */ new Map();
  sceneFactories = /* @__PURE__ */ new Map();
  currentLocation = "LOADING";
  isTransitioning = false;
  renderLoopId = null;
  // Callbacks for state changes
  onLocationChange = null;
  /**
   * Initialize the engine with a canvas element.
   */
  initialize(canvas) {
    if (this.engine) {
      console.warn("[SceneManager] Already initialized");
      return;
    }
    this.canvas = canvas;
    this.engine = new Engine(canvas, true, {
      stencil: true,
      preserveDrawingBuffer: true,
      antialias: true,
      powerPreference: "high-performance"
    });
    window.addEventListener("resize", this.handleResize);
    console.log("[SceneManager] Engine initialized");
  }
  /**
   * Register a scene factory for a location.
   */
  registerSceneFactory(location, factory) {
    this.sceneFactories.set(location, factory);
  }
  /**
   * Create a new scene for a location.
   */
  async createScene(location) {
    if (!this.engine) {
      throw new Error("[SceneManager] Engine not initialized");
    }
    const existingScene = this.scenes.get(location);
    if (existingScene) {
      existingScene.dispose();
    }
    const scene = new Scene(this.engine);
    scene.clearColor = new Color4(0, 0, 0, 1);
    const factory = this.sceneFactories.get(location);
    if (factory) {
      await factory.create(scene);
    }
    this.scenes.set(location, scene);
    return scene;
  }
  /**
   * Transition to a new location with optional fade effect.
   */
  async transitionTo(location, options = {}) {
    if (this.isTransitioning) {
      console.warn("[SceneManager] Transition already in progress");
      return;
    }
    if (!this.engine) {
      throw new Error("[SceneManager] Engine not initialized");
    }
    const { fadeDuration = 500 } = options;
    this.isTransitioning = true;
    console.log(`[SceneManager] Transitioning to ${location}`);
    try {
      let targetScene = this.scenes.get(location);
      if (!targetScene) {
        targetScene = await this.createScene(location);
      }
      this.currentLocation = location;
      if (this.renderLoopId !== null) {
        this.engine.stopRenderLoop();
      }
      this.engine.runRenderLoop(() => {
        const activeScene = this.scenes.get(this.currentLocation);
        activeScene?.render();
      });
      this.onLocationChange?.(location);
    } finally {
      this.isTransitioning = false;
    }
  }
  /**
   * Get the current active scene.
   */
  getActiveScene() {
    return this.scenes.get(this.currentLocation) ?? null;
  }
  /**
   * Get scene for a specific location.
   */
  getScene(location) {
    return this.scenes.get(location) ?? null;
  }
  /**
   * Get the Babylon.js engine.
   */
  getEngine() {
    return this.engine;
  }
  /**
   * Get current location.
   */
  getCurrentLocation() {
    return this.currentLocation;
  }
  /**
   * Check if currently transitioning.
   */
  isInTransition() {
    return this.isTransitioning;
  }
  /**
   * Set callback for location changes.
   */
  setOnLocationChange(callback) {
    this.onLocationChange = callback;
  }
  /**
   * Handle window resize.
   */
  handleResize = () => {
    this.engine?.resize();
  };
  /**
   * Dispose of all resources.
   */
  dispose() {
    console.log("[SceneManager] Disposing...");
    window.removeEventListener("resize", this.handleResize);
    for (const factory of this.sceneFactories.values()) {
      factory.dispose();
    }
    this.sceneFactories.clear();
    for (const scene of this.scenes.values()) {
      scene.dispose();
    }
    this.scenes.clear();
    this.engine?.dispose();
    this.engine = null;
    this.canvas = null;
  }
}
const sceneManager = new SceneManager();
const WorldMapModal = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let displayStats;
  let $$unsubscribe_mapStore;
  $$unsubscribe_mapStore = subscribe(mapStore, (value) => value);
  let { isOpen = false } = $$props;
  let { onClose } = $$props;
  let containerWidth = 0;
  let containerHeight = 0;
  let worldMapData = null;
  let hasRequestedMap = false;
  let cameraZoom = 1;
  let cameraX = 0.5;
  let cameraY = 0.5;
  let activeLayers = /* @__PURE__ */ new Set();
  let showMineralsOverlay = false;
  let simStats = {
    year: 0,
    events: []
  };
  function cleanupRenderers() {
    worldMapData = null;
  }
  function requestWorldMap(highRes = false) {
    if (!highRes) ;
    else {
      console.log("[WorldMap] Proceeding to load 4K background map...");
    }
    const payload = highRes ? { width: 4096, height: 2048 } : {};
    gameWebSocket.sendRawCommand("world_map_image", payload);
    if (!highRes) {
      setTimeout(
        () => {
        },
        3e3
      );
    }
  }
  onDestroy(() => {
  });
  if ($$props.isOpen === void 0 && $$bindings.isOpen && isOpen !== void 0) $$bindings.isOpen(isOpen);
  if ($$props.onClose === void 0 && $$bindings.onClose && onClose !== void 0) $$bindings.onClose(onClose);
  displayStats = {
    age: worldMapData?.simulated_years ? (worldMapData.simulated_years / 1e6).toFixed(1) + "M Years" : "--",
    temp: worldMapData?.avg_temperature !== void 0 ? worldMapData.avg_temperature.toFixed(1) + "°C" : "--",
    elev: worldMapData?.max_elevation !== void 0 ? (worldMapData.max_elevation / 1e3).toFixed(1) + "km" : "--",
    sea: worldMapData?.sea_level !== void 0 ? worldMapData.sea_level.toFixed(0) + "m" : "--",
    land: worldMapData?.land_coverage !== void 0 ? worldMapData.land_coverage.toFixed(1) + "%" : "--"
  };
  {
    if (isOpen && !hasRequestedMap) {
      hasRequestedMap = true;
      console.log("[WorldMapModal] Modal opened, requesting world map...");
      {
        requestWorldMap();
      }
      sceneManager.registerSceneFactory("PREVIEW", {
        create: async (scene) => {
          console.log("[WorldMapModal] Created PREVIEW scene");
        },
        dispose: () => {
          console.log("[WorldMapModal] Disposing PREVIEW scene");
        }
      });
    }
  }
  {
    if (!isOpen) {
      hasRequestedMap = false;
      cleanupRenderers();
    }
  }
  $$unsubscribe_mapStore();
  return `${isOpen ? `<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm"><div class="bg-gray-900 border border-gray-700 rounded-lg shadow-2xl w-[90vw] h-[90vh] flex flex-col overflow-hidden"> <div class="flex justify-between items-center p-4 border-b border-gray-800 bg-gray-800/50"><h2 class="text-xl font-bold text-blue-400" data-svelte-h="svelte-1kpv5fi">World Map &amp; Simulation</h2> <div class="flex gap-4 items-center"> <button class="${"px-3 py-1 rounded text-sm font-medium transition-colors " + escape(
    "bg-blue-600 text-white",
    true
  )}">${escape("🌍 Globe")}</button> <div class="text-sm text-gray-400">Year: <span class="text-white font-mono">${escape(simStats.year)}</span></div> <button class="text-gray-400 hover:text-white transition-colors" data-svelte-h="svelte-lulo1"><svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg></button></div></div>  <div class="flex-1 flex overflow-hidden"><div class="flex-1 bg-black relative">${` ${validate_component(SceneCanvas, "SceneCanvas").$$render($$result, {}, {}, {})} ${``}`}  ${worldMapData?.overlays && (activeLayers.size > 0 || showMineralsOverlay) ? `${validate_component(MapOverlayCanvas, "MapOverlayCanvas").$$render(
    $$result,
    {
      width: containerWidth,
      height: containerHeight,
      gridWidth: worldMapData.grid_width,
      gridHeight: worldMapData.grid_height,
      overlayData: worldMapData.overlays,
      activeLayers,
      showMinerals: showMineralsOverlay,
      cameraX,
      cameraY,
      zoom: cameraZoom
    },
    {},
    {}
  )}` : ``}  ${``}  <div class="absolute top-4 left-4 p-4 rounded-lg bg-gray-900/80 backdrop-blur border border-gray-700 shadow-xl min-w-[200px]"><h3 class="text-xs font-bold text-gray-500 uppercase tracking-wider mb-3 border-b border-gray-700 pb-2" data-svelte-h="svelte-1x6nwi0">Planetary Data</h3> <div class="space-y-3 font-mono text-sm"><div class="flex justify-between items-center"><span class="text-gray-400" data-svelte-h="svelte-12j8070">Age</span> <span class="text-purple-400 font-bold">${escape(displayStats.age)}</span></div> <div class="flex justify-between items-center"><span class="text-gray-400" data-svelte-h="svelte-lhno1r">Avg Temp</span> <span class="text-red-400 font-bold">${escape(displayStats.temp)}</span></div> <div class="flex justify-between items-center"><span class="text-gray-400" data-svelte-h="svelte-o3x6j5">Max Elev</span> <span class="text-yellow-400">${escape(displayStats.elev)}</span></div> <div class="flex justify-between items-center"><span class="text-gray-400" data-svelte-h="svelte-1ygxwci">Sea Level</span> <span class="text-blue-400">${escape(displayStats.sea)}</span></div> <div class="flex justify-between items-center"><span class="text-gray-400" data-svelte-h="svelte-6mrvdw">Land Mass</span> <span class="text-green-400">${escape(displayStats.land)}</span></div> ${worldMapData?.seed ? `<div class="pt-2 mt-2 border-t border-gray-700 text-xs flex justify-between"><span class="text-gray-500" data-svelte-h="svelte-1w94vz7">Seed</span> <span class="text-gray-400">${escape(worldMapData.seed)}</span></div>` : ``}</div></div>  <div class="absolute bottom-4 left-4">${validate_component(WorldMapLegend, "WorldMapLegend").$$render(
    $$result,
    {
      mode: (activeLayers.size === 0 ? worldMapData?.is_simulated ? "terrain" : "biome" : Array.from(activeLayers).pop()) || "terrain",
      activeLayers
    },
    {},
    {}
  )}</div>  <div class="absolute bottom-4 right-4 flex flex-col gap-2 items-end"> ${worldMapData?.overlays ? `<div class="bg-gray-800/90 p-3 rounded-lg border border-gray-700 space-y-3 min-w-[200px]"><div class="text-xs text-gray-400 font-bold uppercase border-b border-gray-700 pb-1" data-svelte-h="svelte-ejec1g">Data Layers</div>  <div class="space-y-1">${each(
    [
      {
        id: "none",
        label: "Clear All",
        icon: "🚫"
      },
      {
        id: "tectonics",
        label: "Tectonics",
        icon: "📐"
      },
      {
        id: "elevation",
        label: "Elevation",
        icon: "🏔️"
      },
      {
        id: "temp",
        label: "Temperature",
        icon: "🌡️"
      },
      {
        id: "moisture",
        label: "Moisture",
        icon: "💧"
      },
      { id: "biome", label: "Biomes", icon: "🌿" },
      {
        id: "features",
        label: "Terrain Features",
        icon: "📍"
      }
    ],
    (layer) => {
      return `<button class="${"w-full text-left px-2 py-1.5 rounded text-xs flex items-center justify-between transition-colors " + escape(
        activeLayers.has(layer.id) || layer.id === "none" && activeLayers.size === 0 ? "bg-blue-600/30 text-blue-200 border border-blue-500/30" : "hover:bg-gray-700 text-gray-300",
        true
      )}"><span class="flex items-center gap-2"><span>${escape(layer.icon)}</span> ${escape(layer.label)}</span> ${activeLayers.has(layer.id) ? `<span class="w-1.5 h-1.5 rounded-full bg-blue-400"></span>` : ``} </button>`;
    }
  )}</div>  ${worldMapData.overlays.resources || worldMapData.overlays.minerals ? `<div class="pt-2 border-t border-gray-700"><label class="flex items-center gap-2 cursor-pointer hover:bg-gray-700/50 px-2 py-1 rounded transition-colors group"><input type="checkbox" class="w-4 h-4 accent-yellow-500 rounded border-gray-600 bg-gray-700"${add_attribute("checked", showMineralsOverlay, 1)}> <div class="flex flex-col"><span class="text-xs text-gray-300 group-hover:text-white" data-svelte-h="svelte-1d5yxe8">Show Resources</span> ${worldMapData.overlays.resources ? `<span class="text-[10px] text-gray-500">${escape(worldMapData.overlays.resources.length)} nodes</span>` : ``}</div></label></div>` : ``}</div>` : ``} <div class="bg-gray-800/80 p-2 rounded text-xs text-gray-300" data-svelte-h="svelte-2h080q">WASD to Pan • Scroll to Zoom</div></div></div>  <div class="w-72 bg-gray-850 border-l border-gray-800 flex flex-col"> ${worldMapData?.satellites?.length > 0 ? `<div class="p-4 border-b border-gray-800"><h3 class="font-bold text-gray-300 mb-3 flex items-center gap-2" data-svelte-h="svelte-qfipe7"><span class="text-lg">🌙</span>
                                Natural Satellites</h3> <div class="space-y-2 text-sm">${each(worldMapData.satellites, (sat) => {
    return `<div class="flex justify-between items-center"><span class="text-gray-300">${escape(sat.name)}</span> <span class="text-gray-500 text-xs">${escape(sat.mass.toFixed(2))}x Luna</span> </div>`;
  })}</div>  <div class="mt-3 pt-3 border-t border-gray-700 text-xs"><div class="flex justify-between"><span class="text-gray-500" data-svelte-h="svelte-89p0h8">Climate Stability</span> <span${add_attribute(
    "class",
    worldMapData.satellites.reduce((a, s) => a + s.mass, 0) > 0.01 ? "text-green-400" : "text-yellow-400",
    0
  )}>${escape(worldMapData.satellites.reduce((a, s) => a + s.mass, 0) > 0.01 ? "Stable" : "Variable")}</span></div> <div class="flex justify-between mt-1"><span class="text-gray-500" data-svelte-h="svelte-hpjymz">Impact Shield</span> <span class="text-blue-400">${escape(Math.min(worldMapData.satellites.length * 5, 20))}%</span></div></div></div>` : `<div class="p-4 border-b border-gray-800" data-svelte-h="svelte-ehi5oe"><h3 class="font-bold text-gray-300 mb-2 flex items-center gap-2"><span class="text-lg">🌙</span>
                                Natural Satellites</h3> <div class="text-gray-500 text-sm italic">No moons detected</div> <div class="mt-2 text-xs flex justify-between"><span class="text-gray-500">Climate Stability</span> <span class="text-red-400">Chaotic</span></div></div>`}  <div class="flex-1 flex flex-col overflow-hidden"><h3 class="font-bold text-gray-300 p-4 pb-2" data-svelte-h="svelte-10tq5ud">Event Log</h3> <div class="flex-1 overflow-y-auto p-4 pt-0 space-y-2 font-mono text-xs">${simStats.events.length ? each(simStats.events, (event) => {
    return `<div class="text-gray-400 border-l-2 border-gray-700 pl-2 py-1">${escape(event)} </div>`;
  }) : `<div class="text-gray-600 italic text-center mt-10" data-svelte-h="svelte-1wo3kiy">No recent events
                                </div>`}</div></div></div></div></div></div>` : ``}`;
});
const hapticEnabled = writable(true);
if (typeof localStorage !== "undefined") {
  const stored = localStorage.getItem("hapticEnabled");
  if (stored !== null) {
    hapticEnabled.set(stored === "true");
  }
}
hapticEnabled.subscribe((value) => {
  if (typeof localStorage !== "undefined") {
    localStorage.setItem("hapticEnabled", String(value));
  }
});
hapticEnabled.subscribe((value) => {
});
const css$4 = {
  code: ".mode-toggle.svelte-ndobyo{z-index:50}.mode-toggle.svelte-ndobyo:focus{outline:none;box-shadow:0 0 0 2px rgba(59, 130, 246, 0.5)}.mode-toggle.svelte-ndobyo:active{transform:scale(0.98)}",
  map: `{"version":3,"file":"ModeToggle.svelte","sources":["ModeToggle.svelte"],"sourcesContent":["<script lang=\\"ts\\">import { interfaceMode, toggleInterfaceMode } from \\"$lib/stores/ui\\";\\nexport let compact = false;\\n<\/script>\\n\\n<button\\n    data-testid=\\"mode-toggle\\"\\n    on:click={toggleInterfaceMode}\\n    class=\\"mode-toggle flex items-center gap-2 px-3 py-2 rounded-lg transition-all duration-200\\n         bg-gray-800/80 hover:bg-gray-700/80 border border-gray-600/50 hover:border-gray-500/50\\n         text-gray-300 hover:text-white shadow-md backdrop-blur-sm\\n         {compact ? 'text-sm' : 'text-base'}\\"\\n    title={$interfaceMode === \\"TEXT\\"\\n        ? \\"Switch to 3D View\\"\\n        : \\"Switch to Text Mode\\"}\\n    aria-label={$interfaceMode === \\"TEXT\\"\\n        ? \\"Switch to 3D simulation view\\"\\n        : \\"Switch to text-based MUD view\\"}\\n>\\n    {#if $interfaceMode === \\"TEXT\\"}\\n        <!-- Currently in TEXT mode, show icon to switch to VISUAL -->\\n        <svg\\n            class=\\"w-5 h-5\\"\\n            fill=\\"none\\"\\n            stroke=\\"currentColor\\"\\n            viewBox=\\"0 0 24 24\\"\\n            aria-hidden=\\"true\\"\\n        >\\n            <!-- Globe/3D icon -->\\n            <path\\n                stroke-linecap=\\"round\\"\\n                stroke-linejoin=\\"round\\"\\n                stroke-width=\\"2\\"\\n                d=\\"M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9\\"\\n            />\\n        </svg>\\n        {#if !compact}\\n            <span>3D View</span>\\n        {/if}\\n    {:else}\\n        <!-- Currently in VISUAL mode, show icon to switch to TEXT -->\\n        <svg\\n            class=\\"w-5 h-5\\"\\n            fill=\\"none\\"\\n            stroke=\\"currentColor\\"\\n            viewBox=\\"0 0 24 24\\"\\n            aria-hidden=\\"true\\"\\n        >\\n            <!-- Document/Text icon -->\\n            <path\\n                stroke-linecap=\\"round\\"\\n                stroke-linejoin=\\"round\\"\\n                stroke-width=\\"2\\"\\n                d=\\"M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z\\"\\n            />\\n        </svg>\\n        {#if !compact}\\n            <span>Text Mode</span>\\n        {/if}\\n    {/if}\\n</button>\\n\\n<style>\\n    .mode-toggle {\\n        /* Ensure clickable on overlay layers */\\n        z-index: 50;\\n    }\\n\\n    .mode-toggle:focus {\\n        outline: none;\\n        box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.5);\\n    }\\n\\n    .mode-toggle:active {\\n        transform: scale(0.98);\\n    }\\n</style>\\n"],"names":[],"mappings":"AA8DI,0BAAa,CAET,OAAO,CAAE,EACb,CAEA,0BAAY,MAAO,CACf,OAAO,CAAE,IAAI,CACb,UAAU,CAAE,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,GAAG,CAAC,KAAK,EAAE,CAAC,CAAC,GAAG,CAAC,CAAC,GAAG,CAAC,CAAC,GAAG,CAChD,CAEA,0BAAY,OAAQ,CAChB,SAAS,CAAE,MAAM,IAAI,CACzB"}`
};
const ModeToggle = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $interfaceMode, $$unsubscribe_interfaceMode;
  $$unsubscribe_interfaceMode = subscribe(interfaceMode, (value) => $interfaceMode = value);
  let { compact = false } = $$props;
  if ($$props.compact === void 0 && $$bindings.compact && compact !== void 0) $$bindings.compact(compact);
  $$result.css.add(css$4);
  $$unsubscribe_interfaceMode();
  return `<button data-testid="mode-toggle" class="${"mode-toggle flex items-center gap-2 px-3 py-2 rounded-lg transition-all duration-200 bg-gray-800/80 hover:bg-gray-700/80 border border-gray-600/50 hover:border-gray-500/50 text-gray-300 hover:text-white shadow-md backdrop-blur-sm " + escape(compact ? "text-sm" : "text-base", true) + " svelte-ndobyo"}"${add_attribute(
    "title",
    $interfaceMode === "TEXT" ? "Switch to 3D View" : "Switch to Text Mode",
    0
  )}${add_attribute(
    "aria-label",
    $interfaceMode === "TEXT" ? "Switch to 3D simulation view" : "Switch to text-based MUD view",
    0
  )}>${$interfaceMode === "TEXT" ? ` <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"></path></svg> ${!compact ? `<span data-svelte-h="svelte-o090ni">3D View</span>` : ``}` : ` <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg> ${!compact ? `<span data-svelte-h="svelte-axquik">Text Mode</span>` : ``}`} </button>`;
});
const css$3 = {
  code: '.mud-layout.svelte-mcvj1n{font-family:"Georgia", "Times New Roman", Times, serif}.mud-layout input,.mud-layout code,.mud-layout pre{font-family:"SF Mono", "Monaco", "Inconsolata", "Fira Code",\n            "Fira Mono", "Source Code Pro", monospace}',
  map: '{"version":3,"file":"MUDModeLayout.svelte","sources":["MUDModeLayout.svelte"],"sourcesContent":["<script lang=\\"ts\\">import { isMobile } from \\"$lib/stores/ui\\";\\n<\/script>\\n\\n<div\\n    class=\\"mud-layout flex flex-col h-screen bg-gray-950 text-gray-100 overflow-hidden\\"\\n>\\n    <!-- Status Bar (Fixed Top) -->\\n    <header\\n        class=\\"h-14 bg-gray-900 border-b border-gray-800 flex items-center px-4 shrink-0 z-20\\"\\n    >\\n        <slot name=\\"status-bar\\">\\n            <div class=\\"text-gray-500 text-sm\\">Status Bar</div>\\n        </slot>\\n\\n        <!-- Mode Toggle Button (right side) -->\\n        <div class=\\"ml-auto\\">\\n            <slot name=\\"mode-toggle\\" />\\n        </div>\\n    </header>\\n\\n    <!-- Main Content Area -->\\n    <div class=\\"flex flex-1 min-h-0 overflow-hidden\\">\\n        <!-- Left Panel: Stats/Inventory (Desktop only) -->\\n        {#if !$isMobile}\\n            <aside\\n                class=\\"w-64 bg-gray-900 border-r border-gray-800 flex flex-col shrink-0 overflow-hidden\\"\\n            >\\n                <div\\n                    class=\\"p-3 border-b border-gray-800 text-xs font-semibold text-gray-500 uppercase tracking-wider\\"\\n                >\\n                    Character\\n                </div>\\n                <div class=\\"flex-1 overflow-y-auto p-3\\">\\n                    <slot name=\\"left-panel\\">\\n                        <div class=\\"text-gray-600 text-sm\\">Stats</div>\\n                    </slot>\\n                </div>\\n            </aside>\\n        {/if}\\n\\n        <!-- Center: Text Display (Main Game Output) -->\\n        <main class=\\"flex-1 flex flex-col min-w-0 overflow-hidden bg-gray-950\\">\\n            <div class=\\"flex-1 overflow-y-auto\\">\\n                <slot name=\\"main-display\\">\\n                    <div class=\\"p-6 text-gray-600 italic text-center\\">\\n                        Awaiting connection...\\n                    </div>\\n                </slot>\\n            </div>\\n        </main>\\n\\n        <!-- Right Panel: Map/Navigation (Desktop only) -->\\n        {#if !$isMobile}\\n            <aside\\n                class=\\"w-56 bg-gray-900 border-l border-gray-800 flex flex-col shrink-0 overflow-hidden\\"\\n            >\\n                <div\\n                    class=\\"p-3 border-b border-gray-800 text-xs font-semibold text-gray-500 uppercase tracking-wider\\"\\n                >\\n                    Navigation\\n                </div>\\n                <div class=\\"flex-1 overflow-y-auto p-3\\">\\n                    <slot name=\\"right-panel\\">\\n                        <div class=\\"text-gray-600 text-sm\\">Map</div>\\n                    </slot>\\n                </div>\\n            </aside>\\n        {/if}\\n    </div>\\n\\n    <!-- Command Input (Fixed Bottom) -->\\n    <footer class=\\"bg-gray-900 border-t border-gray-800 p-3 shrink-0 z-20\\">\\n        <slot name=\\"command-input\\">\\n            <div class=\\"text-gray-500 text-sm\\">Input</div>\\n        </slot>\\n    </footer>\\n\\n    <!-- Mobile Controls (e.g., D-pad) - Mobile only -->\\n    {#if $isMobile}\\n        <div\\n            class=\\"h-16 bg-gray-900 border-t border-gray-800 p-2 shrink-0 flex justify-center items-center\\"\\n        >\\n            <slot name=\\"controls\\">\\n                <div class=\\"text-gray-600 text-xs\\">Controls</div>\\n            </slot>\\n        </div>\\n    {/if}\\n</div>\\n\\n<style>\\n    .mud-layout {\\n        /* Rich typography for MUD feel */\\n        font-family: \\"Georgia\\", \\"Times New Roman\\", Times, serif;\\n    }\\n\\n    /* Use mono for command input areas */\\n    :global(.mud-layout input),\\n    :global(.mud-layout code),\\n    :global(.mud-layout pre) {\\n        font-family: \\"SF Mono\\", \\"Monaco\\", \\"Inconsolata\\", \\"Fira Code\\",\\n            \\"Fira Mono\\", \\"Source Code Pro\\", monospace;\\n    }\\n</style>\\n"],"names":[],"mappings":"AA0FI,yBAAY,CAER,WAAW,CAAE,SAAS,CAAC,CAAC,iBAAiB,CAAC,CAAC,KAAK,CAAC,CAAC,KACtD,CAGQ,iBAAkB,CAClB,gBAAiB,CACjB,eAAiB,CACrB,WAAW,CAAE,SAAS,CAAC,CAAC,QAAQ,CAAC,CAAC,aAAa,CAAC,CAAC,WAAW;AACpE,YAAY,WAAW,CAAC,CAAC,iBAAiB,CAAC,CAAC,SACxC"}'
};
const MUDModeLayout = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $isMobile, $$unsubscribe_isMobile;
  $$unsubscribe_isMobile = subscribe(isMobile, (value) => $isMobile = value);
  $$result.css.add(css$3);
  $$unsubscribe_isMobile();
  return `<div class="mud-layout flex flex-col h-screen bg-gray-950 text-gray-100 overflow-hidden svelte-mcvj1n"> <header class="h-14 bg-gray-900 border-b border-gray-800 flex items-center px-4 shrink-0 z-20">${slots["status-bar"] ? slots["status-bar"]({}) : ` <div class="text-gray-500 text-sm" data-svelte-h="svelte-1mk1yoz">Status Bar</div> `}  <div class="ml-auto">${slots["mode-toggle"] ? slots["mode-toggle"]({}) : ``}</div></header>  <div class="flex flex-1 min-h-0 overflow-hidden"> ${!$isMobile ? `<aside class="w-64 bg-gray-900 border-r border-gray-800 flex flex-col shrink-0 overflow-hidden"><div class="p-3 border-b border-gray-800 text-xs font-semibold text-gray-500 uppercase tracking-wider" data-svelte-h="svelte-19vf3ck">Character</div> <div class="flex-1 overflow-y-auto p-3">${slots["left-panel"] ? slots["left-panel"]({}) : ` <div class="text-gray-600 text-sm" data-svelte-h="svelte-39q014">Stats</div> `}</div></aside>` : ``}  <main class="flex-1 flex flex-col min-w-0 overflow-hidden bg-gray-950"><div class="flex-1 overflow-y-auto">${slots["main-display"] ? slots["main-display"]({}) : ` <div class="p-6 text-gray-600 italic text-center" data-svelte-h="svelte-163xuen">Awaiting connection...</div> `}</div></main>  ${!$isMobile ? `<aside class="w-56 bg-gray-900 border-l border-gray-800 flex flex-col shrink-0 overflow-hidden"><div class="p-3 border-b border-gray-800 text-xs font-semibold text-gray-500 uppercase tracking-wider" data-svelte-h="svelte-hzoznz">Navigation</div> <div class="flex-1 overflow-y-auto p-3">${slots["right-panel"] ? slots["right-panel"]({}) : ` <div class="text-gray-600 text-sm" data-svelte-h="svelte-n705yp">Map</div> `}</div></aside>` : ``}</div>  <footer class="bg-gray-900 border-t border-gray-800 p-3 shrink-0 z-20">${slots["command-input"] ? slots["command-input"]({}) : ` <div class="text-gray-500 text-sm" data-svelte-h="svelte-1lwgc9w">Input</div> `}</footer>  ${$isMobile ? `<div class="h-16 bg-gray-900 border-t border-gray-800 p-2 shrink-0 flex justify-center items-center">${slots.controls ? slots.controls({}) : ` <div class="text-gray-600 text-xs" data-svelte-h="svelte-2siou8">Controls</div> `}</div>` : ``} </div>`;
});
const css$2 = {
  code: ".message-overlay.svelte-19lmdae{position:fixed;bottom:80px;left:20px;right:20px;pointer-events:none;z-index:100;display:flex;flex-direction:column;gap:4px;max-height:200px;overflow:hidden}.message.svelte-19lmdae{background:rgba(0, 0, 0, 0.6);backdrop-filter:blur(4px);padding:8px 16px;border-radius:4px;font-family:monospace;font-size:0.9rem;line-height:1.4;transition:opacity 0.3s ease,\n            transform 0.3s ease;text-shadow:0 1px 2px rgba(0, 0, 0, 0.5)}",
  map: '{"version":3,"file":"MessageOverlay.svelte","sources":["MessageOverlay.svelte"],"sourcesContent":["<script lang=\\"ts\\">import { onDestroy } from \\"svelte\\";\\nimport { gameOutput } from \\"$lib/stores/ui\\";\\nconst MAX_VISIBLE = 4;\\nlet displayMessages = [];\\nlet updateInterval = null;\\nconst unsubscribe = gameOutput.subscribe((messages) => {\\n  const recent = messages.slice(-MAX_VISIBLE);\\n  displayMessages = recent.map((msg, i) => ({\\n    ...msg,\\n    displayId: msg.id,\\n    opacity: 0.3 + (i + 1) / MAX_VISIBLE * 0.6,\\n    // 0.3 to 0.9\\n    age: 0\\n  }));\\n});\\nfunction startAging() {\\n  if (updateInterval) return;\\n  updateInterval = setInterval(() => {\\n    displayMessages = displayMessages.map((msg) => ({\\n      ...msg,\\n      age: msg.age + 0.1,\\n      opacity: Math.max(0, msg.opacity - 0.02 * msg.age)\\n    })).filter((msg) => msg.opacity > 0);\\n  }, 100);\\n}\\nstartAging();\\nonDestroy(() => {\\n  unsubscribe();\\n  if (updateInterval) {\\n    clearInterval(updateInterval);\\n  }\\n});\\nfunction getTypeClass(type) {\\n  switch (type) {\\n    case \\"error\\":\\n      return \\"text-red-400\\";\\n    case \\"system\\":\\n      return \\"text-yellow-400\\";\\n    case \\"player\\":\\n      return \\"text-blue-300\\";\\n    case \\"emote\\":\\n      return \\"text-orange-300 italic\\";\\n    default:\\n      return \\"text-gray-200\\";\\n  }\\n}\\n<\/script>\\n\\n/** * MessageOverlay.svelte * Semi-transparent message display at screen bottom.\\n* Messages fade in and out without blocking FPS camera controls. */\\n<div class=\\"message-overlay\\">\\n    {#each displayMessages as msg (msg.displayId)}\\n        <div\\n            class=\\"message {getTypeClass(msg.type)}\\"\\n            style=\\"opacity: {msg.opacity}; transform: translateY({-msg.age *\\n                5}px);\\"\\n        >\\n            {msg.text}\\n        </div>\\n    {/each}\\n</div>\\n\\n<style>\\n    .message-overlay {\\n        position: fixed;\\n        bottom: 80px; /* Above command input */\\n        left: 20px;\\n        right: 20px;\\n        pointer-events: none; /* Allow click-through for FPS camera */\\n        z-index: 100;\\n        display: flex;\\n        flex-direction: column;\\n        gap: 4px;\\n        max-height: 200px;\\n        overflow: hidden;\\n    }\\n\\n    .message {\\n        background: rgba(0, 0, 0, 0.6);\\n        backdrop-filter: blur(4px);\\n        padding: 8px 16px;\\n        border-radius: 4px;\\n        font-family: monospace;\\n        font-size: 0.9rem;\\n        line-height: 1.4;\\n        transition:\\n            opacity 0.3s ease,\\n            transform 0.3s ease;\\n        text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);\\n    }\\n</style>\\n"],"names":[],"mappings":"AA+DI,+BAAiB,CACb,QAAQ,CAAE,KAAK,CACf,MAAM,CAAE,IAAI,CACZ,IAAI,CAAE,IAAI,CACV,KAAK,CAAE,IAAI,CACX,cAAc,CAAE,IAAI,CACpB,OAAO,CAAE,GAAG,CACZ,OAAO,CAAE,IAAI,CACb,cAAc,CAAE,MAAM,CACtB,GAAG,CAAE,GAAG,CACR,UAAU,CAAE,KAAK,CACjB,QAAQ,CAAE,MACd,CAEA,uBAAS,CACL,UAAU,CAAE,KAAK,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,GAAG,CAAC,CAC9B,eAAe,CAAE,KAAK,GAAG,CAAC,CAC1B,OAAO,CAAE,GAAG,CAAC,IAAI,CACjB,aAAa,CAAE,GAAG,CAClB,WAAW,CAAE,SAAS,CACtB,SAAS,CAAE,MAAM,CACjB,WAAW,CAAE,GAAG,CAChB,UAAU,CACN,OAAO,CAAC,IAAI,CAAC,IAAI;AAC7B,YAAY,SAAS,CAAC,IAAI,CAAC,IAAI,CACvB,WAAW,CAAE,CAAC,CAAC,GAAG,CAAC,GAAG,CAAC,KAAK,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,GAAG,CAC5C"}'
};
const MAX_VISIBLE = 4;
function getTypeClass(type) {
  switch (type) {
    case "error":
      return "text-red-400";
    case "system":
      return "text-yellow-400";
    case "player":
      return "text-blue-300";
    case "emote":
      return "text-orange-300 italic";
    default:
      return "text-gray-200";
  }
}
const MessageOverlay = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let displayMessages = [];
  let updateInterval = null;
  const unsubscribe = gameOutput.subscribe((messages) => {
    const recent = messages.slice(-MAX_VISIBLE);
    displayMessages = recent.map((msg, i) => ({
      ...msg,
      displayId: msg.id,
      opacity: 0.3 + (i + 1) / MAX_VISIBLE * 0.6,
      // 0.3 to 0.9
      age: 0
    }));
  });
  function startAging() {
    if (updateInterval) return;
    updateInterval = setInterval(
      () => {
        displayMessages = displayMessages.map((msg) => ({
          ...msg,
          age: msg.age + 0.1,
          opacity: Math.max(0, msg.opacity - 0.02 * msg.age)
        })).filter((msg) => msg.opacity > 0);
      },
      100
    );
  }
  startAging();
  onDestroy(() => {
    unsubscribe();
    if (updateInterval) {
      clearInterval(updateInterval);
    }
  });
  $$result.css.add(css$2);
  return `/** * MessageOverlay.svelte * Semi-transparent message display at screen bottom.
* Messages fade in and out without blocking FPS camera controls. */
<div class="message-overlay svelte-19lmdae">${each(displayMessages, (msg) => {
    return `<div class="${"message " + escape(getTypeClass(msg.type), true) + " svelte-19lmdae"}" style="${"opacity: " + escape(msg.opacity, true) + "; transform: translateY(" + escape(-msg.age * 5, true) + "px);"}">${escape(msg.text)} </div>`;
  })} </div>`;
});
class FirstPersonController {
  scene;
  camera;
  moveSpeed;
  lookSpeed;
  jumpHeight;
  gravity;
  velocity = Vector3.Zero();
  isGrounded = true;
  disposed = false;
  collisionTarget;
  eyeHeight;
  constructor(scene, startPosition, options = {}) {
    this.scene = scene;
    this.moveSpeed = options.moveSpeed ?? 5;
    this.lookSpeed = options.lookSpeed ?? 5e-3;
    this.jumpHeight = options.jumpHeight ?? 2;
    this.gravity = options.gravity ?? 9.8;
    this.collisionTarget = options.collisionTarget ?? null;
    this.eyeHeight = options.eyeHeight ?? 1.7;
    this.camera = new UniversalCamera("fpsCamera", startPosition, scene);
    this.camera.setTarget(startPosition.add(new Vector3(0, 0, 1)));
    this.camera.ellipsoid = new Vector3(0.5, 0.9, 0.5);
    this.camera.checkCollisions = true;
    this.camera.applyGravity = true;
    this.camera.keysUp = [87];
    this.camera.keysDown = [83];
    this.camera.keysLeft = [65];
    this.camera.keysRight = [68];
    this.camera.keysRotateLeft = [81];
    this.camera.keysRotateRight = [69];
    this.camera.angularSensibility = 500;
    if (this.collisionTarget) {
      this.collisionTarget.checkCollisions = true;
    }
  }
  /**
   * Set a new collision target (useful when transitioning between scenes).
   */
  setCollisionTarget(target) {
    this.collisionTarget = target;
    if (target) {
      target.checkCollisions = true;
    }
  }
  /**
   * Get the current collision target mesh.
   */
  getCollisionTarget() {
    return this.collisionTarget;
  }
  /**
   * Handle input for player movement.
   * Called each frame with delta time.
   */
  handleInput(deltaTime) {
    if (this.disposed) return;
    if (!this.isGrounded) {
      this.velocity.y -= this.gravity * deltaTime;
    }
  }
  /**
   * Get current world position as lat/lon/altitude.
   */
  getPosition() {
    const pos = this.camera.position;
    const distance = Math.sqrt(pos.x * pos.x + pos.z * pos.z);
    const lat = Math.atan2(pos.y, distance) * (180 / Math.PI);
    const lon = Math.atan2(pos.z, pos.x) * (180 / Math.PI);
    const altitude = pos.length() - 1;
    return { lat, lon, altitude };
  }
  /**
   * Get the camera instance.
   */
  getCamera() {
    return this.camera;
  }
  /**
   * Set camera as active.
   */
  activate() {
    if (this.disposed) return;
    this.scene.activeCamera = this.camera;
    const canvas = this.scene.getEngine().getRenderingCanvas();
    if (canvas) {
      this.camera.attachControl(canvas, true);
    }
  }
  /**
   * Deactivate camera controls.
   */
  deactivate() {
    if (this.disposed) return;
    this.camera.detachControl();
  }
  /**
   * Jump if grounded.
   */
  jump() {
    if (this.isGrounded && !this.disposed) {
      this.velocity.y = Math.sqrt(2 * this.gravity * this.jumpHeight);
      this.isGrounded = false;
    }
  }
  /**
   * Teleport to position.
   */
  teleport(position) {
    if (this.disposed) return;
    this.camera.position = position.clone();
    this.velocity = Vector3.Zero();
  }
  /**
   * Check if controller is disposed.
   */
  isDisposed() {
    return this.disposed;
  }
  /**
   * Dispose of resources.
   */
  dispose() {
    if (this.disposed) return;
    this.disposed = true;
    this.camera.detachControl();
    this.camera.dispose();
  }
}
class LobbyScene {
  scene = null;
  fpsController = null;
  floor = null;
  statue = null;
  portal = null;
  portalParticles = null;
  callbacks = {};
  portalEntered = false;
  // Room dimensions
  ROOM_WIDTH = 30;
  ROOM_DEPTH = 40;
  ROOM_HEIGHT = 12;
  PORTAL_RADIUS = 2;
  /**
   * Create the lobby scene.
   */
  // SceneFactory implementation
  async create(scene) {
    this.scene = scene;
    scene.clearColor = new Color4(0.02, 0.02, 0.03, 1);
    const ambient = new HemisphericLight("ambient", new Vector3(0, 1, 0), scene);
    ambient.intensity = 0.4;
    ambient.groundColor = new Color3(0.1, 0.08, 0.06);
    this.createFloor(scene);
    this.createWalls(scene);
    this.createStatue(scene);
    this.createPortal(scene);
    this.createFPSController(scene);
    scene.onBeforeRenderObservable.add(() => {
      const deltaTime = scene.getEngine().getDeltaTime() / 1e3;
      this.update(deltaTime);
    });
    console.log("[LobbyScene] Created");
  }
  /**
   * Set callbacks for interaction.
   */
  setCallbacks(callbacks) {
    this.callbacks = callbacks;
  }
  /**
   * Create marble floor.
   */
  createFloor(scene) {
    this.floor = MeshBuilder.CreateBox("floor", {
      width: this.ROOM_WIDTH,
      depth: this.ROOM_DEPTH,
      height: 0.5
    }, scene);
    this.floor.position.y = -0.25;
    this.floor.checkCollisions = true;
    const floorMat = new StandardMaterial("floorMat", scene);
    floorMat.diffuseColor = new Color3(0.9, 0.85, 0.8);
    floorMat.specularColor = new Color3(0.3, 0.3, 0.3);
    floorMat.specularPower = 32;
    this.floor.material = floorMat;
  }
  /**
   * Create walls using simple boxes (ceiling-less for now).
   */
  createWalls(scene) {
    const wallMat = new StandardMaterial("wallMat", scene);
    wallMat.diffuseColor = new Color3(0.85, 0.8, 0.75);
    wallMat.specularColor = new Color3(0.2, 0.2, 0.2);
    const wallThickness = 0.5;
    const wallHeight = this.ROOM_HEIGHT;
    const northWall = MeshBuilder.CreateBox("northWall", {
      width: this.ROOM_WIDTH,
      height: wallHeight,
      depth: wallThickness
    }, scene);
    northWall.position = new Vector3(0, wallHeight / 2, this.ROOM_DEPTH / 2);
    northWall.material = wallMat;
    northWall.checkCollisions = true;
    const southWall = MeshBuilder.CreateBox("southWall", {
      width: this.ROOM_WIDTH,
      height: wallHeight,
      depth: wallThickness
    }, scene);
    southWall.position = new Vector3(0, wallHeight / 2, -this.ROOM_DEPTH / 2);
    southWall.material = wallMat;
    southWall.checkCollisions = true;
    const eastWall = MeshBuilder.CreateBox("eastWall", {
      width: wallThickness,
      height: wallHeight,
      depth: this.ROOM_DEPTH
    }, scene);
    eastWall.position = new Vector3(this.ROOM_WIDTH / 2, wallHeight / 2, 0);
    eastWall.material = wallMat;
    eastWall.checkCollisions = true;
    const westWallLeft = MeshBuilder.CreateBox("westWallLeft", {
      width: wallThickness,
      height: wallHeight,
      depth: this.ROOM_DEPTH / 2 - this.PORTAL_RADIUS - 1
    }, scene);
    westWallLeft.position = new Vector3(
      -this.ROOM_WIDTH / 2,
      wallHeight / 2,
      this.ROOM_DEPTH / 4 + this.PORTAL_RADIUS / 2
    );
    westWallLeft.material = wallMat;
    westWallLeft.checkCollisions = true;
    const westWallRight = MeshBuilder.CreateBox("westWallRight", {
      width: wallThickness,
      height: wallHeight,
      depth: this.ROOM_DEPTH / 2 - this.PORTAL_RADIUS - 1
    }, scene);
    westWallRight.position = new Vector3(
      -this.ROOM_WIDTH / 2,
      wallHeight / 2,
      -this.ROOM_DEPTH / 4 - this.PORTAL_RADIUS / 2
    );
    westWallRight.material = wallMat;
    westWallRight.checkCollisions = true;
  }
  /**
   * Create the central statue (cylinder + capsule for now).
   */
  createStatue(scene) {
    const pedestal = MeshBuilder.CreateCylinder("pedestal", {
      height: 1.5,
      diameter: 2
    }, scene);
    pedestal.position = new Vector3(0, 0.75, 0);
    const pedestalMat = new StandardMaterial("pedestalMat", scene);
    pedestalMat.diffuseColor = new Color3(0.6, 0.6, 0.65);
    pedestal.material = pedestalMat;
    pedestal.checkCollisions = true;
    this.statue = MeshBuilder.CreateCapsule("statue", {
      height: 4,
      radius: 0.5
    }, scene);
    this.statue.position = new Vector3(0, 3.5, 0);
    const statueMat = new StandardMaterial("statueMat", scene);
    statueMat.diffuseColor = new Color3(0.9, 0.9, 0.95);
    statueMat.emissiveColor = new Color3(0.02, 0.02, 0.05);
    statueMat.specularColor = new Color3(0.4, 0.4, 0.4);
    statueMat.specularPower = 64;
    this.statue.material = statueMat;
    const statueLight = new PointLight("statueLight", new Vector3(0, 6, 0), scene);
    statueLight.intensity = 0.8;
    statueLight.diffuse = new Color3(0.9, 0.85, 0.7);
    statueLight.range = 15;
  }
  /**
   * Create the western portal with particle effect.
   */
  createPortal(scene) {
    this.portal = MeshBuilder.CreateTorus("portal", {
      diameter: this.PORTAL_RADIUS * 2,
      thickness: 0.3,
      tessellation: 32
    }, scene);
    this.portal.position = new Vector3(-this.ROOM_WIDTH / 2 + 0.5, this.PORTAL_RADIUS + 0.5, 0);
    this.portal.rotation.y = Math.PI / 2;
    const portalMat = new StandardMaterial("portalMat", scene);
    portalMat.diffuseColor = new Color3(0.3, 0.2, 0.5);
    portalMat.emissiveColor = new Color3(0.2, 0.1, 0.4);
    this.portal.material = portalMat;
    const portalCenter = MeshBuilder.CreateDisc("portalCenter", {
      radius: this.PORTAL_RADIUS - 0.3,
      tessellation: 32
    }, scene);
    portalCenter.position = this.portal.position.clone();
    portalCenter.rotation.y = Math.PI / 2;
    const centerMat = new StandardMaterial("centerMat", scene);
    centerMat.diffuseColor = new Color3(0.1, 0.05, 0.2);
    centerMat.emissiveColor = new Color3(0.15, 0.1, 0.3);
    centerMat.alpha = 0.7;
    portalCenter.material = centerMat;
    const portalLight = new PointLight("portalLight", this.portal.position.clone(), scene);
    portalLight.intensity = 1.2;
    portalLight.diffuse = new Color3(0.5, 0.3, 0.8);
    portalLight.range = 10;
    this.createPortalParticles(scene, this.portal.position);
  }
  /**
   * Create swirling particle effect for portal.
   */
  createPortalParticles(scene, position) {
    this.portalParticles = new ParticleSystem("portalParticles", 500, scene);
    const size = 32;
    const canvas = document.createElement("canvas");
    canvas.width = size;
    canvas.height = size;
    const ctx = canvas.getContext("2d");
    if (ctx) {
      const gradient = ctx.createRadialGradient(
        size / 2,
        size / 2,
        0,
        size / 2,
        size / 2,
        size / 2
      );
      gradient.addColorStop(0, "rgba(255, 255, 255, 1)");
      gradient.addColorStop(0.3, "rgba(200, 200, 255, 0.8)");
      gradient.addColorStop(0.7, "rgba(100, 50, 200, 0.3)");
      gradient.addColorStop(1, "rgba(0, 0, 0, 0)");
      ctx.fillStyle = gradient;
      ctx.fillRect(0, 0, size, size);
    }
    const particleTexture = new Texture(
      canvas.toDataURL(),
      scene,
      true,
      // noMipmap - prevents glGenerateMipmap error
      false
      // invertY
    );
    this.portalParticles.particleTexture = particleTexture;
    this.portalParticles.emitter = position;
    this.portalParticles.minEmitBox = new Vector3(-0.5, -1.5, -0.5);
    this.portalParticles.maxEmitBox = new Vector3(0.5, 1.5, 0.5);
    this.portalParticles.color1 = new Color4(0.5, 0.2, 0.8, 1);
    this.portalParticles.color2 = new Color4(0.2, 0.1, 0.5, 1);
    this.portalParticles.colorDead = new Color4(0, 0, 0.2, 0);
    this.portalParticles.minSize = 0.05;
    this.portalParticles.maxSize = 0.15;
    this.portalParticles.minLifeTime = 0.5;
    this.portalParticles.maxLifeTime = 1.5;
    this.portalParticles.emitRate = 100;
    this.portalParticles.gravity = new Vector3(0, 0.5, 0);
    this.portalParticles.direction1 = new Vector3(-0.5, 1, -0.5);
    this.portalParticles.direction2 = new Vector3(0.5, 1, 0.5);
    this.portalParticles.minAngularSpeed = 0;
    this.portalParticles.maxAngularSpeed = Math.PI;
    this.portalParticles.minEmitPower = 0.5;
    this.portalParticles.maxEmitPower = 1.5;
    this.portalParticles.updateSpeed = 0.01;
    this.portalParticles.start();
  }
  /**
   * Create FPS controller with floor collision.
   */
  createFPSController(scene) {
    const startPos = new Vector3(0, 1.7, -this.ROOM_DEPTH / 2 + 5);
    this.fpsController = new FirstPersonController(scene, startPos, {
      moveSpeed: 5,
      collisionTarget: this.floor,
      eyeHeight: 1.7
    });
    this.fpsController.activate();
  }
  /**
   * Check for portal proximity and trigger callback.
   */
  update(deltaTime) {
    if (!this.fpsController || !this.portal) return;
    this.fpsController.handleInput(deltaTime);
    if (this.portalEntered) return;
    const pos = this.fpsController.getCamera().position;
    const portalPos = this.portal.position;
    const distance = Vector3.Distance(pos, portalPos);
    if (distance < this.PORTAL_RADIUS + 1) {
      this.portalEntered = true;
      this.callbacks.onPortalEnter?.();
    }
  }
  /**
   * Clean up resources.
   */
  dispose() {
    console.log("[LobbyScene] Disposing");
    this.fpsController?.dispose();
    this.portalParticles?.dispose();
    this.scene = null;
    this.fpsController = null;
    this.floor = null;
    this.statue = null;
    this.portal = null;
  }
}
const css$1 = {
  code: ".simulation-layout.svelte-3oxgig{-moz-user-select:none;user-select:none;-webkit-user-select:none}.simulation-layout input,.simulation-layout .text-log{-moz-user-select:text;user-select:text;-webkit-user-select:text}",
  map: '{"version":3,"file":"SimulationModeLayout.svelte","sources":["SimulationModeLayout.svelte"],"sourcesContent":["<script lang=\\"ts\\">import { onMount, onDestroy } from \\"svelte\\";\\nimport { isMobile } from \\"$lib/stores/ui\\";\\nimport { gameStore } from \\"$lib/stores/game\\";\\nimport MessageOverlay from \\"$lib/components/HUD/MessageOverlay.svelte\\";\\nimport SceneCanvas from \\"$lib/components/Scene/SceneCanvas.svelte\\";\\nimport {\\n  sceneManager\\n} from \\"$lib/components/Scene/SceneManager\\";\\nimport { LobbyScene } from \\"$lib/components/Scene/LobbyScene\\";\\nimport WorldController from \\"$lib/components/Map/WorldController.svelte\\";\\nimport { ArcRotateCamera } from \\"@babylonjs/core/Cameras/arcRotateCamera\\";\\nimport { Vector3 } from \\"@babylonjs/core/Maths/math.vector\\";\\nlet showCommandOverlay = true;\\nlet textLogExpanded = false;\\nlet activeScene = null;\\nlet canvasReady = false;\\nconst lobbyScene = new LobbyScene();\\nlobbyScene.setCallbacks({\\n  onPortalEnter: () => {\\n    console.log(\\"Portal entered! Transitioning to WORLD...\\");\\n    gameStore.enterWorld(\\"new-world\\");\\n  }\\n});\\nsceneManager.registerSceneFactory(\\"LOBBY\\", lobbyScene);\\nsceneManager.registerSceneFactory(\\"WORLD\\", {\\n  create: async (scene) => {\\n    console.log(\\n      \\"[SimulationMode] Created WORLD scene with default camera\\"\\n    );\\n    const canvas = scene.getEngine().getRenderingCanvas();\\n    const defaultCamera = new ArcRotateCamera(\\n      \\"defaultCamera\\",\\n      Math.PI / 2,\\n      Math.PI / 3,\\n      5,\\n      new Vector3(0, 0, 0),\\n      scene\\n    );\\n    if (canvas) {\\n      defaultCamera.attachControl(canvas, true);\\n    }\\n    scene.activeCamera = defaultCamera;\\n  },\\n  dispose: () => {\\n    console.log(\\"[SimulationMode] Disposing WORLD scene\\");\\n  }\\n});\\nsceneManager.setOnLocationChange((loc) => {\\n  activeScene = sceneManager.getActiveScene();\\n});\\nfunction handleCanvasReady(event) {\\n  const canvas = event.detail;\\n  sceneManager.initialize(canvas);\\n  canvasReady = true;\\n  if ($gameStore.gameLocation === \\"LOADING\\") {\\n    gameStore.setGameLocation(\\"LOBBY\\");\\n  } else {\\n    sceneManager.transitionTo($gameStore.gameLocation);\\n  }\\n}\\n$: if (canvasReady && $gameStore.gameLocation && $gameStore.gameLocation !== \\"LOADING\\") {\\n  const currentLoc = sceneManager.getCurrentLocation();\\n  if (currentLoc !== $gameStore.gameLocation && !sceneManager.isInTransition()) {\\n    activeScene = null;\\n    sceneManager.transitionTo($gameStore.gameLocation).then(() => {\\n      activeScene = sceneManager.getActiveScene();\\n    });\\n  }\\n}\\n<\/script>\\n\\n<div\\n    class=\\"simulation-layout relative w-full h-screen bg-black overflow-hidden\\"\\n>\\n    <!-- 3D Canvas Container (Full Screen) -->\\n    <div class=\\"absolute inset-0 z-0\\">\\n        <SceneCanvas on:canvasReady={handleCanvasReady} />\\n\\n        <!-- Render WorldController when in WORLD mode -->\\n        {#if $gameStore.gameLocation === \\"WORLD\\" && activeScene}\\n            <WorldController\\n                scene={activeScene}\\n                globeTextureBlob={$gameStore.world.textureBlob}\\n                globeHeightmapBlob={$gameStore.world.heightmapBlob}\\n                materialBlob={$gameStore.world.materialBlob}\\n                iceBlob={$gameStore.world.iceBlob}\\n                seaLevel={$gameStore.world.geo.seaLevel}\\n                maxElevation={$gameStore.world.geo.maxElevation}\\n                minElevation={$gameStore.world.geo.minElevation}\\n                satellites={$gameStore.world.sim.satellites}\\n            />\\n        {/if}\\n    </div>\\n\\n    <!-- HUD Overlay Layer -->\\n    <div class=\\"absolute inset-0 z-10 pointer-events-none\\">\\n        <!-- Fading Messages Overlay -->\\n        <MessageOverlay />\\n\\n        <!-- Top Bar: Status + Mode Toggle -->\\n        <header\\n            class=\\"absolute top-0 left-0 right-0 h-14 flex items-center px-4 pointer-events-auto bg-gradient-to-b from-black/60 to-transparent\\"\\n        >\\n            <slot name=\\"status-bar\\">\\n                <div class=\\"text-gray-400 text-sm\\">Status</div>\\n            </slot>\\n\\n            <!-- Mode Toggle Button (right side) -->\\n            <div class=\\"ml-auto\\">\\n                <slot name=\\"mode-toggle\\" />\\n            </div>\\n        </header>\\n\\n        <!-- Left Side: Mini Stats (Desktop only) -->\\n        {#if !$isMobile}\\n            <aside class=\\"absolute top-16 left-4 w-48 pointer-events-auto\\">\\n                <div\\n                    class=\\"bg-gray-900/80 backdrop-blur-sm rounded-lg border border-gray-700/50 p-3 shadow-lg\\"\\n                >\\n                    <slot name=\\"hud-stats\\">\\n                        <div class=\\"text-gray-400 text-xs\\">Stats Overlay</div>\\n                    </slot>\\n                </div>\\n            </aside>\\n        {/if}\\n\\n        <!-- Right Side: Minimap (Desktop only) -->\\n        {#if !$isMobile}\\n            <aside class=\\"absolute top-16 right-4 pointer-events-auto\\">\\n                <div\\n                    class=\\"w-40 h-40 bg-gray-900/80 backdrop-blur-sm rounded-lg border border-gray-700/50 overflow-hidden shadow-lg\\"\\n                >\\n                    <slot name=\\"minimap\\">\\n                        <div\\n                            class=\\"w-full h-full flex items-center justify-center text-gray-500 text-xs\\"\\n                        >\\n                            Minimap\\n                        </div>\\n                    </slot>\\n                </div>\\n            </aside>\\n        {/if}\\n\\n        <!-- Bottom: Command Input + Text Log -->\\n        <div class=\\"absolute bottom-0 left-0 right-0 pointer-events-auto\\">\\n            <!-- Collapsible Text Log -->\\n            {#if textLogExpanded}\\n                <div\\n                    class=\\"mx-4 mb-2 max-h-64 bg-gray-900/90 backdrop-blur-sm rounded-t-lg border border-b-0 border-gray-700/50 overflow-hidden shadow-lg\\"\\n                >\\n                    <div\\n                        class=\\"flex items-center justify-between px-3 py-2 border-b border-gray-700/50\\"\\n                    >\\n                        <span\\n                            class=\\"text-xs font-semibold text-gray-400 uppercase tracking-wider\\"\\n                            >Game Log</span\\n                        >\\n                        <button\\n                            class=\\"text-gray-500 hover:text-gray-300 p-1\\"\\n                            on:click={() => (textLogExpanded = false)}\\n                            aria-label=\\"Collapse log\\"\\n                        >\\n                            <svg\\n                                class=\\"w-4 h-4\\"\\n                                fill=\\"none\\"\\n                                stroke=\\"currentColor\\"\\n                                viewBox=\\"0 0 24 24\\"\\n                            >\\n                                <path\\n                                    stroke-linecap=\\"round\\"\\n                                    stroke-linejoin=\\"round\\"\\n                                    stroke-width=\\"2\\"\\n                                    d=\\"M19 9l-7 7-7-7\\"\\n                                />\\n                            </svg>\\n                        </button>\\n                    </div>\\n                    <div class=\\"h-48 overflow-y-auto p-3\\">\\n                        <slot name=\\"text-log\\">\\n                            <div class=\\"text-gray-500 text-sm italic\\">\\n                                No messages yet...\\n                            </div>\\n                        </slot>\\n                    </div>\\n                </div>\\n            {/if}\\n\\n            <!-- Command Input Bar -->\\n            <div\\n                class=\\"bg-gray-900/90 backdrop-blur-sm border-t border-gray-700/50 p-3\\"\\n            >\\n                <div class=\\"flex items-center gap-2\\">\\n                    <!-- Expand Log Button -->\\n                    {#if !textLogExpanded}\\n                        <button\\n                            class=\\"p-2 text-gray-400 hover:text-gray-200 hover:bg-gray-800/50 rounded transition-colors\\"\\n                            on:click={() => (textLogExpanded = true)}\\n                            aria-label=\\"Expand game log\\"\\n                            title=\\"Show game log\\"\\n                        >\\n                            <svg\\n                                class=\\"w-5 h-5\\"\\n                                fill=\\"none\\"\\n                                stroke=\\"currentColor\\"\\n                                viewBox=\\"0 0 24 24\\"\\n                            >\\n                                <path\\n                                    stroke-linecap=\\"round\\"\\n                                    stroke-linejoin=\\"round\\"\\n                                    stroke-width=\\"2\\"\\n                                    d=\\"M4 6h16M4 12h16M4 18h16\\"\\n                                />\\n                            </svg>\\n                        </button>\\n                    {/if}\\n\\n                    <!-- Command Input -->\\n                    <div class=\\"flex-1\\">\\n                        <slot name=\\"command-input\\">\\n                            <div class=\\"text-gray-500 text-sm\\">Input</div>\\n                        </slot>\\n                    </div>\\n                </div>\\n            </div>\\n        </div>\\n\\n        <!-- Mobile: Touch Controls Overlay -->\\n        {#if $isMobile}\\n            <div class=\\"absolute bottom-24 right-4 pointer-events-auto\\">\\n                <slot name=\\"controls\\">\\n                    <div\\n                        class=\\"w-24 h-24 bg-gray-900/60 rounded-full flex items-center justify-center\\"\\n                    >\\n                        <span class=\\"text-gray-500 text-xs\\">D-Pad</span>\\n                    </div>\\n                </slot>\\n            </div>\\n        {/if}\\n    </div>\\n</div>\\n\\n<style>\\n    .simulation-layout {\\n        /* Prevent text selection on HUD */\\n        -moz-user-select: none;\\n             user-select: none;\\n        -webkit-user-select: none;\\n    }\\n\\n    /* Allow text selection in text log and input */\\n    :global(.simulation-layout input),\\n    :global(.simulation-layout .text-log) {\\n        -moz-user-select: text;\\n             user-select: text;\\n        -webkit-user-select: text;\\n    }\\n</style>\\n"],"names":[],"mappings":"AAkPI,gCAAmB,CAEf,gBAAgB,CAAE,IAAI,CACjB,WAAW,CAAE,IAAI,CACtB,mBAAmB,CAAE,IACzB,CAGQ,wBAAyB,CACzB,4BAA8B,CAClC,gBAAgB,CAAE,IAAI,CACjB,WAAW,CAAE,IAAI,CACtB,mBAAmB,CAAE,IACzB"}'
};
const SimulationModeLayout = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $gameStore, $$unsubscribe_gameStore;
  let $isMobile, $$unsubscribe_isMobile;
  $$unsubscribe_gameStore = subscribe(gameStore, (value) => $gameStore = value);
  $$unsubscribe_isMobile = subscribe(isMobile, (value) => $isMobile = value);
  let activeScene = null;
  const lobbyScene = new LobbyScene();
  lobbyScene.setCallbacks({
    onPortalEnter: () => {
      console.log("Portal entered! Transitioning to WORLD...");
      gameStore.enterWorld("new-world");
    }
  });
  sceneManager.registerSceneFactory("LOBBY", lobbyScene);
  sceneManager.registerSceneFactory("WORLD", {
    create: async (scene) => {
      console.log("[SimulationMode] Created WORLD scene with default camera");
      const canvas = scene.getEngine().getRenderingCanvas();
      const defaultCamera = new ArcRotateCamera("defaultCamera", Math.PI / 2, Math.PI / 3, 5, new Vector3(0, 0, 0), scene);
      if (canvas) {
        defaultCamera.attachControl(canvas, true);
      }
      scene.activeCamera = defaultCamera;
    },
    dispose: () => {
      console.log("[SimulationMode] Disposing WORLD scene");
    }
  });
  sceneManager.setOnLocationChange((loc) => {
    activeScene = sceneManager.getActiveScene();
  });
  $$result.css.add(css$1);
  $$unsubscribe_gameStore();
  $$unsubscribe_isMobile();
  return `<div class="simulation-layout relative w-full h-screen bg-black overflow-hidden svelte-3oxgig"> <div class="absolute inset-0 z-0">${validate_component(SceneCanvas, "SceneCanvas").$$render($$result, {}, {}, {})}  ${$gameStore.gameLocation === "WORLD" && activeScene ? `${validate_component(WorldController, "WorldController").$$render(
    $$result,
    {
      scene: activeScene,
      globeTextureBlob: $gameStore.world.textureBlob,
      globeHeightmapBlob: $gameStore.world.heightmapBlob,
      materialBlob: $gameStore.world.materialBlob,
      iceBlob: $gameStore.world.iceBlob,
      seaLevel: $gameStore.world.geo.seaLevel,
      maxElevation: $gameStore.world.geo.maxElevation,
      minElevation: $gameStore.world.geo.minElevation,
      satellites: $gameStore.world.sim.satellites
    },
    {},
    {}
  )}` : ``}</div>  <div class="absolute inset-0 z-10 pointer-events-none"> ${validate_component(MessageOverlay, "MessageOverlay").$$render($$result, {}, {}, {})}  <header class="absolute top-0 left-0 right-0 h-14 flex items-center px-4 pointer-events-auto bg-gradient-to-b from-black/60 to-transparent">${slots["status-bar"] ? slots["status-bar"]({}) : ` <div class="text-gray-400 text-sm" data-svelte-h="svelte-dh9p5p">Status</div> `}  <div class="ml-auto">${slots["mode-toggle"] ? slots["mode-toggle"]({}) : ``}</div></header>  ${!$isMobile ? `<aside class="absolute top-16 left-4 w-48 pointer-events-auto"><div class="bg-gray-900/80 backdrop-blur-sm rounded-lg border border-gray-700/50 p-3 shadow-lg">${slots["hud-stats"] ? slots["hud-stats"]({}) : ` <div class="text-gray-400 text-xs" data-svelte-h="svelte-1dpjz0p">Stats Overlay</div> `}</div></aside>` : ``}  ${!$isMobile ? `<aside class="absolute top-16 right-4 pointer-events-auto"><div class="w-40 h-40 bg-gray-900/80 backdrop-blur-sm rounded-lg border border-gray-700/50 overflow-hidden shadow-lg">${slots.minimap ? slots.minimap({}) : ` <div class="w-full h-full flex items-center justify-center text-gray-500 text-xs" data-svelte-h="svelte-1trdgha">Minimap</div> `}</div></aside>` : ``}  <div class="absolute bottom-0 left-0 right-0 pointer-events-auto"> ${``}  <div class="bg-gray-900/90 backdrop-blur-sm border-t border-gray-700/50 p-3"><div class="flex items-center gap-2"> ${`<button class="p-2 text-gray-400 hover:text-gray-200 hover:bg-gray-800/50 rounded transition-colors" aria-label="Expand game log" title="Show game log" data-svelte-h="svelte-1k4ilpi"><svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path></svg></button>`}  <div class="flex-1">${slots["command-input"] ? slots["command-input"]({}) : ` <div class="text-gray-500 text-sm" data-svelte-h="svelte-1lwgc9w">Input</div> `}</div></div></div></div>  ${$isMobile ? `<div class="absolute bottom-24 right-4 pointer-events-auto">${slots.controls ? slots.controls({}) : ` <div class="w-24 h-24 bg-gray-900/60 rounded-full flex items-center justify-center" data-svelte-h="svelte-1nta38e"><span class="text-gray-500 text-xs">D-Pad</span></div> `}</div>` : ``}</div> </div>`;
});
const GameContainer = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $interfaceMode, $$unsubscribe_interfaceMode;
  $$unsubscribe_interfaceMode = subscribe(interfaceMode, (value) => $interfaceMode = value);
  $$unsubscribe_interfaceMode();
  return `<div class="game-container w-full h-full" data-testid="game-container"${add_attribute("data-mode", $interfaceMode, 0)}>${$interfaceMode === "TEXT" ? `${validate_component(MUDModeLayout, "MUDModeLayout").$$render($$result, {}, {}, {
    "mode-toggle": () => {
      return `${slots["mode-toggle"] ? slots["mode-toggle"]({ slot: "mode-toggle" }) : ``}`;
    },
    controls: () => {
      return `${slots.controls ? slots.controls({ slot: "controls" }) : ``}`;
    },
    "right-panel": () => {
      return `${slots["right-panel"] ? slots["right-panel"]({ slot: "right-panel" }) : ``}`;
    },
    "left-panel": () => {
      return `${slots["left-panel"] ? slots["left-panel"]({ slot: "left-panel" }) : ``}`;
    },
    "command-input": () => {
      return `${slots["command-input"] ? slots["command-input"]({ slot: "command-input" }) : ``}`;
    },
    "main-display": () => {
      return `${slots["main-display"] ? slots["main-display"]({ slot: "main-display" }) : ``}`;
    },
    "status-bar": () => {
      return `${slots["status-bar"] ? slots["status-bar"]({ slot: "status-bar" }) : ``}`;
    }
  })}` : `${validate_component(SimulationModeLayout, "SimulationModeLayout").$$render($$result, {}, {}, {
    "mode-toggle": () => {
      return `${slots["mode-toggle"] ? slots["mode-toggle"]({ slot: "mode-toggle" }) : ``}`;
    },
    controls: () => {
      return `${slots.controls ? slots.controls({ slot: "controls" }) : ``}`;
    },
    minimap: () => {
      return `${slots.minimap ? slots.minimap({ slot: "minimap" }) : ``}`;
    },
    "hud-stats": () => {
      return `${slots["hud-stats"] ? slots["hud-stats"]({ slot: "hud-stats" }) : ``}`;
    },
    "text-log": () => {
      return `${slots["text-log"] ? slots["text-log"]({ slot: "text-log" }) : ``}`;
    },
    "command-input": () => {
      return `${slots["command-input"] ? slots["command-input"]({ slot: "command-input" }) : ``}`;
    },
    "status-bar": () => {
      return `${slots["status-bar"] ? slots["status-bar"]({ slot: "status-bar" }) : ``}`;
    },
    canvas: () => {
      return `${slots.canvas ? slots.canvas({ slot: "canvas" }) : ``}`;
    }
  })}`}</div>`;
});
const MOVEMENT_ALIASES = {
  "n": "north",
  "s": "south",
  "e": "east",
  "w": "west",
  "u": "up",
  "d": "down",
  "ne": "northeast",
  "nw": "northwest",
  "se": "southeast",
  "sw": "southwest"
};
const VALID_DIRECTIONS = /* @__PURE__ */ new Set([
  "north",
  "south",
  "east",
  "west",
  "up",
  "down",
  "northeast",
  "northwest",
  "southeast",
  "southwest"
]);
function parseMovementCommand(command) {
  const trimmed = command.trim().toLowerCase();
  const aliasResult = MOVEMENT_ALIASES[trimmed];
  if (aliasResult) {
    return aliasResult;
  }
  if (VALID_DIRECTIONS.has(trimmed)) {
    return trimmed;
  }
  const goMatch = trimmed.match(/^go\s+(.+)$/);
  if (goMatch && goMatch[1]) {
    const dir = goMatch[1];
    if (VALID_DIRECTIONS.has(dir)) return dir;
    const aliasDir = MOVEMENT_ALIASES[dir];
    if (aliasDir) return aliasDir;
  }
  const moveMatch = trimmed.match(/^move\s+(.+)$/);
  if (moveMatch && moveMatch[1]) {
    const dir = moveMatch[1];
    if (VALID_DIRECTIONS.has(dir)) return dir;
    const aliasDir = MOVEMENT_ALIASES[dir];
    if (aliasDir) return aliasDir;
  }
  return null;
}
class GameSystem {
  fpsController = null;
  currentMode = "TEXT";
  unsubscribe = null;
  constructor() {
    this.unsubscribe = interfaceMode.subscribe((mode) => {
      this.currentMode = mode;
    });
  }
  /**
   * Process a typed command (from CommandInput).
   * Routes to appropriate handlers based on command type.
   */
  processCommand(command) {
    if (!command.trim()) return;
    const direction = parseMovementCommand(command);
    if (direction) {
      this.processMovement(direction);
      return;
    }
    gameWebSocket.sendRawCommand(command);
  }
  /**
   * Process a movement command (from text, keyboard, or DPad).
   * @param direction - Normalized direction (north, south, etc.)
   */
  processMovement(direction) {
    gameWebSocket.sendRawCommand(`go ${direction}`);
    if (this.currentMode === "VISUAL" && this.fpsController) {
      this.triggerFPSMovement(direction);
    }
    this.logMovement(direction);
  }
  /**
   * Trigger FPS controller movement based on direction.
   */
  triggerFPSMovement(direction) {
    if (!this.fpsController) return;
    const pos = this.fpsController.getPosition();
    console.log(`[GameSystem] FPS movement ${direction} from (${pos.lat.toFixed(2)}, ${pos.lon.toFixed(2)})`);
  }
  /**
   * Log a movement action to the game output.
   */
  logMovement(direction) {
    addGameMessage({
      id: crypto.randomUUID(),
      type: "movement",
      text: `You head ${direction}.`,
      timestamp: /* @__PURE__ */ new Date(),
      direction
    });
  }
  /**
   * Handle FPS controller position updates.
   * Called when player moves in 3D space (e.g., from mouse look + WASD).
   * Generates text log entries for position changes.
   */
  onFPSPositionUpdate(newPosition) {
    if (this.currentMode !== "VISUAL") return;
    addGameMessage({
      id: crypto.randomUUID(),
      type: "movement",
      text: `You are at coordinates (${newPosition.lat.toFixed(2)}°, ${newPosition.lon.toFixed(2)}°)`,
      timestamp: /* @__PURE__ */ new Date()
    });
  }
  /**
   * Process a keyboard key press (from global keyboard handler).
   * Returns true if the key was handled.
   */
  processKeyPress(key, shiftKey = false) {
    if (this.currentMode !== "VISUAL") return false;
    let direction = null;
    switch (key.toLowerCase()) {
      case "w":
      case "arrowup":
        direction = "north";
        break;
      case "s":
      case "arrowdown":
        direction = "south";
        break;
      case "a":
      case "arrowleft":
        direction = "west";
        break;
      case "d":
      case "arrowright":
        direction = "east";
        break;
    }
    if (direction) {
      this.processMovement(direction);
      return true;
    }
    return false;
  }
  /**
   * Attach the FPS controller for 3D movement synchronization.
   */
  setFPSController(controller) {
    this.fpsController = controller;
  }
  /**
   * Get the current interface mode.
   */
  getMode() {
    return this.currentMode;
  }
  /**
   * Clean up subscriptions.
   */
  dispose() {
    if (this.unsubscribe) {
      this.unsubscribe();
      this.unsubscribe = null;
    }
    this.fpsController = null;
  }
}
new GameSystem();
const css = {
  code: ".messages-container.svelte-lcmnji::-webkit-scrollbar{width:8px}.messages-container.svelte-lcmnji::-webkit-scrollbar-track{background:rgba(0, 0, 0, 0.3)}.messages-container.svelte-lcmnji::-webkit-scrollbar-thumb{background:rgba(55, 65, 81, 0.5);border-radius:4px}.messages-container.svelte-lcmnji::-webkit-scrollbar-thumb:hover{background:rgba(75, 85, 99, 0.7)}",
  map: '{"version":3,"file":"+page.svelte","sources":["+page.svelte"],"sourcesContent":["<script lang=\\"ts\\">import { onMount, onDestroy, tick } from \\"svelte\\";\\nimport { goto } from \\"$app/navigation\\";\\nimport { gameAPI } from \\"$lib/services/api\\";\\nimport { gameWebSocket } from \\"$lib/services/websocket\\";\\nimport WorldEntry from \\"$lib/components/WorldEntry.svelte\\";\\nimport CharacterSheet from \\"$lib/components/Character/CharacterSheet.svelte\\";\\nimport WorldMapModal from \\"$lib/components/Map/WorldMapModal.svelte\\";\\nimport CommandInput from \\"$lib/components/Input/CommandInput.svelte\\";\\nimport QuickButtons from \\"$lib/components/Input/QuickButtons.svelte\\";\\nimport ModeToggle from \\"$lib/components/Layout/ModeToggle.svelte\\";\\nimport GameContainer from \\"$lib/components/Layout/GameContainer.svelte\\";\\nimport { gameSystem } from \\"$lib/services/GameSystem\\";\\nimport { interfaceMode } from \\"$lib/stores/ui\\";\\nlet onboardingStep = \\"checking\\";\\nlet interviewSessionId = null;\\nlet currentQuestion = \\"\\";\\nlet userResponse = \\"\\";\\nlet conversationHistory = [];\\nlet showEntryModal = false;\\nlet entryOptions = null;\\nlet targetWorldId = null;\\nlet messages = [];\\nlet commandInput = \\"\\";\\nlet unsubscribe = null;\\nlet isConnected = false;\\nlet currentCharacterId = null;\\nlet showCharacterSheet = false;\\nlet characterSkills = {};\\nlet characterStats = {\\n  name: \\"Adventurer\\",\\n  level: 1,\\n  experience: 0,\\n  nextLevelXP: 100,\\n  attributes: {\\n    strength: 10,\\n    dexterity: 10,\\n    constitution: 10,\\n    intelligence: 10,\\n    wisdom: 10,\\n    charisma: 10\\n  }\\n};\\nlet showWorldMap = false;\\nlet latestSimEvent = null;\\nlet currentUser = null;\\nconst API_URL = \\"/api\\";\\nonMount(async () => {\\n  try {\\n    currentUser = await gameAPI.getMe();\\n  } catch (err) {\\n    goto(\\"/\\");\\n    return;\\n  }\\n  const unsubscribeConnection = gameWebSocket.connected.subscribe(\\n    (value) => {\\n      isConnected = value;\\n    }\\n  );\\n  unsubscribe = gameWebSocket.onMessage(handleServerMessage);\\n  const originalUnsubscribe = unsubscribe;\\n  unsubscribe = () => {\\n    if (originalUnsubscribe) originalUnsubscribe();\\n    unsubscribeConnection();\\n  };\\n  await checkOnboardingStatus();\\n});\\nonDestroy(() => {\\n  if (unsubscribe) unsubscribe();\\n});\\nasync function checkOnboardingStatus() {\\n  try {\\n    const interviewRes = await fetch(\\n      `${API_URL}/world/interview/active`,\\n      {\\n        credentials: \\"include\\"\\n        // Send cookies\\n      }\\n    );\\n    if (interviewRes.ok) {\\n      const interview = await interviewRes.json();\\n      if (interview && interview.status === \\"in_progress\\") {\\n        onboardingStep = \\"interview\\";\\n        interviewSessionId = interview.session_id;\\n        if (interview.question === \\"The interview is already complete.\\" || interview.question === \\"This interview is already complete.\\") {\\n          addMessage(\\n            \\"system\\",\\n            \\"World interview previously completed.\\"\\n          );\\n          setTimeout(() => {\\n            joinLobby();\\n          }, 1e3);\\n          return;\\n        }\\n        conversationHistory = interview.conversation || [];\\n        if (conversationHistory.length > 0) {\\n          const lastMessage = conversationHistory[conversationHistory.length - 1];\\n          if (lastMessage.role === \\"assistant\\") {\\n            currentQuestion = lastMessage.text;\\n          }\\n        } else if (interview.question) {\\n          currentQuestion = interview.question;\\n          addMessage(\\"interview\\", currentQuestion);\\n        }\\n        addMessage(\\n          \\"system\\",\\n          \\"Resuming your world creation interview...\\"\\n        );\\n        return;\\n      }\\n    }\\n    if (currentUser?.last_world_id) {\\n      try {\\n        const data = await gameAPI.getCharacters();\\n        if (data && data.characters) {\\n          const char = data.characters.find(\\n            (c) => c.world_id === currentUser?.last_world_id\\n          );\\n          if (char) {\\n            await joinGame(char.character_id);\\n            return;\\n          }\\n        }\\n      } catch (e) {\\n        console.error(\\"Auto-resume check failed\\", e);\\n      }\\n    }\\n    await joinLobby();\\n  } catch (error) {\\n    console.error(\\"Onboarding check failed:\\", error);\\n    onboardingStep = \\"game\\";\\n    addMessage(\\n      \\"error\\",\\n      \\"Failed to check status. Starting in game mode.\\"\\n    );\\n    joinLobby();\\n  }\\n}\\nasync function joinLobby() {\\n  console.log(\\"[Lobby] Joining lobby...\\");\\n  try {\\n    console.log(\\"[Lobby] Disconnecting existing WebSocket...\\");\\n    gameWebSocket.disconnect();\\n    await new Promise((resolve) => setTimeout(resolve, 100));\\n    console.log(\\"[Lobby] Connecting to WebSocket...\\");\\n    gameWebSocket.connect();\\n    onboardingStep = \\"lobby\\";\\n    console.log(\\"[Lobby] Lobby join complete\\");\\n  } catch (error) {\\n    console.error(\\"[Lobby] Failed to join lobby:\\", error);\\n    addMessage(\\"error\\", \\"Failed to join lobby.\\");\\n  }\\n}\\nasync function startWorldInterview() {\\n  try {\\n    const response = await fetch(`${API_URL}/world/interview/start`, {\\n      method: \\"POST\\",\\n      headers: {\\n        \\"Content-Type\\": \\"application/json\\"\\n      },\\n      credentials: \\"include\\"\\n      // Send cookies\\n    });\\n    if (!response.ok) {\\n      const errorText = await response.text();\\n      console.error(\\n        \\"Interview start failed:\\",\\n        response.status,\\n        errorText\\n      );\\n      throw new Error(\\n        `Failed to start interview: ${response.status}`\\n      );\\n    }\\n    const data = await response.json();\\n    console.log(\\"Interview started:\\", data);\\n    if (!data.session_id || !data.question) {\\n      console.error(\\"Invalid interview response:\\", data);\\n      throw new Error(\\"Invalid response from server\\");\\n    }\\n    interviewSessionId = data.session_id;\\n    currentQuestion = data.question || \\"Tell me about the world you\'d like to create.\\";\\n    conversationHistory.push({\\n      role: \\"assistant\\",\\n      text: currentQuestion\\n    });\\n    addMessage(\\"interview\\", currentQuestion);\\n    onboardingStep = \\"interview\\";\\n  } catch (error) {\\n    console.error(\\"Failed to start interview:\\", error);\\n    addMessage(\\n      \\"error\\",\\n      error.message || \\"Failed to start world interview. Please try again.\\"\\n    );\\n  }\\n}\\nasync function sendInterviewResponse() {\\n  if (!userResponse.trim() || !interviewSessionId) return;\\n  const userMessage = userResponse.trim();\\n  conversationHistory.push({ role: \\"user\\", text: userMessage });\\n  addMessage(\\"user\\", userMessage);\\n  userResponse = \\"\\";\\n  try {\\n    addMessage(\\"system\\", \\"Thinking... (this may take 10-20 seconds)\\");\\n    const controller = new AbortController();\\n    const timeoutId = setTimeout(() => controller.abort(), 6e4);\\n    const response = await fetch(`${API_URL}/world/interview/message`, {\\n      method: \\"POST\\",\\n      headers: {\\n        \\"Content-Type\\": \\"application/json\\"\\n      },\\n      credentials: \\"include\\",\\n      // Send cookies\\n      body: JSON.stringify({\\n        session_id: interviewSessionId,\\n        message: userMessage\\n      }),\\n      signal: controller.signal\\n    });\\n    clearTimeout(timeoutId);\\n    if (!response.ok) {\\n      const errorText = await response.text();\\n      console.error(\\n        \\"Interview message failed:\\",\\n        response.status,\\n        errorText\\n      );\\n      throw new Error(`Failed to send message: ${response.status}`);\\n    }\\n    const data = await response.json();\\n    console.log(\\"Interview response:\\", data);\\n    if (data.completed || data.question === \\"The interview is already complete.\\" || data.question === \\"This interview is already complete.\\") {\\n      addMessage(\\n        \\"system\\",\\n        \\"World interview complete! Creating your world...\\"\\n      );\\n      if (data.question && !data.question.includes(\\"already complete\\")) {\\n        conversationHistory.push({\\n          role: \\"assistant\\",\\n          text: data.question\\n        });\\n        addMessage(\\"interview\\", data.question);\\n      }\\n      setTimeout(() => {\\n        joinLobby();\\n        addMessage(\\n          \\"system\\",\\n          \\"World created! Use \'worlds\' to see it.\\"\\n        );\\n      }, 2e3);\\n    } else {\\n      const nextQuestion = data.next_question || data.question || data.response || \\"Please continue...\\";\\n      currentQuestion = nextQuestion;\\n      conversationHistory.push({\\n        role: \\"assistant\\",\\n        text: nextQuestion\\n      });\\n      addMessage(\\"interview\\", nextQuestion);\\n    }\\n  } catch (error) {\\n    console.error(\\"Failed to send response:\\", error);\\n    addMessage(\\n      \\"error\\",\\n      error.message || \\"Failed to process your response. Please try again.\\"\\n    );\\n  }\\n}\\nasync function handleCommand(cmd) {\\n  const input = cmd ? cmd.trim() : commandInput.trim();\\n  console.log(\\"[handleCommand] Called with:\\", {\\n    cmd,\\n    commandInput,\\n    input,\\n    onboardingStep,\\n    isConnected\\n  });\\n  if (!input) {\\n    console.log(\\"[handleCommand] Empty input, returning\\");\\n    return;\\n  }\\n  addMessage(\\"player\\", `> ${input}`);\\n  if (onboardingStep === \\"interview\\") {\\n    console.log(\\n      \\"[handleCommand] In interview mode, redirecting to sendInterviewResponse\\"\\n    );\\n    userResponse = input;\\n    commandInput = \\"\\";\\n    sendInterviewResponse();\\n    return;\\n  }\\n  if (input.toLowerCase().startsWith(\\"create character\\")) {\\n    await handleCharacterCommand(input);\\n    if (!cmd) commandInput = \\"\\";\\n    return;\\n  }\\n  if (input.trim().toLowerCase() === \\"status\\") {\\n    openCharacterSheet();\\n    if (!cmd) commandInput = \\"\\";\\n    return;\\n  }\\n  console.log(\\"[handleCommand] Routing through GameSystem:\\", input);\\n  gameSystem.processCommand(input);\\n  console.log(\\"[handleCommand] Command processed by GameSystem\\");\\n  if (!cmd) commandInput = \\"\\";\\n}\\nasync function listWorlds() {\\n  try {\\n    const res = await fetch(`${API_URL}/game/worlds`, {\\n      credentials: \\"include\\"\\n      // Send cookies\\n    });\\n    if (res.ok) {\\n      const worlds = await res.json();\\n      addMessage(\\"system\\", \\"Available Worlds:\\");\\n      worlds.forEach((w) => {\\n        addMessage(\\"system\\", `- ${w.Name} (ID: ${w.ID})`);\\n      });\\n      addMessage(\\"system\\", \\"Type \'enter <world_id>\' to enter.\\");\\n    } else {\\n      addMessage(\\"error\\", \\"Failed to list worlds.\\");\\n    }\\n  } catch (e) {\\n    addMessage(\\"error\\", \\"Failed to list worlds.\\");\\n  }\\n}\\nasync function enterWorld(worldId) {\\n  try {\\n    addMessage(\\"system\\", `Entering world ${worldId} as watcher...`);\\n    targetWorldId = worldId;\\n    await createCharacter(\\n      \\"Watcher\\",\\n      \\"Spirit\\",\\n      worldId,\\n      \\"An invisible observer.\\",\\n      \\"Watcher\\",\\n      \\"watcher\\"\\n    );\\n  } catch (e) {\\n    addMessage(\\"error\\", \\"Failed to enter world.\\");\\n  }\\n}\\nasync function handleEntrySelection(event) {\\n  const { type, data } = event.detail;\\n  if (!targetWorldId) return;\\n  showEntryModal = false;\\n  try {\\n    if (type === \\"cancel\\") {\\n      targetWorldId = null;\\n      return;\\n    }\\n    if (type === \\"watcher\\") {\\n      await createCharacter(\\n        \\"Watcher\\",\\n        \\"Spirit\\",\\n        targetWorldId,\\n        \\"An invisible observer.\\",\\n        \\"Watcher\\",\\n        \\"watcher\\"\\n      );\\n      return;\\n    }\\n    if (type === \\"npc\\") {\\n      const npc = data;\\n      await createCharacter(\\n        npc.name,\\n        npc.species,\\n        targetWorldId,\\n        npc.description,\\n        npc.occupation,\\n        \\"player\\",\\n        npc.appearance\\n      );\\n    }\\n    if (type === \\"custom\\") {\\n      onboardingStep = \\"character\\";\\n      addMessage(\\n        \\"system\\",\\n        \\"Enter: create character <name> <species>\\"\\n      );\\n    }\\n  } catch (e) {\\n    console.error(e);\\n    addMessage(\\"error\\", \\"Entry failed.\\");\\n  }\\n}\\nasync function createCharacter(name, species, worldId, description, occupation, role, appearance) {\\n  try {\\n    const response = await fetch(`${API_URL}/game/characters`, {\\n      method: \\"POST\\",\\n      headers: {\\n        \\"Content-Type\\": \\"application/json\\"\\n      },\\n      credentials: \\"include\\",\\n      // Send cookies\\n      body: JSON.stringify({\\n        world_id: worldId,\\n        name,\\n        species,\\n        role,\\n        description,\\n        occupation,\\n        appearance\\n      })\\n    });\\n    if (!response.ok) {\\n      const errorData = await response.json();\\n      let errorMessage = \\"Failed to create character\\";\\n      if (typeof errorData.error === \\"string\\") {\\n        errorMessage = errorData.error;\\n      } else if (errorData.error && errorData.error.message) {\\n        errorMessage = errorData.error.message;\\n      }\\n      throw new Error(errorMessage);\\n    }\\n    const data = await response.json();\\n    addMessage(\\n      \\"system\\",\\n      `Character \\"${name}\\" created! Joining world...`\\n    );\\n    await joinGame(data.character.character_id);\\n  } catch (error) {\\n    addMessage(\\"error\\", error.message);\\n  }\\n}\\nasync function joinGame(characterId) {\\n  try {\\n    const joinResponse = await fetch(`${API_URL}/game/join`, {\\n      method: \\"POST\\",\\n      headers: {\\n        \\"Content-Type\\": \\"application/json\\"\\n      },\\n      credentials: \\"include\\",\\n      // Send cookies\\n      body: JSON.stringify({\\n        character_id: characterId\\n      })\\n    });\\n    if (!joinResponse.ok) {\\n      const joinError = await joinResponse.json();\\n      console.error(\\"Failed to join game:\\", joinError);\\n      throw new Error(\\"Failed to join game session.\\");\\n    }\\n    const joinData = await joinResponse.json();\\n    gameWebSocket.disconnect();\\n    gameWebSocket.connect(characterId);\\n    currentCharacterId = characterId;\\n    onboardingStep = \\"game\\";\\n    if (joinData.message) {\\n      addMessage(\\"system\\", joinData.message);\\n    } else {\\n      addMessage(\\"system\\", \\"You have entered the world!\\");\\n    }\\n  } catch (e) {\\n    throw e;\\n  }\\n}\\nasync function handleCharacterCommand(cmd) {\\n  const parts = cmd.toLowerCase().split(\\" \\");\\n  if (parts[0] === \\"create\\" && parts[1] === \\"character\\" && parts.length >= 4) {\\n    const name = parts[2];\\n    const species = parts[3].charAt(0).toUpperCase() + parts[3].slice(1);\\n    const worldId = targetWorldId || \\"00000000-0000-0000-0000-000000000001\\";\\n    await createCharacter(name, species, worldId);\\n  } else {\\n    addMessage(\\n      \\"error\\",\\n      \\"Invalid command. Use: create character <name> <species>\\"\\n    );\\n  }\\n}\\nfunction handleServerMessage(msg) {\\n  switch (msg.type) {\\n    case \\"game_message\\":\\n      const type = msg.data.type;\\n      if (type === \\"trigger_entry_options\\") {\\n        const worldId = msg.data.metadata?.world_id || msg.data.text;\\n        enterWorld(worldId);\\n      } else if (type === \\"start_interview\\") {\\n        startWorldInterview();\\n      } else {\\n        addMessage(type, msg.data.text);\\n        if (type === \\"system\\" && msg.data.text?.includes(\\"Simulation Complete\\")) {\\n          console.log(\\n            \\"Simulation Complete detected, opening world map.\\"\\n          );\\n          showWorldMap = true;\\n        }\\n      }\\n      break;\\n    case \\"error\\":\\n      addMessage(\\"error\\", msg.data.message);\\n      break;\\n    case \\"sim_event\\":\\n      latestSimEvent = msg;\\n      if (msg.data.text && (msg.data.text.includes(\\"Simulation finished\\") || msg.data.text.includes(\\"Simulation Complete\\"))) {\\n        console.log(\\n          \\"Simulation finished detected via sim_event, opening map.\\"\\n        );\\n        showWorldMap = true;\\n      }\\n      if (msg.data.importance && msg.data.importance === \\"high\\") {\\n        addMessage(\\n          \\"system\\",\\n          `[WORLD EVENT] ${msg.data.description || msg.data.text}`\\n        );\\n      }\\n      break;\\n  }\\n}\\nasync function addMessage(type, text) {\\n  let shouldScroll = false;\\n  const container = document.getElementById(\\"game-output\\");\\n  if (container) {\\n    const { scrollTop, scrollHeight, clientHeight } = container;\\n    if (scrollHeight - scrollTop - clientHeight < 100) {\\n      shouldScroll = true;\\n    }\\n  } else {\\n    shouldScroll = true;\\n  }\\n  messages = [\\n    ...messages,\\n    {\\n      id: Date.now().toString() + Math.random(),\\n      type,\\n      text,\\n      timestamp: /* @__PURE__ */ new Date()\\n    }\\n  ];\\n  await tick();\\n  if (shouldScroll && container) {\\n    container.scrollTop = container.scrollHeight;\\n  }\\n}\\nfunction handleLogout() {\\n  gameWebSocket.disconnect();\\n  gameAPI.logout();\\n  goto(\\"/\\");\\n}\\nfunction getMessageColor(type) {\\n  switch (type) {\\n    case \\"error\\":\\n      return \\"text-red-400\\";\\n    case \\"system\\":\\n      return \\"text-cyan-400\\";\\n    case \\"interview\\":\\n      return \\"text-purple-400\\";\\n    case \\"command\\":\\n      return \\"text-gray-500\\";\\n    case \\"user\\":\\n      return \\"text-blue-400\\";\\n    case \\"success\\":\\n      return \\"text-green-400\\";\\n    default:\\n      return \\"text-gray-300\\";\\n  }\\n}\\nasync function openCharacterSheet() {\\n  if (!currentCharacterId) {\\n    addMessage(\\"error\\", \\"No active character.\\");\\n    return;\\n  }\\n  try {\\n    const data = await gameAPI.getSkills(currentCharacterId);\\n    characterSkills = data.skills || {};\\n    showCharacterSheet = true;\\n  } catch (e) {\\n    console.error(e);\\n    addMessage(\\"error\\", \\"Failed to load skills.\\");\\n  }\\n}\\n<\/script>\\n\\n<GameContainer>\\n    <!-- Status Bar (shared by both modes) -->\\n    <div slot=\\"status-bar\\" class=\\"flex items-center justify-between w-full\\">\\n        <h1 class=\\"text-xl font-bold text-blue-400\\">Thousand Worlds</h1>\\n        <div class=\\"flex items-center gap-4\\">\\n            <!-- Connection Status -->\\n            <div\\n                class=\\"flex items-center gap-2 text-sm\\"\\n                title={isConnected ? \\"Connected\\" : \\"Disconnected\\"}\\n            >\\n                <div\\n                    class={`w-3 h-3 rounded-full ${isConnected ? \\"bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)] connected\\" : \\"bg-red-500 animate-pulse\\"}`}\\n                ></div>\\n                <span class={isConnected ? \\"text-gray-400\\" : \\"text-red-400\\"}>\\n                    {isConnected ? \\"Connected\\" : \\"Reconnecting...\\"}\\n                </span>\\n            </div>\\n\\n            {#if onboardingStep === \\"game\\"}\\n                <button\\n                    on:click={openCharacterSheet}\\n                    class=\\"bg-blue-600 hover:bg-blue-500 text-white px-3 py-1 rounded text-sm transition-colors\\"\\n                >\\n                    Character\\n                </button>\\n                <button\\n                    on:click={() => (showWorldMap = true)}\\n                    class=\\"bg-purple-600 hover:bg-purple-500 text-white px-3 py-1 rounded text-sm transition-colors\\"\\n                >\\n                    World Map\\n                </button>\\n            {/if}\\n            <button\\n                on:click={() => {\\n                    gameAPI.logout();\\n                    goto(\\"/\\");\\n                }}\\n                class=\\"text-sm text-gray-400 hover:text-white transition-colors\\"\\n            >\\n                Logout\\n            </button>\\n            <!-- Mode Indicator Badge -->\\n            <span\\n                class=\\"px-2 py-1 text-xs font-medium rounded-full\\n                {$interfaceMode === \'TEXT\'\\n                    ? \'bg-green-900/50 text-green-400 border border-green-700\'\\n                    : \'bg-purple-900/50 text-purple-400 border border-purple-700\'}\\"\\n            >\\n                {$interfaceMode === \\"TEXT\\" ? \\"📜 MUD\\" : \\"🌍 3D\\"}\\n            </span>\\n        </div>\\n    </div>\\n\\n    <!-- Mode Toggle Button -->\\n    <ModeToggle slot=\\"mode-toggle\\" compact={true} />\\n\\n    <!-- Main Display (MUD mode text output) -->\\n    <div slot=\\"main-display\\" class=\\"p-4 space-y-2\\">\\n        {#if onboardingStep === \\"checking\\"}\\n            <div class=\\"flex-1 flex items-center justify-center\\">\\n                <div class=\\"text-xl text-gray-400 animate-pulse\\">\\n                    Loading...\\n                </div>\\n            </div>\\n        {:else if onboardingStep === \\"interview\\"}\\n            <div\\n                class=\\"flex-1 flex flex-col p-4 overflow-y-auto\\"\\n                id=\\"game-output\\"\\n                data-testid=\\"game-output\\"\\n            >\\n                {#each conversationHistory as msg}\\n                    <div\\n                        class={`mb-4 ${msg.role === \\"user\\" ? \\"text-right\\" : \\"text-left\\"}`}\\n                    >\\n                        <div\\n                            class={`inline-block p-3 rounded-lg max-w-[80%] ${\\n                                msg.role === \\"user\\"\\n                                    ? \\"bg-blue-900/50 text-blue-100\\"\\n                                    : \\"bg-gray-800/50 text-gray-100\\"\\n                            }`}\\n                        >\\n                            {msg.text}\\n                        </div>\\n                    </div>\\n                {/each}\\n                {#if currentQuestion}\\n                    <div class=\\"mb-4 text-left\\">\\n                        <div\\n                            class=\\"inline-block p-3 rounded-lg max-w-[80%] bg-gray-800/50 text-gray-100 border-l-4 border-blue-500\\"\\n                        >\\n                            {currentQuestion}\\n                        </div>\\n                    </div>\\n                {/if}\\n            </div>\\n        {:else}\\n            <!-- Game Output -->\\n            <div\\n                class=\\"flex-1 overflow-y-auto p-4 space-y-2 scroll-smooth\\"\\n                id=\\"game-output\\"\\n                data-testid=\\"game-output\\"\\n            >\\n                {#each messages as msg (msg.id)}\\n                    <div\\n                        class={`message leading-relaxed ${\\n                            msg.type === \\"error\\"\\n                                ? \\"text-red-400\\"\\n                                : msg.type === \\"system\\"\\n                                  ? \\"text-yellow-400\\"\\n                                  : msg.type === \\"player\\"\\n                                    ? \\"text-blue-300\\"\\n                                    : msg.type === \\"emote\\"\\n                                      ? \\"text-orange-300 italic\\"\\n                                      : \\"text-gray-300\\"\\n                        }`}\\n                    >\\n                        {@html msg.text.replace(/\\\\n/g, \\"<br>\\")}\\n                    </div>\\n                {/each}\\n            </div>\\n        {/if}\\n    </div>\\n\\n    <!-- Command Input (footer slot for both modes) -->\\n    <div slot=\\"command-input\\">\\n        {#if onboardingStep !== \\"interview\\" && onboardingStep !== \\"checking\\"}\\n            <QuickButtons on:submit={(e) => handleCommand(e.detail)} />\\n        {/if}\\n\\n        {#if onboardingStep === \\"interview\\"}\\n            <!-- Interview mode: simple input -->\\n            <div class=\\"relative\\">\\n                <input\\n                    type=\\"text\\"\\n                    bind:value={commandInput}\\n                    on:keydown={(e) => e.key === \\"Enter\\" && handleCommand()}\\n                    placeholder=\\"Answer the question...\\"\\n                    class=\\"w-full bg-gray-900 border border-gray-600 rounded-lg px-4 py-3 text-gray-100 text-base focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all\\"\\n                    style=\\"font-size: 16px;\\"\\n                />\\n            </div>\\n        {:else if onboardingStep !== \\"checking\\"}\\n            <!-- Game mode: use thin-client CommandInput -->\\n            <CommandInput on:submit={(e) => handleCommand(e.detail)} />\\n        {/if}\\n    </div>\\n\\n    <!-- Entry Modal -->\\n\\n    {#if showEntryModal && entryOptions}\\n        <WorldEntry options={entryOptions} on:select={handleEntrySelection} />\\n    {/if}\\n\\n    <!-- Character Sheet Modal -->\\n    {#if showCharacterSheet}\\n        <div\\n            class=\\"absolute inset-0 bg-black/80 flex items-center justify-center p-4 z-50 transition-opacity\\"\\n            role=\\"button\\"\\n            tabindex=\\"0\\"\\n            on:click={() => (showCharacterSheet = false)}\\n            on:keydown={(e) =>\\n                e.key === \\"Escape\\" && (showCharacterSheet = false)}\\n        >\\n            <div\\n                class=\\"relative max-w-md w-full\\"\\n                role=\\"document\\"\\n                on:click|stopPropagation\\n                on:keydown|stopPropagation\\n                tabindex=\\"-1\\"\\n            >\\n                <button\\n                    class=\\"absolute -top-2 -right-2 bg-gray-700 rounded-full w-8 h-8 flex items-center justify-center text-white hover:bg-gray-600 z-10\\"\\n                    on:click={() => (showCharacterSheet = false)}\\n                    data-testid=\\"close-character-sheet\\"\\n                >\\n                    ✕\\n                </button>\\n                <CharacterSheet\\n                    characterName={characterStats.name}\\n                    level={characterStats.level}\\n                    experience={characterStats.experience}\\n                    nextLevelXP={characterStats.nextLevelXP}\\n                    strength={characterStats.attributes.strength}\\n                    dexterity={characterStats.attributes.dexterity}\\n                    constitution={characterStats.attributes.constitution}\\n                    intelligence={characterStats.attributes.intelligence}\\n                    wisdom={characterStats.attributes.wisdom}\\n                    charisma={characterStats.attributes.charisma}\\n                    skills={characterSkills}\\n                />\\n            </div>\\n        </div>\\n    {/if}\\n\\n    <!-- World Map Modal -->\\n    <WorldMapModal\\n        isOpen={showWorldMap}\\n        onClose={() => (showWorldMap = false)}\\n        {latestSimEvent}\\n    />\\n</GameContainer>\\n\\n<style>\\n    .messages-container::-webkit-scrollbar {\\n        width: 8px;\\n    }\\n\\n    .messages-container::-webkit-scrollbar-track {\\n        background: rgba(0, 0, 0, 0.3);\\n    }\\n\\n    .messages-container::-webkit-scrollbar-thumb {\\n        background: rgba(55, 65, 81, 0.5);\\n        border-radius: 4px;\\n    }\\n\\n    .messages-container::-webkit-scrollbar-thumb:hover {\\n        background: rgba(75, 85, 99, 0.7);\\n    }\\n</style>\\n"],"names":[],"mappings":"AAuwBI,iCAAmB,mBAAoB,CACnC,KAAK,CAAE,GACX,CAEA,iCAAmB,yBAA0B,CACzC,UAAU,CAAE,KAAK,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,GAAG,CACjC,CAEA,iCAAmB,yBAA0B,CACzC,UAAU,CAAE,KAAK,EAAE,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,CAAC,GAAG,CAAC,CACjC,aAAa,CAAE,GACnB,CAEA,iCAAmB,yBAAyB,MAAO,CAC/C,UAAU,CAAE,KAAK,EAAE,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,CAAC,GAAG,CACpC"}'
};
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $interfaceMode, $$unsubscribe_interfaceMode;
  $$unsubscribe_interfaceMode = subscribe(interfaceMode, (value) => $interfaceMode = value);
  let showWorldMap = false;
  let latestSimEvent = null;
  onDestroy(() => {
  });
  $$result.css.add(css);
  $$unsubscribe_interfaceMode();
  return `${validate_component(GameContainer, "GameContainer").$$render($$result, {}, {}, {
    "command-input": () => {
      return `<div slot="command-input">${``} ${`${``}`}</div>`;
    },
    "main-display": () => {
      return `<div slot="main-display" class="p-4 space-y-2">${`<div class="flex-1 flex items-center justify-center" data-svelte-h="svelte-1nxqtd8"><div class="text-xl text-gray-400 animate-pulse">Loading...</div></div>`}</div>`;
    },
    "mode-toggle": () => {
      return `${validate_component(ModeToggle, "ModeToggle").$$render($$result, { slot: "mode-toggle", compact: true }, {}, {})}`;
    },
    "status-bar": () => {
      return `<div slot="status-bar" class="flex items-center justify-between w-full"><h1 class="text-xl font-bold text-blue-400" data-svelte-h="svelte-1p9gpg2">Thousand Worlds</h1> <div class="flex items-center gap-4"> <div class="flex items-center gap-2 text-sm"${add_attribute("title", "Disconnected", 0)}><div class="${escape(
        null_to_empty(`w-3 h-3 rounded-full ${"bg-red-500 animate-pulse"}`),
        true
      ) + " svelte-lcmnji"}"></div> <span${add_attribute("class", "text-red-400", 0)}>${escape("Reconnecting...")}</span></div> ${``} <button class="text-sm text-gray-400 hover:text-white transition-colors" data-svelte-h="svelte-fhceqg">Logout</button>  <span class="${"px-2 py-1 text-xs font-medium rounded-full " + escape(
        $interfaceMode === "TEXT" ? "bg-green-900/50 text-green-400 border border-green-700" : "bg-purple-900/50 text-purple-400 border border-purple-700",
        true
      )}">${escape($interfaceMode === "TEXT" ? "📜 MUD" : "🌍 3D")}</span></div></div>`;
    },
    default: () => {
      return `     ${``}  ${``}  ${validate_component(WorldMapModal, "WorldMapModal").$$render(
        $$result,
        {
          isOpen: showWorldMap,
          onClose: () => showWorldMap = false,
          latestSimEvent
        },
        {},
        {}
      )}`;
    }
  })}`;
});
export {
  Page as default
};
