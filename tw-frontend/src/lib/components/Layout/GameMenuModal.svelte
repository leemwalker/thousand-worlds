<script lang="ts">
    import { createEventDispatcher } from "svelte";

    // Props
    export let isOpen = false;

    const dispatch = createEventDispatcher<{
        close: void;
        resetWorld: void;
        returnToLobby: void;
        logout: void;
    }>();

    type Tab = "world" | "character" | "account";
    let activeTab: Tab = "world";

    function handleClose() {
        dispatch("close");
    }

    function handleResetWorld() {
        dispatch("resetWorld");
    }

    function handleReturnToLobby() {
        dispatch("returnToLobby");
    }

    function handleLogout() {
        dispatch("logout");
    }
</script>

{#if isOpen}
    <!-- Modal Backdrop -->
    <div
        class="modal-backdrop"
        on:click={handleClose}
        on:keydown={(e) => e.key === "Escape" && handleClose()}
        role="button"
        tabindex="0"
    >
        <!-- Modal Content -->
        <div
            class="modal-content"
            on:click|stopPropagation
            role="dialog"
            aria-modal="true"
            aria-labelledby="menu-title"
        >
            <header class="modal-header">
                <h2 id="menu-title">Menu</h2>
                <button
                    class="close-btn"
                    on:click={handleClose}
                    aria-label="Close menu">×</button
                >
            </header>

            <!-- Tab Navigation -->
            <nav class="tab-nav">
                <button
                    class="tab-btn"
                    class:active={activeTab === "world"}
                    on:click={() => (activeTab = "world")}
                >
                    World
                </button>
                <button
                    class="tab-btn"
                    class:active={activeTab === "character"}
                    on:click={() => (activeTab = "character")}
                >
                    Character
                </button>
                <button
                    class="tab-btn"
                    class:active={activeTab === "account"}
                    on:click={() => (activeTab = "account")}
                >
                    Account
                </button>
            </nav>

            <!-- Tab Content -->
            <div class="tab-content">
                {#if activeTab === "world"}
                    <div class="tab-panel">
                        <button
                            class="action-btn warning"
                            on:click={handleResetWorld}
                        >
                            Reset World
                        </button>
                        <button
                            class="action-btn"
                            on:click={handleReturnToLobby}
                        >
                            Return to Lobby
                        </button>
                    </div>
                {:else if activeTab === "character"}
                    <div class="tab-panel">
                        <p class="placeholder-text">
                            Character options coming soon...
                        </p>
                        <!-- Future: Character customization, stats, etc. -->
                    </div>
                {:else if activeTab === "account"}
                    <div class="tab-panel">
                        <button
                            class="action-btn danger"
                            on:click={handleLogout}
                        >
                            Logout
                        </button>
                    </div>
                {/if}
            </div>
        </div>
    </div>
{/if}

<style>
    .modal-backdrop {
        position: fixed;
        inset: 0;
        background: rgba(0, 0, 0, 0.7);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
        backdrop-filter: blur(4px);
    }

    .modal-content {
        background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 12px;
        width: 90%;
        max-width: 400px;
        box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
    }

    .modal-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 16px 20px;
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    }

    .modal-header h2 {
        margin: 0;
        color: #fff;
        font-size: 1.25rem;
        font-weight: 600;
    }

    .close-btn {
        background: none;
        border: none;
        color: rgba(255, 255, 255, 0.6);
        font-size: 1.5rem;
        cursor: pointer;
        padding: 0;
        width: 32px;
        height: 32px;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: 50%;
        transition:
            background 0.2s,
            color 0.2s;
    }

    .close-btn:hover {
        background: rgba(255, 255, 255, 0.1);
        color: #fff;
    }

    .tab-nav {
        display: flex;
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    }

    .tab-btn {
        flex: 1;
        background: none;
        border: none;
        color: rgba(255, 255, 255, 0.6);
        padding: 12px 16px;
        cursor: pointer;
        font-size: 0.9rem;
        transition:
            color 0.2s,
            background 0.2s;
    }

    .tab-btn:hover {
        color: #fff;
        background: rgba(255, 255, 255, 0.05);
    }

    .tab-btn.active {
        color: #4dabf7;
        border-bottom: 2px solid #4dabf7;
        margin-bottom: -1px;
    }

    .tab-content {
        padding: 20px;
    }

    .tab-panel {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .action-btn {
        background: rgba(255, 255, 255, 0.1);
        border: 1px solid rgba(255, 255, 255, 0.2);
        color: #fff;
        padding: 12px 20px;
        border-radius: 8px;
        font-size: 1rem;
        cursor: pointer;
        transition:
            background 0.2s,
            transform 0.1s;
    }

    .action-btn:hover {
        background: rgba(255, 255, 255, 0.15);
        transform: translateY(-1px);
    }

    .action-btn.warning {
        border-color: rgba(255, 193, 7, 0.4);
        color: #ffc107;
    }

    .action-btn.warning:hover {
        background: rgba(255, 193, 7, 0.15);
    }

    .action-btn.danger {
        border-color: rgba(220, 53, 69, 0.4);
        color: #dc3545;
    }

    .action-btn.danger:hover {
        background: rgba(220, 53, 69, 0.15);
    }

    .placeholder-text {
        color: rgba(255, 255, 255, 0.5);
        text-align: center;
        padding: 20px;
    }
</style>
