<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { haptic } from "$lib/stores/haptic";

    const dispatch = createEventDispatcher();

    function move(direction: string) {
        haptic.selection();
        dispatch("command", direction);
    }
</script>

<div
    class="grid grid-cols-3 gap-1 w-fit p-2 bg-gray-800 rounded-xl border border-gray-700 shadow-lg"
    role="group"
    aria-label="Movement Controls"
>
    <!-- Top Row -->
    <button
        on:click={() => move("nw")}
        class="w-12 h-12 flex items-center justify-center bg-gray-700/50 hover:bg-gray-600 active:bg-blue-600 rounded-tl-lg transition-colors"
        aria-label="Northwest"
    >
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            style="transform: rotate(-45deg)"><path d="m18 15-6-6-6 6" /></svg
        >
    </button>
    <button
        on:click={() => move("north")}
        class="w-12 h-12 flex items-center justify-center bg-gray-700/50 hover:bg-gray-600 active:bg-blue-600 transition-colors"
        aria-label="North"
    >
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"><path d="m18 15-6-6-6 6" /></svg
        >
    </button>
    <button
        on:click={() => move("ne")}
        class="w-12 h-12 flex items-center justify-center bg-gray-700/50 hover:bg-gray-600 active:bg-blue-600 rounded-tr-lg transition-colors"
        aria-label="Northeast"
    >
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            style="transform: rotate(45deg)"><path d="m18 15-6-6-6 6" /></svg
        >
    </button>

    <!-- Middle Row -->
    <button
        on:click={() => move("west")}
        class="w-12 h-12 flex items-center justify-center bg-gray-700/50 hover:bg-gray-600 active:bg-blue-600 transition-colors"
        aria-label="West"
    >
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            style="transform: rotate(-90deg)"><path d="m18 15-6-6-6 6" /></svg
        >
    </button>
    <button
        on:click={() => move("look")}
        class="w-12 h-12 flex items-center justify-center bg-gray-700 hover:bg-gray-600 active:bg-blue-600 font-bold text-xs text-gray-300 transition-colors"
        aria-label="Look"
    >
        LOOK
    </button>
    <button
        on:click={() => move("east")}
        class="w-12 h-12 flex items-center justify-center bg-gray-700/50 hover:bg-gray-600 active:bg-blue-600 transition-colors"
        aria-label="East"
    >
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            style="transform: rotate(90deg)"><path d="m18 15-6-6-6 6" /></svg
        >
    </button>

    <!-- Bottom Row -->
    <button
        on:click={() => move("sw")}
        class="w-12 h-12 flex items-center justify-center bg-gray-700/50 hover:bg-gray-600 active:bg-blue-600 rounded-bl-lg transition-colors"
        aria-label="Southwest"
    >
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            style="transform: rotate(-135deg)"><path d="m18 15-6-6-6 6" /></svg
        >
    </button>
    <button
        on:click={() => move("south")}
        class="w-12 h-12 flex items-center justify-center bg-gray-700/50 hover:bg-gray-600 active:bg-blue-600 transition-colors"
        aria-label="South"
    >
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            style="transform: rotate(180deg)"><path d="m18 15-6-6-6 6" /></svg
        >
    </button>
    <button
        on:click={() => move("se")}
        class="w-12 h-12 flex items-center justify-center bg-gray-700/50 hover:bg-gray-600 active:bg-blue-600 rounded-br-lg transition-colors"
        aria-label="Southeast"
    >
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            style="transform: rotate(135deg)"><path d="m18 15-6-6-6 6" /></svg
        >
    </button>
</div>
