const BACKEND_URL = process.env.BACKEND_URL || "http://game-server:8080";
const EXCLUDED_HEADERS = /* @__PURE__ */ new Set([
  "connection",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailers",
  "transfer-encoding",
  "upgrade",
  "host"
]);
const handle = async ({ event, resolve }) => {
  if (event.url.pathname.startsWith("/api")) {
    const backendPath = event.url.pathname;
    const backendUrl = `${BACKEND_URL}${backendPath}${event.url.search}`;
    try {
      const headers = new Headers();
      for (const [key, value] of event.request.headers.entries()) {
        if (!EXCLUDED_HEADERS.has(key.toLowerCase())) {
          headers.set(key, value);
        }
      }
      const response = await fetch(backendUrl, {
        method: event.request.method,
        headers,
        body: event.request.method !== "GET" && event.request.method !== "HEAD" ? await event.request.text() : void 0
      });
      const responseHeaders = new Headers();
      for (const [key, value] of response.headers.entries()) {
        if (!EXCLUDED_HEADERS.has(key.toLowerCase())) {
          responseHeaders.set(key, value);
        }
      }
      return new Response(response.body, {
        status: response.status,
        statusText: response.statusText,
        headers: responseHeaders
      });
    } catch (error) {
      console.error("API proxy error:", error);
      return new Response(JSON.stringify({ error: "Unable to reach backend server" }), {
        status: 503,
        headers: { "Content-Type": "application/json" }
      });
    }
  }
  return resolve(event);
};
export {
  handle
};
