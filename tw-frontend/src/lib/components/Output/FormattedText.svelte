<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import type { TextSegment } from "$lib/types/ui";

    export let segments: TextSegment[] = [];

    const dispatch = createEventDispatcher();

    function handleEntityClick(segment: TextSegment) {
        if (segment.entityID && segment.entityType) {
            dispatch("entityClick", {
                entityID: segment.entityID,
                entityType: segment.entityType,
            });
        }
    }
</script>

<div class="output-line mb-1 leading-relaxed break-words">
    {#each segments as segment}
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <!-- svelte-ignore a11y-no-static-element-interactions -->
        <span
            class="{segment.color || 'text-gray-300'} {segment.bold
                ? 'font-bold'
                : ''} {segment.italic ? 'italic' : ''} {segment.entityID
                ? 'cursor-pointer hover:underline'
                : ''}"
            on:click={() => handleEntityClick(segment)}
        >
            {segment.text}
        </span>
    {/each}
</div>
