<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { goto } from "$app/navigation";
    import { gameAPI } from "$lib/services/api";
    import { gameWebSocket, type ServerMessage } from "$lib/services/websocket";
    import WorldEntry from "$lib/components/WorldEntry.svelte";

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

    // Use relative URL to go through Vite proxy with extended timeout
    const API_URL = "/api";

    onMount(async () => {
        const token = gameAPI.getToken();
        if (!token) {
            goto("/");
            return;
        }

        // Don't connect WebSocket yet - will connect after checking character status
        unsubscribe = gameWebSocket.onMessage(handleServerMessage);

        // Check onboarding status
        await checkOnboardingStatus(token);
    });

    onDestroy(() => {
        if (unsubscribe) unsubscribe();
    });

    async function checkOnboardingStatus(token: string) {
        try {
            // Check for active interview
            const interviewRes = await fetch(
                `${API_URL}/world/interview/active`,
                {
                    headers: { Authorization: `Bearer ${token.trim()}` },
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
                            joinLobby(token);
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

            // Always join Lobby first
            await joinLobby(token);
        } catch (error) {
            console.error("Onboarding check failed:", error);
            onboardingStep = "game";
            addMessage(
                "error",
                "Failed to check status. Starting in game mode.",
            );
        }
    }

    async function joinLobby(token: string) {
        console.log(
            "[Lobby] Joining lobby with token:",
            token.substring(0, 20) + "...",
        );
        try {
            // Join Lobby (backend handles Ghost creation or existing character)
            // Lobby ID is 00000000-0000-0000-0000-000000000000, but we just connect without character_id to auto-join lobby
            // Wait, handler logic: if character_id is nil, join lobby.

            console.log("[Lobby] Disconnecting existing WebSocket...");
            gameWebSocket.disconnect();
            await new Promise((resolve) => setTimeout(resolve, 100));

            console.log("[Lobby] Connecting to WebSocket...");
            gameWebSocket.connect(token); // No character ID -> Lobby

            onboardingStep = "lobby";
            addMessage("system", "Welcome to the Lobby.");
            addMessage("system", "Type 'worlds' to see available worlds.");
            addMessage(
                "system",
                "Type 'create world' to start a new world interview.",
            );
            console.log("[Lobby] Lobby join complete");
        } catch (error) {
            console.error("[Lobby] Failed to join lobby:", error);
            addMessage("error", "Failed to join lobby.");
        }
    }

    async function startWorldInterview(token: string) {
        try {
            addMessage(
                "system",
                "Welcome to Thousand Worlds! Let's create your custom world...",
            );

            const response = await fetch(`${API_URL}/world/interview/start`, {
                method: "POST",
                headers: {
                    Authorization: `Bearer ${token}`,
                    "Content-Type": "application/json",
                },
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

        const token = gameAPI.getToken();
        if (!token) return;

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
                    Authorization: `Bearer ${token}`,
                    "Content-Type": "application/json",
                },
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
                    joinLobby(token);
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

    async function handleCommand() {
        if (!commandInput.trim()) return;

        const input = commandInput.trim();
        addMessage("player", `> ${input}`);

        if (onboardingStep === "interview") {
            userResponse = input;
            commandInput = "";
            sendInterviewResponse();
            return;
        }

        if (onboardingStep === "character") {
            // Legacy character creation step, now handled via modal
            handleCharacterCommand(input); // Keep for fallback
            commandInput = "";
            return;
        }

        const parts = input.toLowerCase().split(/\s+/);
        const action = parts[0];
        const rest = parts.slice(1);

        // Lobby specific commands
        if (onboardingStep === "lobby") {
            // Format lobby commands properly for backend
            const command: any = { action };

            if (action === "create" && rest[0] === "world") {
                command.action = "create";
                command.target = "world";
                gameWebSocket.sendCommand(command);
                commandInput = "";
                return;
            }

            if (action === "enter" && rest.length > 0) {
                command.action = "enter";
                command.target = rest[0];
                gameWebSocket.sendCommand(command);
                commandInput = "";
                return;
            }

            if (["look", "l", "who", "help"].includes(action)) {
                command.action = action;
                gameWebSocket.sendCommand(command);
                commandInput = "";
                return;
            }

            if (action === "say" && rest.length > 0) {
                command.action = "say";
                command.message = rest.join(" ");
                gameWebSocket.sendCommand(command);
                commandInput = "";
                return;
            }
        }

        // Standard game commands
        const command: any = { action };

        // Parse based on command type (keep existing logic)
        if (["move", "m"].includes(action)) {
            if (rest.length > 0) {
                // Map common variations to directions
                const directionMap: Record<string, string> = {
                    n: "north",
                    north: "north",
                    s: "south",
                    south: "south",
                    e: "east",
                    east: "east",
                    w: "west",
                    west: "west",
                    ne: "northeast",
                    northeast: "northeast",
                    nw: "northwest",
                    northwest: "northwest",
                    se: "southeast",
                    southeast: "southeast",
                    sw: "southwest",
                    southwest: "southwest",
                    u: "up",
                    up: "up",
                    d: "down",
                    down: "down",
                };
                command.action = "move";
                command.direction = directionMap[rest[0]] || rest[0];
            }
        } else if (["look", "l", "examine", "ex"].includes(action)) {
            command.action = "look";
            if (rest.length > 0) {
                command.target = rest.join(" ");
            }
        } else if (["take", "get", "grab"].includes(action)) {
            command.action = "take";
            if (rest.length > 0) {
                command.target = rest.join(" ");
            }
        } else if (["drop"].includes(action)) {
            command.action = "drop";
            if (rest.length > 0) {
                command.target = rest.join(" ");
            }
        } else if (["attack", "kill", "fight"].includes(action)) {
            command.action = "attack";
            if (rest.length > 0) {
                command.target = rest.join(" ");
            }
        } else if (["talk", "speak", "say"].includes(action)) {
            command.action = "talk";
            if (rest.length > 0) {
                command.target = rest.join(" ");
            }
        } else if (["craft", "make", "create"].includes(action)) {
            command.action = "craft";
            if (rest.length > 0) {
                command.target = rest.join(" ");
            }
        } else if (["use"].includes(action)) {
            command.action = "use";
            if (rest.length > 0) {
                command.target = rest.join(" ");
            }
        } else if (["inventory", "inv", "i"].includes(action)) {
            command.action = "inventory";
        } else if (["help", "h", "?"].includes(action)) {
            command.action = "help";
        }

        gameWebSocket.sendCommand(command);
        commandInput = "";
    }

    async function listWorlds() {
        const token = gameAPI.getToken();
        if (!token) return;

        try {
            const res = await fetch(`${API_URL}/game/worlds`, {
                headers: { Authorization: `Bearer ${token}` },
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
        const token = gameAPI.getToken();
        if (!token) return;

        try {
            addMessage(
                "system",
                `Checking entry options for world ${worldId}...`,
            );
            const res = await fetch(
                `${API_URL}/game/entry-options?world_id=${worldId}`,
                {
                    headers: { Authorization: `Bearer ${token}` },
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
        const token = gameAPI.getToken();
        if (!token || !targetWorldId) return;

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
        const token = gameAPI.getToken();
        if (!token) return;

        try {
            const response = await fetch(`${API_URL}/game/characters`, {
                method: "POST",
                headers: {
                    Authorization: `Bearer ${token}`,
                    "Content-Type": "application/json",
                },
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
                const error = await response.json();
                throw new Error(error.error || "Failed to create character");
            }

            const data = await response.json();
            addMessage(
                "success",
                `Character "${name}" created! Joining world...`,
            );

            // Join game
            await fetch(`${API_URL}/game/join`, {
                method: "POST",
                headers: {
                    Authorization: `Bearer ${token}`,
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    character_id: data.character.character_id,
                }),
            });

            // Reconnect WebSocket
            gameWebSocket.disconnect();
            gameWebSocket.connect(token, data.character.character_id);
            onboardingStep = "game";
            addMessage("system", "You have entered the world!");
        } catch (error: any) {
            addMessage("error", error.message);
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
                    const token = gameAPI.getToken();
                    if (token) startWorldInterview(token);
                } else {
                    addMessage(type, msg.data.text);
                }
                break;
            case "error":
                addMessage("error", msg.data.message);
                break;
        }
    }

    function addMessage(type: string, text: string) {
        messages = [
            ...messages,
            {
                id: Date.now().toString() + Math.random(),
                type,
                text,
                timestamp: new Date(),
            },
        ];

        // Auto-scroll to bottom
        setTimeout(() => {
            const container = document.querySelector(".messages-container");
            if (container) {
                container.scrollTop = container.scrollHeight;
            }
        }, 50);
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
</script>

<div class="min-h-screen bg-black flex items-center justify-center p-4">
    <!-- Logout button -->
    <button
        on:click={handleLogout}
        class="absolute top-4 right-4 z-50 text-gray-600 hover:text-blue-400 px-4 py-2 text-xs font-mono uppercase tracking-widest transition-colors"
    >
        [ Logout ]
    </button>

    <!-- Terminal Container -->
    <div
        class="w-full max-w-4xl bg-gray-900 rounded-lg shadow-2xl border-2 border-blue-500/30 overflow-hidden flex flex-col h-[85vh] max-h-[700px]"
    >
        <!-- Terminal Header -->
        <div
            class="bg-blue-950/50 px-4 py-2 border-b border-blue-500/30 flex items-center justify-between shrink-0"
        >
            <div class="flex gap-2">
                <div
                    class="w-3 h-3 rounded-full bg-red-500/30 border border-red-500/50"
                ></div>
                <div
                    class="w-3 h-3 rounded-full bg-yellow-500/30 border border-yellow-500/50"
                ></div>
                <div
                    class="w-3 h-3 rounded-full bg-green-500/30 border border-green-500/50"
                ></div>
            </div>
            <span
                class="text-blue-400 text-xs font-mono uppercase tracking-widest"
            >
                {#if onboardingStep === "interview"}
                    World_Creation_Protocol
                {:else if onboardingStep === "character"}
                    Character_Initialization
                {:else if onboardingStep === "checking"}
                    System_Check
                {:else}
                    Thousand_Worlds_Terminal
                {/if}
            </span>
            <div class="w-14"></div>
            <!-- Spacer for centering -->
        </div>

        <!-- Messages Display -->
        <div
            class="messages-container flex-1 overflow-y-auto p-6 font-mono text-sm md:text-base leading-relaxed bg-black"
        >
            {#each messages as message}
                <div class="mb-2 {getMessageColor(message.type)}">
                    <span class="whitespace-pre-wrap font-medium"
                        >{message.text || ""}</span
                    >
                </div>
            {/each}

            {#if onboardingStep === "checking"}
                <div class="text-cyan-500 animate-pulse mt-4">
                    > System check in progress...
                </div>
            {/if}
        </div>

        <!-- Command Input -->
        <div class="bg-black p-4 shrink-0 border-t border-blue-500/20">
            <form
                on:submit|preventDefault={handleCommand}
                class="flex gap-3 items-center"
            >
                <span class="text-blue-400 font-mono text-lg font-bold"
                    >{">"}</span
                >
                <input
                    bind:value={commandInput}
                    type="text"
                    placeholder={onboardingStep === "interview" ? "" : ""}
                    class="flex-1 bg-transparent border-none p-0 text-blue-100 font-mono text-base focus:ring-0 focus:outline-none placeholder-gray-700 caret-blue-400"
                    autocomplete="off"
                    autoFocus
                />
            </form>
        </div>
    </div>

    {#if showEntryModal && entryOptions}
        <WorldEntry options={entryOptions} on:select={handleEntrySelection} />
    {/if}
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
