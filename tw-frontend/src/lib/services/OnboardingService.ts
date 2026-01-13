import { gameAPI } from "$lib/services/api";
import { gameWebSocket } from "$lib/services/websocket";

export interface InterviewMessage {
    role: string;
    text: string;
}

export type OnboardingStep = "checking" | "interview" | "character" | "game" | "lobby";

export interface OnboardingState {
    step: OnboardingStep;
    sessionId: string | null;
    currentQuestion: string;
    conversationHistory: InterviewMessage[];
    error: string | null;
}

export class OnboardingService {
    private api: typeof gameAPI;
    private ws: typeof gameWebSocket;

    constructor(api = gameAPI, ws = gameWebSocket) {
        this.api = api;
        this.ws = ws;
    }

    async checkStatus(currentUser: any): Promise<OnboardingState> {
        const state: OnboardingState = {
            step: "checking",
            sessionId: null,
            currentQuestion: "",
            conversationHistory: [],
            error: null
        };

        try {
            // Check for active interview
            try {
                // We need to use fetch directly or expose a method in gameAPI
                // Since gameAPI is imported, let's assume we can extend it or use fetch if gameAPI doesn't have it.
                // Looking at gameAPI usage in +page.svelte, it uses fetch for these specific endpoints.
                // We should probably move these fetches *into* this service.

                const response = await fetch("/api/world/interview/active");
                if (response.ok) {
                    const interview = await response.json();
                    if (interview && interview.status === "in_progress") {
                        state.step = "interview";
                        state.sessionId = interview.session_id;

                        if (OnboardingService.isInterviewComplete(interview.question)) {
                            // Already complete
                            state.step = "lobby"; // Will trigger joinLobby in UI
                            return state;
                        }

                        state.conversationHistory = interview.conversation || [];
                        if (state.conversationHistory.length > 0) {
                            const last = state.conversationHistory[state.conversationHistory.length - 1];
                            if (last && last.role === "assistant") {
                                state.currentQuestion = last.text;
                            }
                        } else if (interview.question) {
                            state.currentQuestion = interview.question;
                        }
                        return state;
                    }
                }
            } catch (e) {
                console.warn("Failed to check active interview", e);
            }

            // Check for auto-resume
            if (currentUser?.last_world_id) {
                try {
                    const data = await this.api.getCharacters();
                    if (data && data.characters) {
                        const char = data.characters.find((c: any) => c.world_id === currentUser.last_world_id);
                        if (char) {
                            // Let UI handle the actual join call using the ID, or we return it
                            // Ideally state should include "readyToJoinCharacterId"
                            // For now, let's return "game" and let logic handle it? 
                            // Or better: UI logic is complex. 
                            // Let's stick to returning "lobby" as default fallback.
                        }
                    }
                } catch (e) {
                    console.warn("Auto-resume check failed", e);
                }
            }

            state.step = "lobby";
            return state;

        } catch (error: any) {
            state.step = "game"; // Default error state in original code
            state.error = error.message || "Failed to check status";
            return state;
        }
    }

    async startInterview(): Promise<OnboardingState> {
        const state: OnboardingState = {
            step: "interview",
            sessionId: null,
            currentQuestion: "",
            conversationHistory: [],
            error: null
        };

        try {
            const response = await fetch("/api/world/interview/start", {
                method: "POST",
                headers: { "Content-Type": "application/json" }
            });

            if (!response.ok) throw new Error(`Failed to start interview: ${response.status}`);

            const data = await response.json();
            if (!data.session_id || !data.question) throw new Error("Invalid response from server");

            state.sessionId = data.session_id;
            state.currentQuestion = data.question || "Tell me about the world you'd like to create.";
            state.conversationHistory.push({ role: "assistant", text: state.currentQuestion });

            return state;

        } catch (error: any) {
            state.error = error.message;
            return state;
        }
    }

    async sendResponse(sessionId: string, message: string, history: InterviewMessage[]): Promise<{
        completed: boolean;
        nextQuestion?: string;
        error?: string;
        newHistory: InterviewMessage[];
    }> {
        const newHistory = [...history, { role: "user", text: message }];

        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 60000);

            const response = await fetch("/api/world/interview/message", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ session_id: sessionId, message }),
                signal: controller.signal
            });

            clearTimeout(timeoutId);

            if (!response.ok) throw new Error(`Failed to send message: ${response.status}`);

            const data = await response.json();

            if (data.completed || OnboardingService.isInterviewComplete(data.question)) {
                if (data.question && !OnboardingService.isInterviewComplete(data.question)) {
                    newHistory.push({ role: "assistant", text: data.question });
                }
                return { completed: true, newHistory };
            }

            const nextQ = data.next_question || data.question || data.response || "Please continue...";
            newHistory.push({ role: "assistant", text: nextQ });

            return { completed: false, nextQuestion: nextQ, newHistory };

        } catch (error: any) {
            return { completed: false, error: error.message, newHistory };
        }
    }

    static isInterviewComplete(text?: string): boolean {
        if (!text) return false;
        return text === "The interview is already complete." ||
            text === "This interview is already complete.";
    }
}
