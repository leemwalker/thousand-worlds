import { c as create_ssr_component, e as escape, b as add_attribute } from "../../chunks/ssr.js";
import "@sveltejs/kit/internal";
import "../../chunks/exports.js";
import "../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../chunks/state.svelte.js";
import "../../chunks/game.js";
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let email = "";
  let password = "";
  return `<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900"><div class="bg-gray-800 p-8 rounded-lg shadow-2xl w-full max-w-md border border-gray-700"><h1 class="text-3xl font-bold text-center mb-6 text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-500" data-svelte-h="svelte-622iwc">Thousand Worlds</h1> <h2 class="text-xl font-semibold text-gray-100 mb-6 text-center">${escape("Sign In")}</h2> ${``} <form class="space-y-4"><div><label for="email" class="block text-sm font-medium text-gray-300 mb-2" data-svelte-h="svelte-6hciz4">Email</label> <input id="email" type="email" required ${""} class="w-full px-4 py-2 bg-gray-900 border border-gray-600 rounded text-gray-100 focus:outline-none focus:border-blue-500 disabled:opacity-50" placeholder="your@email.com"${add_attribute("value", email, 0)}></div> ${``} <div><label for="password" class="block text-sm font-medium text-gray-300 mb-2" data-svelte-h="svelte-phjo48">Password</label> <input id="password" type="password" required ${""}${add_attribute("minlength", void 0, 0)} class="w-full px-4 py-2 bg-gray-900 border border-gray-600 rounded text-gray-100 focus:outline-none focus:border-blue-500 disabled:opacity-50" placeholder="••••••••"${add_attribute("value", password, 0)}> ${``}</div> <button type="submit" ${""} class="w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-bold py-3 px-4 rounded transition-all disabled:opacity-50 disabled:cursor-not-allowed">${escape("Sign In")}</button></form> <div class="mt-6 text-center"><button ${""} class="text-blue-400 hover:text-blue-300 text-sm transition-colors disabled:opacity-50">${escape(
    "Don't have an account? Sign up"
  )}</button></div></div></div>`;
});
export {
  Page as default
};
