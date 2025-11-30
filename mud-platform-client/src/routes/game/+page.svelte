<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { goto } from "$app/navigation";
    import { gameAPI } from "$lib/services/api";
    import { gameWebSocket, type ServerMessage } from "$lib/services/websocket";

    // Onboarding state
    let onboardingStep: "checking" | "interview" | "character" | "game" =
        "checking";
    let interviewSessionId: string | null = null;
    let currentQuestion: string = "";
    let userResponse: string = "";
    let conversationHistory: Array<{ role: string; text: string }> = [];

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
                            onboardingStep = "character";
                            addMessage(
                                "system",
                                "Now let's create your character...",
                            );
                            addMessage(
                                "system",
                                "Available species: Human, Elf, Dwarf",
                            );
                            addMessage(
                                "system",
                                "Type: create character <name> <species>",
                            );
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

            // Check for characters
            const charactersRes = await fetch(`${API_URL}/game/characters`, {
                headers: { Authorization: `Bearer ${token}` },
            });

            if (charactersRes.ok) {
                const data = await charactersRes.json();
                if (data.characters && data.characters.length > 0) {
                    // Has characters - join the first one
                    const character = data.characters[0];

                    try {
                        await fetch(`${API_URL}/game/join`, {
                            method: "POST",
                            headers: {
                                Authorization: `Bearer ${token}`,
                                "Content-Type": "application/json",
                            },
                            body: JSON.stringify({
                                character_id: character.character_id,
                            }),
                        });

                        // Reconnect WebSocket with character context
                        gameWebSocket.disconnect();
                        await new Promise((resolve) =>
                            setTimeout(resolve, 100),
                        ); // Brief delay
                        gameWebSocket.connect(token, character.character_id);

                        onboardingStep = "game";
                        addMessage(
                            "system",
                            `Welcome back, ${character.name}! Type 'help' to see available commands.`,
                        );
                    } catch (error) {
                        console.error("Failed to join game:", error);
                        onboardingStep = "game";
                        addMessage(
                            "error",
                            "Failed to join game. Please try refreshing.",
                        );
                    }
                    return;
                }
            }

            // New user - start world interview
            gameWebSocket.connect(token); // Connect without character for interview
            onboardingStep = "interview";
            await startWorldInterview(token);
        } catch (error) {
            console.error("Onboarding check failed:", error);
            onboardingStep = "game";
            addMessage(
                "error",
                "Failed to check status. Starting in game mode.",
            );
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
                const summary =
                    data.summary ||
                    data.world_summary ||
                    "Your world is ready!";
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

                // Transition to character creation
                setTimeout(() => {
                    onboardingStep = "character";
                    addMessage("system", "Now let's create your character...");
                    addMessage(
                        "system",
                        "Available species: Human, Elf, Dwarf",
                    );
                    addMessage(
                        "system",
                        "Type: create character <name> <species>",
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

            // Attempt recovery: Check if the backend actually processed the message
            try {
                console.log("Attempting to recover from error...");
                const statusRes = await fetch(
                    `${API_URL}/world/interview/active`,
                    {
                        headers: { Authorization: `Bearer ${token}` },
                    },
                );

                if (statusRes.ok) {
                    const statusData = await statusRes.json();
                    console.log("Recovery status check:", statusData);

                    // If we got a new question that's different from the last one we showed, recover!
                    const lastAssistantMessage = conversationHistory
                        .filter((m) => m.role === "assistant")
                        .pop();
                    const newQuestion =
                        statusData.next_question || statusData.question;

                    // Check for completion messages in recovery
                    if (
                        newQuestion === "The interview is already complete." ||
                        newQuestion === "This interview is already complete."
                    ) {
                        console.log("Recovered! Interview is complete.");
                        addMessage(
                            "system",
                            "World interview complete! Creating your world...",
                        );

                        setTimeout(() => {
                            onboardingStep = "character";
                            addMessage(
                                "system",
                                "Now let's create your character...",
                            );
                            addMessage(
                                "system",
                                "Available species: Human, Elf, Dwarf",
                            );
                            addMessage(
                                "system",
                                "Type: create character <name> <species>",
                            );
                        }, 2000);
                        return;
                    }

                    if (
                        newQuestion &&
                        (!lastAssistantMessage ||
                            lastAssistantMessage.text !== newQuestion)
                    ) {
                        console.log(
                            "Recovered! Backend processed the message despite error.",
                        );
                        currentQuestion = newQuestion;
                        conversationHistory.push({
                            role: "assistant",
                            text: newQuestion,
                        });
                        addMessage("interview", newQuestion);
                        return; // Successfully recovered
                    }
                }
            } catch (recoveryError) {
                console.error("Recovery failed:", recoveryError);
            }

            if (error.name === "AbortError") {
                addMessage(
                    "error",
                    "Request timed out. The LLM is taking too long. Please try again.",
                );
            } else if (
                error.message?.includes("network") ||
                error.message?.includes("fetch")
            ) {
                addMessage(
                    "error",
                    "Network error. Please check your connection and try again.",
                );
            } else {
                addMessage(
                    "error",
                    error.message ||
                        "Failed to process your response. Please try again.",
                );
            }
        }
    }

    function handleCommand() {
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
            commandInput = "";
            return;
        }

        // Send to game server
        gameWebSocket.sendCommand({ action: cmd });
        commandInput = "";
    }

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

            const token = gameAPI.getToken();
            if (!token) return;

            try {
                addMessage(
                    "system",
                    `Creating character "${name}" as ${species}...`,
                );

                const response = await fetch(`${API_URL}/game/characters`, {
                    method: "POST",
                    headers: {
                        Authorization: `Bearer ${token}`,
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify({
                        world_id: "00000000-0000-0000-0000-000000000001", // TODO: Use actual world ID
                        name: name,
                        species: species,
                    }),
                });

                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(
                        error.error || "Failed to create character",
                    );
                }

                const data = await response.json();
                addMessage(
                    "success",
                    `Character "${name}" created successfully!`,
                );
                addMessage(
                    "system",
                    `HP: ${data.secondary_attributes.max_hp}, Stamina: ${data.secondary_attributes.max_stamina}`,
                );

                // Join game
                setTimeout(async () => {
                    try {
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

                        // Reconnect WebSocket with character context
                        gameWebSocket.disconnect();
                        gameWebSocket.connect(
                            token,
                            data.character.character_id,
                        );

                        onboardingStep = "game";
                        addMessage(
                            "system",
                            "You have entered the world! Type 'help' to see commands.",
                        );
                    } catch (error) {
                        addMessage(
                            "error",
                            "Failed to join game. Please refresh.",
                        );
                    }
                }, 1000);
            } catch (error: any) {
                addMessage(
                    "error",
                    error.message || "Failed to create character",
                );
            }
        } else {
            addMessage(
                "error",
                "Invalid command. Use: create character <name> <species>",
            );
        }
    }

    function handleServerMessage(msg: ServerMessage) {
        switch (msg.type) {
            case "game_message":
                addMessage(msg.data.type, msg.data.text);
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
