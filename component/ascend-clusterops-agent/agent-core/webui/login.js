// Copyright 2026 Huawei Technologies Co., Ltd. Licensed under the Apache License, Version 2.0.
import { authFetch, initializeAuth, APP_PAGE, APP_MODULE, APP_ENTRY } from "./auth.js";

// The application shell lives in index.html, which is protected: a browser navigation
// cannot carry an Authorization header, so the shell is fetched with the verified token and
// its #app is swapped into this document instead of navigating to /ui/. The auth.js module
// instance stays alive, so app.js starts already authenticated.
async function fetchApplicationShell() {
  const response = await authFetch(APP_PAGE);
  return new DOMParser().parseFromString(await response.text(), "text/html");
}

function readBootState(shell) {
  // The first-paint state is inlined in the page and reused here so the model badge and
  // placeholder match the server. DOMParser never runs scripts (the CSP only allows self),
  // so only the string value is read. The payload is taken whole (it may be nested) and a
  // malformed value only costs the first-paint state, never the page swap.
  for (const script of shell.querySelectorAll("script")) {
    const text = script.textContent || "";
    const marker = "window.__AGENT_BOOT__";
    const at = text.indexOf(marker);
    if (at < 0) { continue; }
    const eq = text.indexOf("=", at + marker.length);
    if (eq < 0) { continue; }
    const payload = text.slice(eq + 1).trim().replace(/;\s*$/, "");
    try {
      window.__AGENT_BOOT__ = JSON.parse(payload);
    } catch (error) {
      console.warn("AGENT_BOOT payload is not valid JSON; the first-paint state is unavailable", error);
    }
  }
}

function mountApplicationPage(shell) {
  const app = shell.getElementById("app");
  if (!app) { throw new Error("index.html carries no #app container"); }
  for (const link of shell.querySelectorAll('link[rel="stylesheet"]')) {
    document.head.append(link.cloneNode());
  }
  document.title = shell.title || document.title;
  document.body.replaceChildren(app);
}

async function loadApplicationModule() {
  // The renderer is self-contained (no bundled third-party code), so one module import is enough.
  await import(APP_MODULE);
}

async function enterApplication() {
  const shell = await fetchApplicationShell();
  readBootState(shell);
  mountApplicationPage(shell);
  await loadApplicationModule();
  try { history.replaceState(null, "", APP_ENTRY); } catch (error) { /* URL cleanup */ }
}

function showApplicationLoadFailure() {
  const error = document.getElementById("login-error");
  if (error) { error.textContent = "登录成功，但页面加载失败，请刷新重试。"; }
}

// The token is already stored, so a refresh re-verifies it and swaps the shell again.
document.addEventListener("webui-authenticated", () => {
  enterApplication().catch(showApplicationLoadFailure);
});
initializeAuth().catch(showApplicationLoadFailure);
