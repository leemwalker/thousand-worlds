<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { CommandParser } from "./CommandParser";

    const dispatch = createEventDispatcher();
    const parser = new CommandParser();

    let inputValue = "";
    let inputElement: HTMLInputElement;
    let history: string[] = [];
    let historyIndex = -1;

    function handleSubmit() {
        if (!inputValue.trim()) return;

        // Add to history
        history = [inputValue, ...history].slice(0, 50);
        historyIndex = -1;

        // Parse and dispatch
        const result = parser.parse(inputValue);
        dispatch("command", result);

        // Clear input
        inputValue = "";
    }

    function handleKeydown(event: KeyboardEvent) {
        if (event.key === "Enter") {
            handleSubmit();
        } else if (event.key === "ArrowUp") {
            event.preventDefault();
            if (history.length > 0) {
                historyIndex = Math.min(historyIndex + 1, history.length - 1);
                inputValue = history[historyIndex];
            }
        } else if (event.key === "ArrowDown") {
            event.preventDefault();
            if (historyIndex > 0) {
                historyIndex--;
                inputValue = history[historyIndex];
            } else if (historyIndex === 0) {
                historyIndex = -1;
                inputValue = "";
            }
        }
    }
</script>

<div class="flex w-full h-full gap-2">
    <input
        bind:this={inputElement}
        bind:value={inputValue}
        on:keydown={handleKeydown}
        type="text"
        placeholder="Enter command..."
        class="flex-1 bg-gray-900 border border-gray-600 rounded px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
    />
    <button
        on:click={handleSubmit}
        class="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded font-bold transition-colors"
    >
        Send
    </button>
</div>
