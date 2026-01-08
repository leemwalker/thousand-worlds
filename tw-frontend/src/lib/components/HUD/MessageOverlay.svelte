<script lang="ts">
    import { onDestroy } from "svelte";
    import { gameOutput } from "$lib/stores/ui";
    import type { GameMessage } from "$lib/types/ui";

    // Number of messages to show
    const MAX_VISIBLE = 4;

    // Track which messages have been shown (for fade animation)
    interface DisplayMessage extends GameMessage {
        displayId: string;
        opacity: number;
        age: number; // seconds since display
    }

    let displayMessages: DisplayMessage[] = [];
    let updateInterval: ReturnType<typeof setInterval> | null = null;

    // Subscribe to game output
    const unsubscribe = gameOutput.subscribe((messages) => {
        // Get the latest messages
        const recent = messages.slice(-MAX_VISIBLE);

        // Create display messages with opacity based on position
        displayMessages = recent.map((msg, i) => ({
            ...msg,
            displayId: msg.id,
            opacity: 0.3 + ((i + 1) / MAX_VISIBLE) * 0.6, // 0.3 to 0.9
            age: 0,
        }));
    });

    // Age messages and fade them out
    function startAging(): void {
        if (updateInterval) return;

        updateInterval = setInterval(() => {
            displayMessages = displayMessages
                .map((msg) => ({
                    ...msg,
                    age: msg.age + 0.1,
                    opacity: Math.max(0, msg.opacity - 0.02 * msg.age),
                }))
                .filter((msg) => msg.opacity > 0);
        }, 100);
    }

    // Start aging on mount
    startAging();

    onDestroy(() => {
        unsubscribe();
        if (updateInterval) {
            clearInterval(updateInterval);
        }
    });

    // Get color class based on message type
    function getTypeClass(type: string): string {
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
</script>

/** * MessageOverlay.svelte * Semi-transparent message display at screen bottom.
* Messages fade in and out without blocking FPS camera controls. */
<div class="message-overlay">
    {#each displayMessages as msg (msg.displayId)}
        <div
            class="message {getTypeClass(msg.type)}"
            style="opacity: {msg.opacity}; transform: translateY({-msg.age *
                5}px);"
        >
            {msg.text}
        </div>
    {/each}
</div>

<style>
    .message-overlay {
        position: fixed;
        bottom: 80px; /* Above command input */
        left: 20px;
        right: 20px;
        pointer-events: none; /* Allow click-through for FPS camera */
        z-index: 100;
        display: flex;
        flex-direction: column;
        gap: 4px;
        max-height: 200px;
        overflow: hidden;
    }

    .message {
        background: rgba(0, 0, 0, 0.6);
        backdrop-filter: blur(4px);
        padding: 8px 16px;
        border-radius: 4px;
        font-family: monospace;
        font-size: 0.9rem;
        line-height: 1.4;
        transition:
            opacity 0.3s ease,
            transform 0.3s ease;
        text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
    }
</style>
