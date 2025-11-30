<script lang="ts">
    import { afterUpdate } from "svelte";
    import { gameOutput } from "$lib/stores/ui";
    import { OutputFormatter } from "./OutputFormatter";
    import FormattedText from "./FormattedText.svelte";

    const formatter = new OutputFormatter();
    let container: HTMLDivElement;

    // Auto-scroll to bottom
    afterUpdate(() => {
        if (container) {
            container.scrollTop = container.scrollHeight;
        }
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
            Welcome to Thousand Worlds. Type a command to begin.
        </div>
    {/if}
</div>
