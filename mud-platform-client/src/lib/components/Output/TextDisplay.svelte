<script lang="ts">
    import { onMount, onDestroy, tick } from "svelte";
    import { gameOutput } from "$lib/stores/ui";
    import { OutputFormatter } from "./OutputFormatter";
    import FormattedText from "./FormattedText.svelte";

    const formatter = new OutputFormatter();
    let container: HTMLDivElement;
    let resizeObserver: ResizeObserver;

    const scrollToBottom = async () => {
        if (!container) return;
        await tick();
        // Use instant scroll for terminal feel
        container.scrollTo({ top: container.scrollHeight, behavior: "auto" });
        // Direct assignment fallback
        container.scrollTop = container.scrollHeight;
    };

    // Auto-scroll on new messages
    $: if ($gameOutput) {
        scrollToBottom();
    }

    onMount(() => {
        if (container) {
            resizeObserver = new ResizeObserver(() => {
                scrollToBottom();
            });
            resizeObserver.observe(container);
        }
    });

    onDestroy(() => {
        if (resizeObserver) resizeObserver.disconnect();
    });
</script>

<div
    bind:this={container}
    class="w-full h-full overflow-y-auto flex flex-col p-4 font-mono text-sm md:text-base"
>
    {#each $gameOutput as message (message.id)}
        <FormattedText segments={formatter.formatGameOutput(message)} />
    {/each}

    {#if $gameOutput.length === 0}
        <div class="text-gray-600 italic text-center mt-10">
            <!-- Waiting for server message... -->
        </div>
    {/if}
</div>
