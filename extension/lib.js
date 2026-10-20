// Pure functions of the Thalamus extension: what to send and what to say.
// No DOM, no chrome.* here, so that node --test covers every rule.

// writableBrains keeps the brains the employee may write to (write or
// curate), with a key issued, sorted by name.
export function writableBrains(brains) {
  return (brains || [])
    .filter((b) => (b.access === "write" || b.access === "curate") && b.key)
    .sort((a, b) => (a.brain < b.brain ? -1 : a.brain > b.brain ? 1 : 0));
}

// documentID derives a stable id from the page address, without the
// fragment: proposing the same page again updates the same document.
export function documentID(url) {
  const s = String(url).split("#")[0];
  let h = 0x811c9dc5;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 0x01000193) >>> 0;
  }
  return "web-" + h.toString(16).padStart(8, "0");
}

const collapse = (s) => String(s || "").replace(/\s+/g, " ").trim();

// buildAdd is the body of POST /v1/add: the selection when there is one,
// the page text otherwise, bounded, with the address as source.
export function buildAdd({ brain, title, content, url, selection }) {
  if (!brain) throw new Error("choose a brain");
  const sel = collapse(selection);
  const text = (sel || collapse(content)).slice(0, 100000);
  if (!text) throw new Error("nothing to propose: the page has no text and nothing is selected");
  return {
    brain,
    document_id: documentID(url),
    title: collapse(title).slice(0, 200),
    content: text,
    source_url: String(url),
    metadata: { via: "chrome-extension", kind: sel ? "selection" : "page" },
  };
}

// statusText is the one sentence shown after a proposal, by HTTP status.
export function statusText(status, body, brain, access) {
  const err = body && typeof body === "object" && body.error ? body.error : typeof body === "string" ? body : "";
  switch (status) {
    case 200:
      return access === "curate"
        ? `Added to ${brain}: you curate it, so it is served at once.`
        : `Proposed to ${brain}: it waits for a curator.`;
    case 0:
      return "The gateway did not answer: check the address in the options.";
    case 400:
      return `Refused: ${err}`;
    case 401:
      return "Your key is unknown to the gateway: check the options.";
    case 403:
      return "You may not write to this brain.";
    case 404:
      return `Unknown brain: ${err}`;
    case 502:
      return `The brain refused: ${err}`;
    default:
      return `Error ${status}: ${err}`;
  }
}

// settingsOf reads the two settings, trimmed; null when one is missing.
export function settingsOf(stored) {
  const gateway = String((stored && stored.gateway) || "").trim().replace(/\/+$/, "");
  const key = String((stored && stored.key) || "").trim();
  return gateway && key ? { gateway, key } : null;
}

// authorizeURL is the provider's sign-in address for the implicit flow:
// the id_token comes back in the fragment of the redirect.
export function authorizeURL({ endpoint, clientID, redirect, state, nonce }) {
  const u = new URL(endpoint);
  for (const [k, v] of Object.entries({ response_type: "id_token", response_mode: "fragment", scope: "openid email", client_id: clientID, redirect_uri: redirect, state, nonce })) {
    u.searchParams.set(k, v);
  }
  return u.toString();
}

function nonceOf(token) {
  try {
    return JSON.parse(atob(String(token).split(".")[1].replace(/-/g, "+").replace(/_/g, "/"))).nonce;
  } catch {
    return undefined;
  }
}

// tokenOf reads the id_token from the redirect the provider made, and
// checks that it answers this sign-in (state) and this request (nonce).
export function tokenOf(redirectURL, state, nonce) {
  if (!redirectURL) throw new Error("the provider gave no answer: the sign-in window was closed");
  const p = new URLSearchParams(String(redirectURL).split("#")[1] || "");
  if (p.get("error")) throw new Error("the provider refused: " + p.get("error"));
  if (p.get("state") !== state) throw new Error("the state of the answer is not this sign-in's");
  const token = p.get("id_token");
  if (!token) throw new Error("no id_token in the provider's answer");
  if (nonceOf(token) !== nonce) throw new Error("the id_token does not carry the nonce of this sign-in");
  return token;
}

// signInText is the one sentence shown after a sign-in, by HTTP status.
export function signInText(status, body) {
  const err = body && body.error ? body.error : "";
  switch (status) {
    case 200:
      return `Signed in as ${body.login}: ${body.brains.length} brain(s). Your key is saved.`;
    case 0:
      return "The gateway did not answer: check the address.";
    case 401:
      return `The gateway refused the sign-in: ${err}`;
    case 403:
      return "You have no brain in the rights: ask the administrator.";
    case 404:
      return "This gateway has no identity provider: paste your key instead.";
    default:
      return `Error ${status}: ${err}`;
  }
}

// randomString is a state or a nonce: 32 bytes, URL-safe.
export function randomString() {
  const b = new Uint8Array(32);
  crypto.getRandomValues(b);
  return btoa(String.fromCharCode(...b)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}
