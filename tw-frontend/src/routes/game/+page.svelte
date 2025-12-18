<script lang="ts">
    import { onMount, onDestroy, tick } from "svelte";
    import { goto } from "$app/navigation";
    import { gameAPI, type User } from "$lib/services/api";
    import { gameWebSocket, type ServerMessage } from "$lib/services/websocket";
    import WorldEntry from "$lib/components/WorldEntry.svelte";
    import CharacterSheet from "$lib/components/Character/CharacterSheet.svelte";
    import type { Skill } from "$lib/types/game";
    import CommandInput from "$lib/components/Input/CommandInput.svelte";
    import QuickButtons from "$lib/components/Input/QuickButtons.svelte";

    // Onboarding state
    let onboardingStep:
        | "checking"
        | "interview"
        | "character"
        | "game"
        | "lobby" = "checking";
    let interviewSessionId: string | null = null;
    let currentQuestion: string = "";
    let userResponse: string = "";
    let conversationHistory: Array<{ role: string; text: string }> = [];

    // Entry state
    let showEntryModal = false;
    let entryOptions: any = null;
    let targetWorldId: string | null = null;

    // Game state
    let messages: Array<{
        id: string;
        type: string;
        text: string;
        timestamp: Date;
    }> = [];
    let commandInput: string = "";
    let unsubscribe: (() => void) | null = null;

    // Connection status
    let isConnected = false;

    // Character Sheet state
    let currentCharacterId: string | null = null;
    let showCharacterSheet = false;
    let characterSkills: Record<string, Skill> = {};
    let characterStats = {
        name: "Adventurer",
        level: 1,
        experience: 0,
        nextLevelXP: 100,
        attributes: {
            strength: 10,
            dexterity: 10,
            constitution: 10,
            intelligence: 10,
            wisdom: 10,
            charisma: 10,
        },
    };

    let currentUser: User | null = null;
    // Use relative URL to go through Vite proxy with extended timeout
    const API_URL = "/api";

    onMount(async () => {
        // Check if user is authenticated via cookie
        try {
            currentUser = await gameAPI.getMe();
        } catch (err) {
            // Not authenticated, redirect to login
            goto("/");
            return;
        }

        // Subscribe to connection status
        const unsubscribeConnection = gameWebSocket.connected.subscribe(
            (value) => {
                isConnected = value;
            },
        );

        // Don't connect WebSocket yet - will connect after checking character status
        unsubscribe = gameWebSocket.onMessage(handleServerMessage);

        // Store both unsubscribe functions
        const originalUnsubscribe = unsubscribe;
        unsubscribe = () => {
            if (originalUnsubscribe) originalUnsubscribe();
            unsubscribeConnection();
        };

        // Check onboarding status
        await checkOnboardingStatus();
    });

    onDestroy(() => {
        if (unsubscribe) unsubscribe();
    });

    async function checkOnboardingStatus() {
        try {
            // Check for active interview
            const interviewRes = await fetch(
                `${API_URL}/world/interview/active`,
                {
                    credentials: "include", // Send cookies
                },
            );

            if (interviewRes.ok) {
                const interview = await interviewRes.json();
                if (interview && interview.status === "in_progress") {
                    // Resume interview
                    onboardingStep = "interview";
                    interviewSessionId = interview.session_id;

                    // Check if interview is already complete
                    if (
                        interview.question ===
                            "The interview is already complete." ||
                        interview.question ===
                            "This interview is already complete."
                    ) {
                        addMessage(
                            "system",
                            "World interview previously completed.",
                        );
                        setTimeout(() => {
                            // Go to lobby instead of character creation directly
                            joinLobby();
                        }, 1000);
                        return;
                    }

                    conversationHistory = interview.conversation || [];
                    if (conversationHistory.length > 0) {
                        const lastMessage =
                            conversationHistory[conversationHistory.length - 1];
                        if (lastMessage.role === "assistant") {
                            currentQuestion = lastMessage.text;
                        }
                    } else if (interview.question) {
                        currentQuestion = interview.question;
                        addMessage("interview", currentQuestion);
                    }

                    addMessage(
                        "system",
                        "Resuming your world creation interview...",
                    );
                    return;
                }
            }

            // check for auto-resume
            if (currentUser?.last_world_id) {
                try {
                    const data = await gameAPI.getCharacters();
                    // Find character for this world
                    // Assuming getCharacters returns list of characters
                    // We need to type check roughly or assume any[]
                    if (data && data.characters) {
                        const char = data.characters.find(
                            (c) => c.world_id === currentUser?.last_world_id,
                        );
                        if (char) {
                            await joinGame(char.character_id);
                            return;
                        }
                    }
                } catch (e) {
                    console.error("Auto-resume check failed", e);
                }
            }

            // Always join Lobby first
            await joinLobby();
        } catch (error) {
            console.error("Onboarding check failed:", error);
            onboardingStep = "game";
            addMessage(
                "error",
                "Failed to check status. Starting in game mode.",
            );
            // Attempt to join lobby (connect WS) even if API check failed
            joinLobby();
        }
    }

    async function joinLobby() {
        console.log("[Lobby] Joining lobby...");
        try {
            // Join Lobby (backend handles Ghost creation or existing character)
            // Lobby ID is 00000000-0000-0000-0000-000000000000, but we just connect without character_id to auto-join lobby
            // Wait, handler logic: if character_id is nil, join lobby.

            console.log("[Lobby] Disconnecting existing WebSocket...");
            gameWebSocket.disconnect();
            await new Promise((resolve) => setTimeout(resolve, 100));

            console.log("[Lobby] Connecting to WebSocket...");
            gameWebSocket.connect(); // No token or character ID needed - cookie sent automatically

            onboardingStep = "lobby";
            console.log("[Lobby] Lobby join complete");
        } catch (error) {
            console.error("[Lobby] Failed to join lobby:", error);
            addMessage("error", "Failed to join lobby.");
        }
    }

    async function startWorldInterview() {
        try {
            // addMessage(
            //     "system",
            //     "Welcome to Thousand Worlds! Let's create your custom world...",
            // );

            const response = await fetch(`${API_URL}/world/interview/start`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                credentials: "include", // Send cookies
            });

            if (!response.ok) {
                const errorText = await response.text();
                console.error(
                    "Interview start failed:",
                    response.status,
                    errorText,
                );
                throw new Error(
                    `Failed to start interview: ${response.status}`,
                );
            }

            const data = await response.json();
            console.log("Interview started:", data);

            if (!data.session_id || !data.question) {
                console.error("Invalid interview response:", data);
                throw new Error("Invalid response from server");
            }

            interviewSessionId = data.session_id;
            currentQuestion =
                data.question ||
                "Tell me about the world you'd like to create.";
            conversationHistory.push({
                role: "assistant",
                text: currentQuestion,
            });
            addMessage("interview", currentQuestion);
            onboardingStep = "interview";
        } catch (error: any) {
            console.error("Failed to start interview:", error);
            addMessage(
                "error",
                error.message ||
                    "Failed to start world interview. Please try again.",
            );
        }
    }

    async function sendInterviewResponse() {
        if (!userResponse.trim() || !interviewSessionId) return;

        const userMessage = userResponse.trim();
        conversationHistory.push({ role: "user", text: userMessage });
        addMessage("user", userMessage);
        userResponse = "";

        try {
            addMessage("system", "Thinking... (this may take 10-20 seconds)");

            // Create AbortController for manual timeout control (60 seconds for LLM)
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 60000);

            const response = await fetch(`${API_URL}/world/interview/message`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                credentials: "include", // Send cookies
                body: JSON.stringify({
                    session_id: interviewSessionId,
                    message: userMessage,
                }),
                signal: controller.signal,
            });

            clearTimeout(timeoutId);

            if (!response.ok) {
                const errorText = await response.text();
                console.error(
                    "Interview message failed:",
                    response.status,
                    errorText,
                );
                throw new Error(`Failed to send message: ${response.status}`);
            }

            const data = await response.json();
            console.log("Interview response:", data);

            if (
                data.completed ||
                data.question === "The interview is already complete." ||
                data.question === "This interview is already complete."
            ) {
                addMessage(
                    "system",
                    "World interview complete! Creating your world...",
                );
                if (
                    data.question &&
                    !data.question.includes("already complete")
                ) {
                    conversationHistory.push({
                        role: "assistant",
                        text: data.question,
                    });
                    addMessage("interview", data.question);
                }

                // Return to Lobby
                setTimeout(() => {
                    joinLobby();
                    addMessage(
                        "system",
                        "World created! Use 'worlds' to see it.",
                    );
                }, 2000);
            } else {
                // Get next question - check multiple possible field names
                const nextQuestion =
                    data.next_question ||
                    data.question ||
                    data.response ||
                    "Please continue...";
                currentQuestion = nextQuestion;
                conversationHistory.push({
                    role: "assistant",
                    text: nextQuestion,
                });
                addMessage("interview", nextQuestion);
            }
        } catch (error: any) {
            console.error("Failed to send response:", error);
            // ... (Recovery logic omitted for brevity, keep existing if possible or simplify)
            addMessage(
                "error",
                error.message ||
                    "Failed to process your response. Please try again.",
            );
        }
    }

    async function handleCommand(cmd?: string) {
        const input = cmd ? cmd.trim() : commandInput.trim();
        if (!input) return;

        addMessage("player", `> ${input}`);

        if (onboardingStep === "interview") {
            userResponse = input;
            commandInput = "";
            sendInterviewResponse();
            return;
        }

        // Intercept character creation command
        if (input.toLowerCase().startsWith("create character")) {
            await handleCharacterCommand(input);
            if (!cmd) commandInput = ""; // Only clear binding if used
            return;
        }

        // Intercept status command for UI modal
        if (input.trim().toLowerCase() === "status") {
            openCharacterSheet();
            if (!cmd) commandInput = "";
            return;
        }

        // Send raw text command to backend - backend will parse and process it
        gameWebSocket.sendCommandWithQueue(input);
        if (!cmd) commandInput = ""; // Only clear binding if used
    }

    async function listWorlds() {
        try {
            const res = await fetch(`${API_URL}/game/worlds`, {
                credentials: "include", // Send cookies
            });
            if (res.ok) {
                const worlds = await res.json();
                addMessage("system", "Available Worlds:");
                worlds.forEach((w: any) => {
                    // Make ID clickable or copyable
                    addMessage("system", `- ${w.Name} (ID: ${w.ID})`);
                });
                addMessage("system", "Type 'enter <world_id>' to enter.");
            } else {
                addMessage("error", "Failed to list worlds.");
            }
        } catch (e) {
            addMessage("error", "Failed to list worlds.");
        }
    }

    async function enterWorld(worldId: string) {
        try {
            addMessage(
                "system",
                `Checking entry options for world ${worldId}...`,
            );
            const res = await fetch(
                `${API_URL}/game/entry-options?world_id=${worldId}`,
                {
                    credentials: "include", // Send cookies
                },
            );

            if (res.ok) {
                const options = await res.json();
                entryOptions = options;
                targetWorldId = worldId;
                showEntryModal = true;
            } else {
                const err = await res.text();
                addMessage("error", `Cannot enter world: ${err}`);
            }
        } catch (e) {
            addMessage("error", "Failed to get entry options.");
        }
    }

    async function handleEntrySelection(event: CustomEvent) {
        const { type, data } = event.detail;
        if (!targetWorldId) return;

        showEntryModal = false;

        try {
            if (type === "cancel") {
                targetWorldId = null;
                return;
            }

            if (type === "watcher") {
                // Join as watcher
                await createCharacter(
                    "Watcher",
                    "Spirit",
                    targetWorldId,
                    "An invisible observer.",
                    "Watcher",
                    "watcher",
                );
                return;
            }

            if (type === "npc") {
                // Take over NPC
                // Create character with NPC details
                const npc = data;
                await createCharacter(
                    npc.name,
                    npc.species,
                    targetWorldId,
                    npc.description,
                    npc.occupation,
                    "player",
                    npc.appearance,
                );
            }

            if (type === "custom") {
                // Show custom character creation (text based for now)
                onboardingStep = "character";
                addMessage(
                    "system",
                    "Enter: create character <name> <species>",
                );
                // I need to store targetWorldId to use it when creating character
                // But `handleCharacterCommand` uses hardcoded world ID currently.
                // I should update `handleCharacterCommand` to use `targetWorldId`.
            }
        } catch (e) {
            console.error(e);
            addMessage("error", "Entry failed.");
        }
    }

    async function createCharacter(
        name: string,
        species: string,
        worldId: string,
        description?: string,
        occupation?: string,
        role?: string,
        appearance?: string,
    ) {
        try {
            const response = await fetch(`${API_URL}/game/characters`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                credentials: "include", // Send cookies
                body: JSON.stringify({
                    world_id: worldId,
                    name: name,
                    species: species,
                    role: role,
                    description: description,
                    occupation: occupation,
                    appearance: appearance,
                }),
            });

            if (!response.ok) {
                const errorData = await response.json();
                let errorMessage = "Failed to create character";
                if (typeof errorData.error === "string") {
                    errorMessage = errorData.error;
                } else if (errorData.error && errorData.error.message) {
                    errorMessage = errorData.error.message;
                }
                throw new Error(errorMessage);
            }

            const data = await response.json();
            addMessage(
                "system",
                `Character "${name}" created! Joining world...`,
            );

            // Join game
            await joinGame(data.character.character_id);
        } catch (error: any) {
            addMessage("error", error.message);
        }
    }

    async function joinGame(characterId: string) {
        try {
            const joinResponse = await fetch(`${API_URL}/game/join`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                credentials: "include", // Send cookies
                body: JSON.stringify({
                    character_id: characterId,
                }),
            });

            if (!joinResponse.ok) {
                const joinError = await joinResponse.json();
                console.error("Failed to join game:", joinError);
                throw new Error("Failed to join game session.");
            }

            const joinData = await joinResponse.json();

            // Reconnect WebSocket
            gameWebSocket.disconnect();
            gameWebSocket.connect(characterId);
            currentCharacterId = characterId;
            onboardingStep = "game";

            // Use the message from the backend (which includes welcome + look)
            if (joinData.message) {
                addMessage("system", joinData.message);
            } else {
                addMessage("system", "You have entered the world!");
            }
        } catch (e: any) {
            throw e;
        }
    }

    // ... (Keep existing handleCharacterCommand but update to use targetWorldId if available)
    async function handleCharacterCommand(cmd: string) {
        const parts = cmd.toLowerCase().split(" ");

        if (
            parts[0] === "create" &&
            parts[1] === "character" &&
            parts.length >= 4
        ) {
            const name = parts[2];
            const species =
                parts[3].charAt(0).toUpperCase() + parts[3].slice(1);

            // Use targetWorldId if set, otherwise default (which shouldn't happen in new flow)
            const worldId =
                targetWorldId || "00000000-0000-0000-0000-000000000001";

            await createCharacter(name, species, worldId);
        } else {
            addMessage(
                "error",
                "Invalid command. Use: create character <name> <species>",
            );
        }
    }

    // ... (Rest of functions)

    function handleServerMessage(msg: ServerMessage) {
        switch (msg.type) {
            case "game_message":
                const type = msg.data.type;
                if (type === "trigger_entry_options") {
                    const worldId =
                        msg.data.metadata?.world_id || msg.data.text;
                    enterWorld(worldId);
                } else if (type === "start_interview") {
                    startWorldInterview();
                } else {
                    addMessage(type, msg.data.text);
                }
                break;
            case "error":
                addMessage("error", msg.data.message);
                break;
        }
    }

    async function addMessage(type: string, text: string) {
        // Check scroll position before adding message
        let shouldScroll = false;
        const container = document.getElementById("game-output");

        if (container) {
            const { scrollTop, scrollHeight, clientHeight } = container;
            // If we are within 100px of the bottom, stay attached
            if (scrollHeight - scrollTop - clientHeight < 100) {
                shouldScroll = true;
            }
        } else {
            shouldScroll = true;
        }

        messages = [
            ...messages,
            {
                id: Date.now().toString() + Math.random(),
                type,
                text,
                timestamp: new Date(),
            },
        ];

        await tick();

        if (shouldScroll && container) {
            container.scrollTop = container.scrollHeight;
        }
    }

    function handleLogout() {
        gameWebSocket.disconnect();
        gameAPI.logout();
        goto("/");
    }

    function getMessageColor(type: string): string {
        switch (type) {
            case "error":
                return "text-red-400";
            case "system":
                return "text-cyan-400";
            case "interview":
                return "text-purple-400";
            case "command":
                return "text-gray-500";
            case "user":
                return "text-blue-400";
            case "success":
                return "text-green-400";
            default:
                return "text-gray-300";
        }
    }

    async function openCharacterSheet() {
        if (!currentCharacterId) {
            addMessage("error", "No active character.");
            return;
        }

        try {
            const data = await gameAPI.getSkills(currentCharacterId);
            characterSkills = data.skills || {};
            // In a real app we'd fetch stats too, skipping for now
            showCharacterSheet = true;
        } catch (e) {
            console.error(e);
            addMessage("error", "Failed to load skills.");
        }
    }
</script>

<div class="flex flex-col h-screen bg-gray-900 text-gray-100 font-mono">
    <!-- Header -->
    <header
        class="bg-gray-800 border-b border-gray-700 p-4 flex justify-between items-center"
    >
        <h1 class="text-xl font-bold text-blue-400">Thousand Worlds</h1>
        <div class="flex items-center gap-4">
            <!-- Connection Status -->
            <div
                class="flex items-center gap-2 text-sm"
                title={isConnected ? "Connected" : "Disconnected"}
            >
                <div
                    class={`w-3 h-3 rounded-full ${isConnected ? "bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)] connected" : "bg-red-500 animate-pulse"}`}
                ></div>
                <span class={isConnected ? "text-gray-400" : "text-red-400"}>
                    {isConnected ? "Connected" : "Reconnecting..."}
                </span>
            </div>

            {#if onboardingStep === "game"}
                <button
                    on:click={openCharacterSheet}
                    class="bg-blue-600 hover:bg-blue-500 text-white px-3 py-1 rounded text-sm transition-colors"
                >
                    Character
                </button>
            {/if}
            <button
                on:click={() => {
                    gameAPI.logout();
                    goto("/");
                }}
                class="text-sm text-gray-400 hover:text-white transition-colors"
            >
                Logout
            </button>
        </div>
    </header>

    <!-- Main Game Area -->
    <main class="flex-1 overflow-hidden flex flex-col relative">
        {#if onboardingStep === "checking"}
            <div class="flex-1 flex items-center justify-center">
                <div class="text-xl text-gray-400 animate-pulse">
                    Loading...
                </div>
            </div>
        {:else if onboardingStep === "interview"}
            <div
                class="flex-1 flex flex-col p-4 overflow-y-auto"
                id="game-output"
                data-testid="game-output"
            >
                {#each conversationHistory as msg}
                    <div
                        class={`mb-4 ${msg.role === "user" ? "text-right" : "text-left"}`}
                    >
                        <div
                            class={`inline-block p-3 rounded-lg max-w-[80%] ${
                                msg.role === "user"
                                    ? "bg-blue-900/50 text-blue-100"
                                    : "bg-gray-800/50 text-gray-100"
                            }`}
                        >
                            {msg.text}
                        </div>
                    </div>
                {/each}
                {#if currentQuestion}
                    <div class="mb-4 text-left">
                        <div
                            class="inline-block p-3 rounded-lg max-w-[80%] bg-gray-800/50 text-gray-100 border-l-4 border-blue-500"
                        >
                            {currentQuestion}
                        </div>
                    </div>
                {/if}
            </div>
        {:else}
            <!-- Game Output -->
            <div
                class="flex-1 overflow-y-auto p-4 space-y-2 scroll-smooth"
                id="game-output"
                data-testid="game-output"
            >
                {#each messages as msg (msg.id)}
                    <div
                        class={`message leading-relaxed ${
                            msg.type === "error"
                                ? "text-red-400"
                                : msg.type === "system"
                                  ? "text-yellow-400"
                                  : msg.type === "player"
                                    ? "text-blue-300"
                                    : msg.type === "emote"
                                      ? "text-orange-300 italic"
                                      : "text-gray-300"
                        }`}
                    >
                        {@html msg.text.replace(/\n/g, "<br>")}
                    </div>
                {/each}
            </div>
        {/if}

        <!-- Input Area -->
        <div class="p-4 bg-gray-800 border-t border-gray-700 space-y-3">
            {#if onboardingStep !== "interview" && onboardingStep !== "checking"}
                <QuickButtons on:submit={(e) => handleCommand(e.detail)} />
            {/if}

            {#if onboardingStep === "interview"}
                <!-- Interview mode: simple input -->
                <div class="relative">
                    <input
                        type="text"
                        bind:value={commandInput}
                        on:keydown={(e) => e.key === "Enter" && handleCommand()}
                        placeholder="Answer the question..."
                        class="w-full bg-gray-900 border border-gray-600 rounded-lg px-4 py-3 text-gray-100 text-base focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all"
                        style="font-size: 16px;"
                    />
                </div>
            {:else if onboardingStep !== "checking"}
                <!-- Game mode: use thin-client CommandInput -->
                <CommandInput on:submit={(e) => handleCommand(e.detail)} />
            {/if}
        </div>

        <!-- Entry Modal -->

        {#if showEntryModal && entryOptions}
            <WorldEntry
                options={entryOptions}
                on:select={handleEntrySelection}
            />
        {/if}

        <!-- Character Sheet Modal -->
        {#if showCharacterSheet}
            <div
                class="absolute inset-0 bg-black/80 flex items-center justify-center p-4 z-50 transition-opacity"
                role="button"
                tabindex="0"
                on:click={() => (showCharacterSheet = false)}
                on:keydown={(e) =>
                    e.key === "Escape" && (showCharacterSheet = false)}
            >
                <div
                    class="relative max-w-md w-full"
                    role="document"
                    on:click|stopPropagation
                    on:keydown|stopPropagation
                    tabindex="-1"
                >
                    <button
                        class="absolute -top-2 -right-2 bg-gray-700 rounded-full w-8 h-8 flex items-center justify-center text-white hover:bg-gray-600 z-10"
                        on:click={() => (showCharacterSheet = false)}
                        data-testid="close-character-sheet"
                    >
                        ✕
                    </button>
                    <CharacterSheet
                        characterName={characterStats.name}
                        level={characterStats.level}
                        experience={characterStats.experience}
                        nextLevelXP={characterStats.nextLevelXP}
                        strength={characterStats.attributes.strength}
                        dexterity={characterStats.attributes.dexterity}
                        constitution={characterStats.attributes.constitution}
                        intelligence={characterStats.attributes.intelligence}
                        wisdom={characterStats.attributes.wisdom}
                        charisma={characterStats.attributes.charisma}
                        skills={characterSkills}
                    />
                </div>
            </div>
        {/if}
    </main>
</div>

<style>
    .messages-container::-webkit-scrollbar {
        width: 8px;
    }

    .messages-container::-webkit-scrollbar-track {
        background: rgba(0, 0, 0, 0.3);
    }

    .messages-container::-webkit-scrollbar-thumb {
        background: rgba(55, 65, 81, 0.5);
        border-radius: 4px;
    }

    .messages-container::-webkit-scrollbar-thumb:hover {
        background: rgba(75, 85, 99, 0.7);
    }
</style>
