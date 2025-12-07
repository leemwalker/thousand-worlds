<script lang="ts">
    import { gameWebSocket } from "$lib/services/websocket";
    import { haptic } from "$lib/stores/haptic";

    let inputText = "";
    let commandHistory: string[] = [];
    let historyIndex = -1;

    function sendCommand() {
        if (!inputText.trim()) return;

        // Trigger haptic feedback
        haptic.light();

        // Send raw text to backend - NO PARSING
        gameWebSocket.sendRawCommand(inputText.trim());

        // Track history for up/down navigation
        commandHistory.unshift(inputText);
        if (commandHistory.length > 50) commandHistory.pop();

        inputText = "";
        historyIndex = -1;
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            sendCommand();
        } else if (e.key === "ArrowUp") {
            e.preventDefault();
            if (historyIndex < commandHistory.length - 1) {
                historyIndex++;
                inputText = commandHistory[historyIndex];
            }
        } else if (e.key === "ArrowDown") {
            e.preventDefault();
            if (historyIndex > 0) {
                historyIndex--;
                inputText = commandHistory[historyIndex];
            } else if (historyIndex === 0) {
                historyIndex = -1;
                inputText = "";
            }
        } else if (e.key === "Escape") {
            inputText = "";
            historyIndex = -1;
        }
    }
</script>

<div class="flex gap-2 w-full">
    <input
        type="text"
        bind:value={inputText}
        on:keydown={handleKeydown}
        placeholder="Enter command..."
        class="flex-1 px-4 py-3 bg-gray-900 border border-gray-700 rounded-lg text-gray-100 text-base placeholder-gray-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
        style="font-size: 16px;"
        autocomplete="off"
        autocorrect="off"
        autocapitalize="off"
        spellcheck="false"
    />

    <button
        on:click={sendCommand}
        disabled={!inputText.trim()}
        class="min-w-[44px] min-h-[44px] px-6 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:text-gray-500 text-white font-semibold rounded-lg transition-colors flex items-center justify-center"
        aria-label="Send command"
    >
        Send
    </button>
</div>

<style>
    /* Prevent iOS zoom on input focus */
    input {
        font-size: 16px;
        -webkit-appearance: none;
    }
</style>
