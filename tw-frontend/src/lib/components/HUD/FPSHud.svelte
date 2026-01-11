<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { gameOutput } from "$lib/stores/ui";
    import type { GameMessage } from "$lib/types/ui";

    const dispatch = createEventDispatcher<{
        openMenu: void;
        sendMessage: { text: string };
    }>();

    // Message input
    let messageInput = "";

    // Get last few messages for display
    $: recentMessages = $gameOutput.slice(-4);

    function handleOpenMenu() {
        dispatch("openMenu");
    }

    function handleSendMessage() {
        if (messageInput.trim()) {
            dispatch("sendMessage", { text: messageInput.trim() });
            messageInput = "";
        }
    }

    function handleKeyDown(event: KeyboardEvent) {
        if (event.key === "Enter" && !event.shiftKey) {
            event.preventDefault();
            handleSendMessage();
        }
    }

    // Get color based on message type
    function getMessageColor(type: string): string {
        switch (type) {
            case "error":
                return "#ff6b6b";
            case "system":
                return "#ffd93d";
            case "player":
                return "#74c0fc";
            case "emote":
                return "#ffa94d";
            default:
                return "#e0e0e0";
        }
    }
</script>

<!-- Minimal FPS HUD -->
<div class="fps-hud">
    <!-- Menu Button (Top Left) -->
    <button class="menu-btn" on:click={handleOpenMenu} aria-label="Open menu">
        <svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24">
            <path d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z" />
        </svg>
    </button>

    <!-- Message Area (Bottom Left) -->
    <div class="message-area">
        <!-- Message Log -->
        <div class="message-log">
            {#each recentMessages as msg (msg.id)}
                <div class="message" style="color: {getMessageColor(msg.type)}">
                    {msg.text}
                </div>
            {/each}
        </div>

        <!-- Message Input -->
        <input
            type="text"
            class="message-input"
            placeholder="Type message..."
            bind:value={messageInput}
            on:keydown={handleKeyDown}
        />
    </div>
</div>

<style>
    .fps-hud {
        position: fixed;
        inset: 0;
        pointer-events: none; /* Allow click-through for FPS camera */
        z-index: 50;
    }

    /* Menu Button - Top Left */
    .menu-btn {
        position: absolute;
        top: 16px;
        left: 16px;
        width: 48px;
        height: 48px;
        border-radius: 50%;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(255, 255, 255, 0.2);
        color: rgba(255, 255, 255, 0.8);
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        pointer-events: auto;
        transition:
            background 0.2s,
            transform 0.1s;
        backdrop-filter: blur(4px);
    }

    .menu-btn:hover {
        background: rgba(0, 0, 0, 0.7);
        color: #fff;
        transform: scale(1.05);
    }

    /* Message Area - Bottom Left */
    .message-area {
        position: absolute;
        bottom: 16px;
        left: 16px;
        width: 320px;
        max-width: calc(100vw - 32px);
        pointer-events: auto;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .message-log {
        display: flex;
        flex-direction: column;
        gap: 4px;
        max-height: 150px;
        overflow-y: auto;
        padding: 8px;
        background: rgba(0, 0, 0, 0.4);
        border-radius: 8px;
        backdrop-filter: blur(4px);
    }

    .message {
        font-family: monospace;
        font-size: 0.85rem;
        line-height: 1.3;
        text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
    }

    .message-input {
        padding: 10px 14px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(255, 255, 255, 0.2);
        border-radius: 8px;
        color: #fff;
        font-family: monospace;
        font-size: 0.9rem;
        outline: none;
        backdrop-filter: blur(4px);
    }

    .message-input::placeholder {
        color: rgba(255, 255, 255, 0.4);
    }

    .message-input:focus {
        border-color: rgba(77, 171, 247, 0.5);
        background: rgba(0, 0, 0, 0.6);
    }
</style>
