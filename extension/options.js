import { settingsOf, authorizeURL, tokenOf, signInText, randomString } from "./lib.js";

const $ = (id) => document.getElementById(id);
const stored = await chrome.storage.local.get(["gateway", "key"]);
$("gateway").value = stored.gateway || "";
$("key").value = stored.key || "";
$("save").addEventListener("click", async () => {
  const s = settingsOf({ gateway: $("gateway").value, key: $("key").value });
  if (!s) {
    $("saved").textContent = "Both fields are required.";
    return;
  }
  await chrome.storage.local.set(s);
  $("saved").textContent = "Saved.";
});

// Sign in: the gateway names the provider, the provider answers with an
// id_token in the fragment of the redirect, the gateway turns it into the
// key, which is saved with the gateway address.
const say = (text, state) => {
  $("signed").textContent = text;
  $("signed").dataset.status = state;
};
$("signin").addEventListener("click", async () => {
  const gateway = String($("gateway").value).trim().replace(/\/+$/, "");
  if (!gateway) {
    say("Give the gateway address first.", "no-gateway");
    return;
  }
  $("signin").disabled = true;
  try {
    let res = await fetch(gateway + "/v1/identity");
    let body = await res.json().catch(() => null);
    if (!res.ok) {
      say(signInText(res.status, body), "refused");
      return;
    }
    const disc = await (await fetch(String(body.issuer).replace(/\/+$/, "") + "/.well-known/openid-configuration")).json();
    const state = randomString(), nonce = randomString();
    const url = authorizeURL({ endpoint: disc.authorization_endpoint, clientID: body.client_id, redirect: chrome.identity.getRedirectURL(), state, nonce });
    const back = await chrome.identity.launchWebAuthFlow({ url, interactive: true });
    const token = tokenOf(back, state, nonce);
    res = await fetch(gateway + "/v1/login", { method: "POST", headers: { Authorization: "Bearer " + token } });
    body = await res.json().catch(() => null);
    if (!res.ok) {
      say(signInText(res.status, body), "refused");
      return;
    }
    await chrome.storage.local.set({ gateway, key: body.key });
    $("key").value = body.key;
    say(signInText(200, body), "signed");
  } catch (e) {
    say(e instanceof TypeError ? signInText(0, null) : e.message, "failed");
  } finally {
    $("signin").disabled = false;
  }
});
