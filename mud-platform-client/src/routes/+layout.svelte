<script>
    import "../app.css";
    import { onMount } from "svelte";
    import { checkIOSInstallPrompt } from "$lib/stores/pwa";
    import IOSInstallPrompt from "$lib/components/PWA/IOSInstallPrompt.svelte";

    export let params = {};

    onMount(() => {
        // Check if we should show iOS install prompt
        checkIOSInstallPrompt();
    });
</script>

<IOSInstallPrompt />
<slot />
