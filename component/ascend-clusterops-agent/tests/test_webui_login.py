#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""HTTPS-only login behavior and opt-in real-browser acceptance.

Set CLUSTEROPS_WEBUI_URL and CLUSTEROPS_WEBUI_AUTH_JSON for live tests. A private
CLUSTEROPS_WEBUI_AUTH_FILE is an optional fallback. Only a WebUI token is used.
"""

# pylint: disable=redefined-outer-name  # pytest fixtures are intentionally named like the test parameters

import json
import os
from pathlib import Path
import shutil
import subprocess
import pytest

WEBUI = Path(__file__).resolve().parents[1] / "agent-core" / "webui"
BASE = os.environ.get("CLUSTEROPS_WEBUI_URL", "").rstrip("/")
REAL_BROWSER = pytest.mark.skipif(not BASE, reason="set CLUSTEROPS_WEBUI_URL for real HTTPS acceptance")
# The checks need ESM top-level await (Node 14.8+) and node: imports (Node 14.18/16+);
# older Node cannot load them, so skip instead of failing.
MIN_NODE_MAJOR = 16


def _node_major(binary: str) -> int:
    """Node major version; 0 when it cannot be determined."""
    try:
        output = subprocess.run([binary, "--version"], capture_output=True, text=True, check=True).stdout
    except (OSError, subprocess.CalledProcessError):
        return 0
    digits = output.strip().lstrip("v").split(".", 1)[0]
    return int(digits) if digits.isdigit() else 0


def _node(tmp_path, body, *, protocol="https:", stored=" MiX-令牌 ", pathname="/ui/"):
    binary = os.environ.get("CLUSTEROPS_NODE_BINARY") or shutil.which("node")
    if not binary:
        pytest.skip("Node.js is required for JavaScript behavior checks")
    major = _node_major(binary)
    if major < MIN_NODE_MAJOR:
        pytest.skip(
            f"node {major} is too old: the checks need ESM top-level await and node: imports "
            f"(Node {MIN_NODE_MAJOR}+); point CLUSTEROPS_NODE_BINARY at a newer node"
        )
    module = tmp_path / "auth.mjs"
    module.write_text((WEBUI / "auth.js").read_text(encoding="utf-8"), encoding="utf-8")
    setup = (
        """
import assert from 'node:assert/strict';
const requests = [], events = [], saved = new Map();
const elements = Object.fromEntries(['app','login-panel','login-form','webui-token',
  'login-error','login-submit','webui-logout'].map(id => [id, {
    hidden: true, disabled: false, value: '', textContent: '',
    addEventListener(name, callback) { this[name] = callback; }, focus() {},
  }]));
globalThis.document = {
  getElementById: id => elements[id], dispatchEvent: event => events.push(event.type),
  addEventListener(name, callback) { this[name] = callback; },
};
globalThis.CustomEvent = class { constructor(type) { this.type = type; } };
globalThis.location = {protocol: PROTOCOL, origin:'https://example.test', href:'https://example.test'+PATHNAME, pathname:PATHNAME};
globalThis.localStorage = {
  getItem: key => saved.has(key) ? saved.get(key) : null,
  setItem: (key,value) => saved.set(key,value), removeItem: key => saved.delete(key),
};
const replaced = [];
globalThis.history = {replaceState(...args) { replaced.push(args); }};
const original = STORED;
if (original !== null) saved.set('mindcluster.webuiToken.v1', original);
let responseStatus = 200;
let responseBody = {ok:true};
let responseAuth = null;
globalThis.fetch = async (input, options) => {
  requests.push({input, options});
  return {
    status:responseStatus, ok:responseStatus >= 200 && responseStatus < 300,
    headers: {get: name => (name.toLowerCase() === 'x-webui-auth' ? responseAuth : null)},
    json: async () => responseBody,
  };
};
const auth = await import(process.argv[2]);
""".replace("PROTOCOL", json.dumps(protocol))
        .replace("STORED", json.dumps(stored))
        .replace("PATHNAME", json.dumps(pathname))
    )
    runner = tmp_path / "check.mjs"
    runner.write_text(setup + body, encoding="utf-8")
    subprocess.run([binary, str(runner), module.as_uri()], capture_output=True, text=True, check=True, timeout=20)


def test_exact_token_authentication_remember_and_auth_failure_reset(tmp_path):
    _node(
        tmp_path,
        """
await auth.initializeAuth();
assert.equal(auth.isAuthenticated(), true);
assert.equal(requests[0].input, '/api/v1/auth/check');
assert.equal(requests[0].options.method, 'POST');
const headers = requests[0].options.headers;
assert.equal(headers.get('X-WebUI-Token-Encoding'), 'base64');
assert.equal(Buffer.from(headers.get('Authorization').slice(7), 'base64').toString('utf8'), original);
assert.equal(saved.get('mindcluster.webuiToken.v1'), original);
await auth.authFetch('/diag', {method:'POST'});
assert.equal(requests[1].options.headers.get('Authorization'), headers.get('Authorization'));
responseAuth = 'unauthenticated';
await assert.rejects(auth.authFetch('/api/v1/auth/check'), error => error.code === 'UNAUTHENTICATED');
assert.equal(auth.isAuthenticated(), false);
assert.equal(saved.has('mindcluster.webuiToken.v1'), false);
assert.equal(elements['app'].hidden, true);
assert.equal(elements['login-panel'].hidden, false);
assert.ok(events.includes('webui-auth-required'));
""",
    )


def test_logout_click_clears_the_saved_token(tmp_path):
    """The logout button arrives with the swapped-in application page, after initializeAuth."""
    _node(
        tmp_path,
        """
await auth.initializeAuth();
assert.equal(auth.isAuthenticated(), true);
assert.equal(typeof document.click, 'function');
document.click({target: {id: 'webui-logout'}});
assert.equal(auth.isAuthenticated(), false);
assert.equal(saved.has('mindcluster.webuiToken.v1'), false);
assert.equal(elements['login-panel'].hidden, false);
assert.ok(events.includes('webui-auth-required'));
""",
    )


def test_login_route_is_canonicalized_after_authentication(tmp_path):
    _node(
        tmp_path,
        """
await auth.initializeAuth();
assert.equal(auth.isAuthenticated(), true);
assert.equal(replaced.length, 1);
assert.equal(replaced[0][2], '/ui/');
""",
        pathname="/ui/login",
    )


def test_ui_entry_path_is_not_rewritten(tmp_path):
    _node(
        tmp_path,
        """
await auth.initializeAuth();
assert.equal(auth.isAuthenticated(), true);
assert.equal(replaced.length, 0);
""",
        pathname="/ui/",
    )


def test_token_never_sent_to_another_origin(tmp_path):
    _node(
        tmp_path,
        """
await auth.initializeAuth();
const before = requests.length;
await assert.rejects(auth.authFetch('https://other.test/diag'));
assert.equal(requests.length, before);
""",
    )


@pytest.mark.parametrize("value", ["", "x", " ABCD-efgh ", "令牌😀"])
def test_utf8_encoding_does_not_normalize_tokens(tmp_path, value):
    _node(
        tmp_path,
        """
const value = VALUE;
assert.equal(Buffer.from(auth.encodeToken(value), 'base64').toString('utf8'), value);
""".replace("VALUE", json.dumps(value)),
    )


def test_http_page_never_sends_saved_token(tmp_path):
    _node(
        tmp_path,
        """
await auth.initializeAuth();
assert.equal(auth.isAuthenticated(), false);
assert.equal(requests.length, 0);
assert.equal(elements['login-submit'].disabled, true);
await assert.rejects(auth.authFetch('/diag'));
assert.equal(requests.length, 0);
""",
        protocol="http:",
    )


def _auth():
    __tracebackhide__ = True
    try:
        raw = os.environ.get("CLUSTEROPS_WEBUI_AUTH_JSON")
        if raw is None:
            raw = Path(os.environ["CLUSTEROPS_WEBUI_AUTH_FILE"]).read_text(encoding="utf-8")
        data = json.loads(raw)
        valid = isinstance(data.get("token"), str)
    except Exception:
        data, valid = None, False
    if not valid:
        pytest.fail("Missing private browser authentication configuration", pytrace=False)
    return data


@pytest.fixture(scope="module")
def browser():
    driver = pytest.importorskip("playwright.sync_api")
    options = {"headless": True, "args": ["--no-proxy-server"]}
    if os.environ.get("CLUSTEROPS_BROWSER_PATH"):
        options["executable_path"] = os.environ["CLUSTEROPS_BROWSER_PATH"]
    with driver.sync_playwright() as playwright:
        instance = playwright.chromium.launch(**options)
        yield instance
        instance.close()


def _login(page):
    __tracebackhide__ = True
    failed = False
    try:
        page.locator("#webui-token").fill(_auth()["token"])
        page.locator("#login-submit").click()
        page.locator("#app").wait_for(state="visible", timeout=15000)
    except Exception:
        failed = True
    if failed:
        pytest.fail("HTTPS browser login failed", pytrace=False)


@REAL_BROWSER
def test_https_login_wrong_correct_remember_logout_and_rejected_token(browser):
    context = browser.new_context(ignore_https_errors=True)
    page = context.new_page()
    try:
        page.goto(BASE + "/ui/", wait_until="domcontentloaded")
        assert page.locator("#login-panel").is_visible()
        page.locator("#webui-token").fill("wrong-test-token")
        with page.expect_response(lambda response: response.url.endswith("/api/v1/auth/check")) as rejected:
            page.locator("#login-submit").click()
        assert rejected.value.status == 200
        assert rejected.value.headers.get("x-webui-auth") == "unauthenticated"
        assert page.locator("#login-panel").is_visible()
        _login(page)
        page.reload(wait_until="domcontentloaded")
        page.locator("#app").wait_for(state="visible", timeout=15000)
        page.locator("#webui-logout").click()
        # Logging out leaves the application page for the whitelisted login shell.
        page.locator("#login-panel").wait_for(state="visible", timeout=15000)
        assert page.evaluate("localStorage.getItem('mindcluster.webuiToken.v1') === null")
    finally:
        context.close()


@REAL_BROWSER
def test_untrusted_self_signed_certificate_needs_browser_trust(browser):
    certificate = _auth().get("certificate", "")
    if certificate:
        from cryptography import x509

        raw = certificate.encode() if "BEGIN CERTIFICATE" in certificate else Path(certificate).read_bytes()
        parsed = x509.load_pem_x509_certificate(raw)
        if parsed.issuer != parsed.subject:
            pytest.skip("certificate is not self-signed")
    context = browser.new_context(ignore_https_errors=False)
    page = context.new_page()
    message = ""
    try:
        try:
            page.goto(BASE + "/ui/", timeout=15000)
        except Exception as error:
            message = str(error)
        assert "ERR_CERT_AUTHORITY_INVALID" in message
    finally:
        context.close()
