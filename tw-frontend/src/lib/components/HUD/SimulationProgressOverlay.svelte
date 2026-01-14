<script lang="ts">
    import { fade } from "svelte/transition";

    export let yearsElapsed: number = 0;
    export let totalYears: number = 1000000000;
    export let isSimulating: boolean = false;

    let displayMessage: string = "";
    let showOverlay: boolean = false;
    let fadeTimeout: ReturnType<typeof setTimeout>;

    // Format years for display
    function formatYears(years: number): string {
        if (years >= 1e9) return `${(years / 1e9).toFixed(2)} billion`;
        if (years >= 1e6) return `${(years / 1e6).toFixed(1)} million`;
        if (years >= 1e3) return `${(years / 1e3).toFixed(0)} thousand`;
        return `${years}`;
    }

    function getEraName(years: number): string {
        if (years < 500_000_000) return "Hadean Eon";
        if (years < 2_500_000_000) return "Archean Eon";
        if (years < 4_000_000_000) return "Proterozoic Eon";
        return "Phanerozoic Eon";
    }

    // Trigger overlay display when yearsElapsed changes significantly
    $: if (isSimulating && yearsElapsed > 0) {
        const percentage =
            totalYears > 0 ? (yearsElapsed / totalYears) * 100 : 0;
        const era = getEraName(yearsElapsed);
        displayMessage = `${formatYears(yearsElapsed)} years — ${era} (${percentage.toFixed(0)}%)`;

        showOverlay = true;
        clearTimeout(fadeTimeout);
        fadeTimeout = setTimeout(() => {
            showOverlay = false;
        }, 10000); // Fade after 10 seconds
    }

    // Reset when simulation ends
    $: if (!isSimulating) {
        showOverlay = false;
        clearTimeout(fadeTimeout);
    }
</script>

{#if showOverlay}
    <div
        class="simulation-progress-overlay"
        transition:fade={{ duration: 500 }}
    >
        <div class="progress-text">
            {displayMessage}
        </div>
    </div>
{/if}

<style>
    .simulation-progress-overlay {
        position: fixed;
        bottom: 20%;
        left: 50%;
        transform: translateX(-50%);
        z-index: 100;
        pointer-events: none;
    }

    .progress-text {
        background: rgba(0, 0, 0, 0.7);
        backdrop-filter: blur(4px);
        border: 1px solid rgba(100, 150, 255, 0.3);
        border-radius: 8px;
        padding: 12px 24px;
        color: #e0e8ff;
        font-family: "Roboto Mono", monospace;
        font-size: 1.1rem;
        text-shadow: 0 0 10px rgba(100, 150, 255, 0.5);
        white-space: nowrap;
    }
</style>
