<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { gameAPI } from "$lib/services/api";
  import { gameWebSocket } from "$lib/services/websocket";

  let email = "";
  let username = "";
  let password = "";
  let isLogin = true;
  let error = "";
  let loading = false;

  onMount(async () => {
    // Check if already logged in by attempting to get current user
    // Cookie will be sent automatically with the request
    try {
      await gameAPI.getMe();
      // If getMe succeeds, user is authenticated via cookie
      goto("/game");
    } catch (err) {
      // Not authenticated, stay on login page
    }
  });

  async function handleSubmit() {
    error = "";
    loading = true;

    try {
      if (isLogin) {
        // Login
        let response;
        try {
          response = await gameAPI.login(email, password);
          console.log("Login successful");
        } catch (e: any) {
          console.error("Login API error:", e);
          throw new Error(e.message || "Login failed");
        }

        // Redirect to game
        goto("/game");
      } else {
        // Register
        await gameAPI.register(email, username, password);

        // Auto-login after registration
        await gameAPI.login(email, password);
        goto("/game");
      }
    } catch (err: any) {
      console.error("Form submission error:", err);

      if (err.name === "AuthError" && err.userFriendly) {
        error = err.userFriendly.message;
        // You could also use err.userFriendly.title or err.userFriendly.action here
        // For now, we'll just show the message to keep it simple but actionable
      } else if (typeof err === "string") {
        error = err;
      } else if (err.message) {
        error = err.message;
      } else {
        error = "An unexpected error occurred. Please try again.";
      }
    } finally {
      loading = false;
    }
  }

  function toggleMode() {
    isLogin = !isLogin;
    error = "";
  }
</script>

<div
  class="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900"
>
  <div
    class="bg-gray-800 p-8 rounded-lg shadow-2xl w-full max-w-md border border-gray-700"
  >
    <h1
      class="text-3xl font-bold text-center mb-6 text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-500"
    >
      Thousand Worlds
    </h1>

    <h2 class="text-xl font-semibold text-gray-100 mb-6 text-center">
      {isLogin ? "Sign In" : "Create Account"}
    </h2>

    {#if error}
      <div
        class="bg-red-900/50 border border-red-500 text-red-200 px-4 py-3 rounded mb-4"
      >
        {error}
      </div>
    {/if}

    <form on:submit|preventDefault={handleSubmit} class="space-y-4">
      <div>
        <label for="email" class="block text-sm font-medium text-gray-300 mb-2">
          Email
        </label>
        <input
          id="email"
          type="email"
          bind:value={email}
          required
          disabled={loading}
          class="w-full px-4 py-2 bg-gray-900 border border-gray-600 rounded text-gray-100 focus:outline-none focus:border-blue-500 disabled:opacity-50"
          placeholder="your@email.com"
        />
      </div>

      {#if !isLogin}
        <div>
          <label
            for="username"
            class="block text-sm font-medium text-gray-300 mb-2"
          >
            Username
          </label>
          <input
            id="username"
            type="text"
            bind:value={username}
            required
            disabled={loading}
            minlength="3"
            maxlength="20"
            class="w-full px-4 py-2 bg-gray-900 border border-gray-600 rounded text-gray-100 focus:outline-none focus:border-blue-500 disabled:opacity-50"
            placeholder="Username"
          />
        </div>
      {/if}

      <div>
        <label
          for="password"
          class="block text-sm font-medium text-gray-300 mb-2"
        >
          Password
        </label>
        <input
          id="password"
          type="password"
          bind:value={password}
          required
          disabled={loading}
          minlength={isLogin ? undefined : 8}
          class="w-full px-4 py-2 bg-gray-900 border border-gray-600 rounded text-gray-100 focus:outline-none focus:border-blue-500 disabled:opacity-50"
          placeholder="••••••••"
        />
        {#if !isLogin}
          <p class="text-xs text-gray-400 mt-1">Minimum 8 characters</p>
        {/if}
      </div>

      <button
        type="submit"
        disabled={loading}
        class="w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-bold py-3 px-4 rounded transition-all disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {loading ? "Please wait..." : isLogin ? "Sign In" : "Create Account"}
      </button>
    </form>

    <div class="mt-6 text-center">
      <button
        on:click={toggleMode}
        disabled={loading}
        class="text-blue-400 hover:text-blue-300 text-sm transition-colors disabled:opacity-50"
      >
        {isLogin
          ? "Don't have an account? Sign up"
          : "Already have an account? Sign in"}
      </button>
    </div>
  </div>
</div>
