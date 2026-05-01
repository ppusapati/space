import { h as head, a as attr } from "../../../../chunks/index2.js";
import "@sveltejs/kit/internal";
import "../../../../chunks/exports.js";
import "../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../chunks/state.svelte.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let email = "";
    let password = "";
    let rememberMe = false;
    let isLoading = false;
    head("8k30lk", $$renderer2, ($$renderer3) => {
      $$renderer3.title(($$renderer4) => {
        $$renderer4.push(`<title>Login - samavāya ERP</title>`);
      });
    });
    $$renderer2.push(`<div class="w-full"><h2 class="text-2xl font-semibold text-text mb-xs">Welcome back</h2> <p class="text-text-secondary mb-xl">Sign in to your account</p> <form class="flex flex-col gap-lg">`);
    {
      $$renderer2.push("<!--[!-->");
    }
    $$renderer2.push(`<!--]--> <div class="flex flex-col gap-xs"><label for="email" class="flex justify-between items-center text-sm font-medium text-text">Email</label> <input type="email" id="email"${attr("value", email)} required autocomplete="email" class="form-input" placeholder="you@company.com"${attr("disabled", isLoading, true)}/></div> <div class="flex flex-col gap-xs"><label for="password" class="flex justify-between items-center text-sm font-medium text-text">Password <a href="/forgot-password" class="text-xs font-normal text-primary">Forgot password?</a></label> <input type="password" id="password"${attr("value", password)} required autocomplete="current-password" class="form-input" placeholder="Enter your password"${attr("disabled", isLoading, true)}/></div> <div class="flex items-center"><label class="flex items-center gap-sm text-sm text-text-secondary cursor-pointer"><input type="checkbox"${attr("checked", rememberMe, true)}${attr("disabled", isLoading, true)} class="form-checkbox"/> <span>Remember me</span></label></div> <button type="submit" class="btn btn-primary w-full"${attr("disabled", isLoading, true)}>`);
    {
      $$renderer2.push("<!--[!-->");
      $$renderer2.push(`Sign in`);
    }
    $$renderer2.push(`<!--]--></button></form> <p class="text-center mt-xl text-sm text-text-secondary">Don't have an account? <a href="/register" class="text-primary font-medium">Sign up</a></p></div>`);
  });
}
export {
  _page as default
};
