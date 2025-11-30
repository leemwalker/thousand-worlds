<script lang="ts">
    import type { BehavioralBaseline } from "$lib/types/ui";

    export let baseline: BehavioralBaseline = {
        openness: 0.5,
        conscientiousness: 0.5,
        extraversion: 0.5,
        agreeableness: 0.5,
        neuroticism: 0.5,
    };

    export let currentBehavior: BehavioralBaseline = {
        openness: 0.5,
        conscientiousness: 0.5,
        extraversion: 0.5,
        agreeableness: 0.5,
        neuroticism: 0.5,
    };

    export let driftLevel: "none" | "subtle" | "moderate" | "severe" = "none";
    export let relationshipImpacts: Array<{
        name: string;
        affinityDelta: number;
    }> = [];
    export let recentComments: Array<{ speaker: string; comment: string }> = [];

    function calculateDrift(trait: keyof BehavioralBaseline): number {
        return Math.abs(currentBehavior[trait] - baseline[trait]);
    }

    function getDriftColor(drift: number): string {
        if (drift < 0.3) return "text-green-500";
        if (drift < 0.5) return "text-yellow-500";
        if (drift < 0.7) return "text-orange-500";
        return "text-red-500";
    }

    function getDriftIcon(drift: number): string {
        if (drift < 0.3) return "✓";
        if (drift < 0.5) return "⚠";
        if (drift < 0.7) return "⚠⚠";
        return "⚠⚠⚠";
    }

    function getDriftBarColor(drift: number): string {
        if (drift < 0.3) return "bg-green-500";
        if (drift < 0.5) return "bg-yellow-500";
        if (drift < 0.7) return "bg-orange-500";
        return "bg-red-500";
    }

    function getDriftBorderColor(): string {
        if (driftLevel === "severe") return "border-red-500";
        if (driftLevel === "moderate") return "border-orange-500";
        if (driftLevel === "subtle") return "border-yellow-500";
        return "border-gray-700";
    }

    function getDriftLevelColor(): string {
        if (driftLevel === "severe") return "text-red-500";
        if (driftLevel === "moderate") return "text-orange-500";
        return "text-yellow-500";
    }
</script>

<div
    class="drift-monitor bg-gray-900 rounded-lg p-4 border-2 {getDriftBorderColor()}"
>
    <h3 class="text-lg font-bold text-gray-100 mb-3">
        Personality Monitor
        {#if driftLevel !== "none"}
            <span class="text-sm {getDriftLevelColor()}">
                ({driftLevel} drift)
            </span>
        {/if}
    </h3>

    <!-- Behavioral Traits -->
    <div class="traits-grid grid grid-cols-2 gap-3 mb-4">
        {#each Object.entries(baseline) as [trait, baseValue]}
            {@const currentValue = currentBehavior[trait as keyof BehavioralBaseline]}
            {@const drift = calculateDrift(trait as keyof BehavioralBaseline)}

            <div class="trait-item">
                <div class="flex justify-between items-center mb-1">
                    <span class="text-sm text-gray-300 capitalize">{trait}</span
                    >
                    <span class="{getDriftColor(drift)} text-sm font-bold">
                        {getDriftIcon(drift)}
                    </span>
                </div>

                <!-- Baseline bar -->
                <div class="relative h-4 bg-gray-700 rounded">
                    <!-- Original baseline -->
                    <div
                        class="absolute top-0 left-0 h-full bg-gray-500 rounded opacity-50"
                        style="width: {baseValue * 100}%"
                    />

                    <!-- Current value -->
                    <div
                        class="absolute top-0 left-0 h-full rounded transition-all {getDriftBarColor(
                            drift,
                        )}"
                        style="width: {currentValue * 100}%"
                    />
                </div>

                <div class="flex justify-between text-xs text-gray-400 mt-1">
                    <span>Base: {(baseValue * 100).toFixed(0)}%</span>
                    <span class={getDriftColor(drift)}>
                        Now: {(currentValue * 100).toFixed(0)}%
                        {#if drift > 0.3}
                            (Δ{(drift * 100).toFixed(0)}%)
                        {/if}
                    </span>
                </div>
            </div>
        {/each}
    </div>

    <!-- Relationship Impacts -->
    {#if relationshipImpacts.length > 0}
        <div class="relationship-impacts mb-4">
            <h4 class="text-sm font-bold text-gray-300 mb-2">
                Relationship Changes
            </h4>
            <div class="space-y-1">
                {#each relationshipImpacts as impact}
                    <div class="flex justify-between text-sm">
                        <span class="text-gray-400">{impact.name}</span>
                        <span
                            class={impact.affinityDelta < 0
                                ? "text-red-400"
                                : "text-green-400"}
                        >
                            {impact.affinityDelta > 0
                                ? "+"
                                : ""}{impact.affinityDelta}
                        </span>
                    </div>
                {/each}
            </div>
        </div>
    {/if}

    <!-- Recent NPC Comments -->
    {#if recentComments.length > 0}
        <div class="recent-comments">
            <h4 class="text-sm font-bold text-gray-300 mb-2">
                What Others Say
            </h4>
            <div class="space-y-2 max-h-32 overflow-y-auto">
                {#each recentComments as comment}
                    <div class="bg-gray-800 rounded p-2">
                        <div class="text-xs text-cyan-400 font-bold">
                            {comment.speaker}:
                        </div>
                        <div class="text-xs text-gray-300 italic">
                            "{comment.comment}"
                        </div>
                    </div>
                {/each}
            </div>
        </div>
    {/if}
</div>
