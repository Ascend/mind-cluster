/* Copyright 2026 Huawei Technologies Co., Ltd.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
// The browser stores only the shared WebUI token; model API keys never reach it.
const STORAGE_KEY = "mindcluster.webuiToken.v1";
// Server-side entries shared with login.js: the explicit login route and the protected app shell.
export const LOGIN_ENTRY = "/ui/login";
export const APP_PAGE = "/ui/index.html";
export const APP_MODULE = "/ui/app.js";
export const APP_ENTRY = "/ui/";
// Server contract: a valid path with bad credentials returns 200 and explains why in this header and body.
const AUTH_RESULT_HEADER = "x-webui-auth";
const AUTH_RESULT_UNAUTHENTICATED = "unauthenticated";
const nativeFetch = typeof globalThis.fetch === "function" ? globalThis.fetch.bind(globalThis) : null;
const secureEntry = typeof globalThis.location !== "undefined" && location.protocol === "https:";
let token = null;
let authenticated = false;
let loginFormBound = false;
const pendingLoginResolvers = [];

export function isAuthenticated() { return authenticated; }

// ---------------- Page elements and view switching ----------------

function loginElements() {
  if (typeof document === "undefined") { return {}; }
  return {
    app: document.getElementById("app"), panel: document.getElementById("login-panel"),
    form: document.getElementById("login-form"), input: document.getElementById("webui-token"),
    error: document.getElementById("login-error"), submit: document.getElementById("login-submit"),
    logout: document.getElementById("webui-logout"),
  };
}

function emitAuthEvent(name) {
  if (typeof document !== "undefined") { document.dispatchEvent(new CustomEvent(name)); }
}

function showLoginForm(message = "") {
  const ui = loginElements();
  if (ui.app) { ui.app.hidden = true; }
  if (ui.panel) { ui.panel.hidden = false; }
  if (ui.logout) { ui.logout.hidden = true; }
  if (ui.error) { ui.error.textContent = message; }
}

function showApplicationPage() {
  const ui = loginElements();
  if (ui.app) { ui.app.hidden = false; }
  if (ui.panel) { ui.panel.hidden = true; }
  if (ui.logout) { ui.logout.hidden = !secureEntry; }
  if (ui.input) { ui.input.value = ""; }
  if (ui.error) { ui.error.textContent = ""; }
}

// ---------------- Token transport and storage ----------------

// Headers trim surrounding whitespace and cannot carry arbitrary Unicode, so the token
// travels as UTF-8 + base64; the server compares the raw bytes.
export function encodeToken(value) {
  const bytes = new TextEncoder().encode(value);
  let binary = "";
  for (const byte of bytes) { binary += String.fromCharCode(byte); }
  return btoa(binary);
}

function buildAuthHeaders(value, initial) {
  const headers = new Headers(initial);
  headers.set("Authorization", "Bearer " + encodeToken(value));
  headers.set("X-WebUI-Token-Encoding", "base64");
  return headers;
}

function readStoredToken() {
  try { return localStorage.getItem(STORAGE_KEY); } catch (error) { return null; }
}

function saveStoredToken(value) {
  try { localStorage.setItem(STORAGE_KEY, value); } catch (error) { /* private mode only affects the next visit */ }
}

function clearStoredToken() {
  try { localStorage.removeItem(STORAGE_KEY); } catch (error) { /* unavailable storage must not block logout */ }
}

// ---------------- Authentication state transitions ----------------

function requireLogin(message) {
  token = null;
  authenticated = false;
  clearStoredToken();
  emitAuthEvent("webui-auth-required");
  if (!loginElements().panel) {
    // The application page has no login form: fall back to the login shell, which needs no Authorization header.
    if (typeof location !== "undefined" && typeof location.replace === "function") {
      location.replace(LOGIN_ENTRY);
    }
    return;
  }
  showLoginForm(message || "登录已失效，请重新输入共享 token。");
}

// The logout button belongs to the application page, which is swapped in after the login
// page initializes; bind it at document level so it works as soon as it appears.
function handleLogoutClick(event) {
  const target = event && event.target;
  if (target && target.id === "webui-logout") { requireLogin("已退出登录，请重新输入共享 token。"); }
}

function bindLogoutDelegation() {
  if (typeof document !== "undefined" && typeof document.addEventListener === "function") {
    document.addEventListener("click", handleLogoutClick);
  }
}

function canonicalizeEntryPath() {
  // After login the address bar shows /ui/ and drops the login path.
  try {
    if (typeof location === "undefined" || typeof history === "undefined" || !history.replaceState) { return; }
    if (location.pathname === LOGIN_ENTRY || location.pathname === LOGIN_ENTRY + ".html" || location.pathname === APP_PAGE) {
      history.replaceState(null, "", APP_ENTRY);
    }
  } catch (error) { /* a failed URL rewrite must not break login */ }
}

function markAuthenticated(value) {
  token = value;
  authenticated = true;
  saveStoredToken(value);
  showApplicationPage();
  canonicalizeEntryPath();
  while (pendingLoginResolvers.length) { pendingLoginResolvers.shift()(); }
  emitAuthEvent("webui-authenticated");
}

function waitForLogin() {
  return new Promise((resolve) => pendingLoginResolvers.push(resolve));
}

// ---------------- Server-side verification ----------------

async function verifyTokenWithServer(value) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 10000);
  try {
    const response = await nativeFetch("/api/v1/auth/check", {
      method: "POST", headers: buildAuthHeaders(value), signal: controller.signal, cache: "no-store",
    });
    if (response.status !== 200) { throw new Error("登录验证暂不可用，请稍后重试。"); }
    const payload = await response.json();
    return Boolean(payload && payload.ok === true);
  } finally { clearTimeout(timer); }
}

async function handleLoginSubmit(event) {
  event.preventDefault();
  const ui = loginElements();
  if (ui.submit.disabled) { return; }
  const value = ui.input.value;
  ui.submit.disabled = true;
  ui.error.textContent = "正在验证…";
  try {
    if (await verifyTokenWithServer(value)) {
      markAuthenticated(value);
    } else {
      token = null;
      authenticated = false;
      clearStoredToken();
      ui.input.value = "";
      showLoginForm("token 无效或已轮换，请重新输入。");
      ui.input.focus();
    }
  } catch (error) {
    showLoginForm("暂时无法完成登录验证，请检查连接后重试。");
  } finally { ui.submit.disabled = false; }
}

function bindLoginForm() {
  if (loginFormBound) { return; }
  const ui = loginElements();
  if (!ui.form || !ui.input) { throw new Error("登录页面未正确加载"); }
  ui.form.addEventListener("submit", handleLoginSubmit);
  loginFormBound = true;
}

async function restoreStoredToken(ui, saved) {
  ui.input.disabled = true;
  ui.submit.disabled = true;
  try {
    if (await verifyTokenWithServer(saved)) {
      markAuthenticated(saved);
      return true;
    }
    clearStoredToken();
    token = null;
    showLoginForm("登录已失效，请重新输入共享 token。");
  } catch (error) {
    showLoginForm("暂时无法验证已保存的登录信息，请检查连接后重试。");
  } finally {
    ui.input.disabled = false;
    ui.submit.disabled = false;
  }
  return false;
}

export async function initializeAuth() {
  if (!secureEntry) {
    showLoginForm("请通过 HTTPS 入口打开本页面。");
    const ui = loginElements();
    if (ui.submit) { ui.submit.disabled = true; }
    return false;
  }
  const ui = loginElements();
  bindLoginForm();
  showLoginForm();
  const saved = readStoredToken();
  if (saved !== null) {
    token = saved;
    if (await restoreStoredToken(ui, saved)) { return; }
  }
  return waitForLogin();
}

// ---------------- Business request wrapper ----------------

function assertSameOrigin(input) {
  const target = new URL(typeof input === "string" ? input : input instanceof URL ? input.href : input.url, location.href);
  if (target.origin !== location.origin) { throw new Error("认证请求必须使用当前站点"); }
}

function unauthorizedError(message) {
  const error = new Error(message);
  error.code = "UNAUTHENTICATED";
  return error;
}

/** Attach the shared token to same-origin requests; on auth failure clear it and return to login. */
export async function authFetch(input, options = {}) {
  if (!nativeFetch) { throw new Error("当前环境没有可用的 fetch"); }
  if (!secureEntry) { throw new Error("认证请求必须通过 HTTPS 发送。"); }
  assertSameOrigin(input);
  if (!authenticated || token === null) {
    requireLogin("请先输入共享 token 登录。");
    throw unauthorizedError("需要重新登录");
  }
  const response = await nativeFetch(input, {
    ...options, headers: buildAuthHeaders(token, options.headers), credentials: "same-origin",
  });
  if (response.headers.get(AUTH_RESULT_HEADER) === AUTH_RESULT_UNAUTHENTICATED) {
    requireLogin("登录已失效或 token 已轮换，请重新登录。");
    throw unauthorizedError("登录已失效，请重新登录");
  }
  return response;
}

bindLogoutDelegation();
