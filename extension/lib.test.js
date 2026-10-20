import test from "node:test";
import assert from "node:assert/strict";
import { writableBrains, documentID, buildAdd, statusText, settingsOf, authorizeURL, tokenOf, signInText, randomString } from "./lib.js";

// Only the brains the employee may write to, with a key, sorted by name.
test("writableBrains keeps the brains one may write to, with a key, sorted", () => {
  const got = writableBrains([
    { brain: "team/zeta", access: "write", key: true },
    { brain: "org/secrets", access: "read", key: true },
    { brain: "team/acme", access: "curate", key: true },
    { brain: "team/nokey", access: "write", key: false },
  ]);
  assert.deepEqual(got.map((b) => b.brain), ["team/acme", "team/zeta"]);
  assert.deepEqual(writableBrains([]), []);
});

// The document id follows the page: stable, without the fragment, and
// distinct for two pages.
test("documentID is stable per page and ignores the fragment", () => {
  const a = documentID("https://intranet/spec?v=2#top");
  assert.equal(a, documentID("https://intranet/spec?v=2"));
  assert.notEqual(a, documentID("https://intranet/spec?v=3"));
  assert.match(a, /^web-[0-9a-f]{8}$/);
});

// The body of POST /v1/add: the selection wins over the page text, the
// whitespace is collapsed, the title and the content are bounded, and the
// metadata says where it comes from. Nothing to propose is an error.
test("buildAdd prefers the selection and bounds the fields", () => {
  const page = { brain: "team/acme", title: "  Spec  ", content: "The  page\n\n text", url: "https://intranet/spec#x", selection: "" };
  let got = buildAdd(page);
  assert.equal(got.brain, "team/acme");
  assert.equal(got.title, "Spec");
  assert.equal(got.content, "The page text");
  assert.equal(got.source_url, "https://intranet/spec#x");
  assert.equal(got.document_id, documentID(page.url));
  assert.deepEqual(got.metadata, { via: "chrome-extension", kind: "page" });
  got = buildAdd({ ...page, selection: " the  chosen words " });
  assert.equal(got.content, "the chosen words");
  assert.equal(got.metadata.kind, "selection");
  assert.equal(buildAdd({ ...page, title: "t".repeat(300) }).title.length, 200);
  assert.equal(buildAdd({ ...page, content: "c".repeat(200000) }).content.length, 100000);
  assert.throws(() => buildAdd({ ...page, content: "  ", selection: "" }), /nothing to propose/);
  assert.throws(() => buildAdd({ ...page, brain: "" }), /choose a brain/);
});

// One sentence per outcome, by status, with the brain and the access.
test("statusText says what happened, by status", () => {
  assert.equal(statusText(200, {}, "team/acme", "write"), "Proposed to team/acme: it waits for a curator.");
  assert.equal(statusText(200, {}, "team/acme", "curate"), "Added to team/acme: you curate it, so it is served at once.");
  assert.equal(statusText(401, { error: "refused: unknown key" }), "Your key is unknown to the gateway: check the options.");
  assert.equal(statusText(403, { error: "th_add: x may not write to team/acme" }), "You may not write to this brain.");
  assert.equal(statusText(404, { error: "th_add: unknown brain" }), "Unknown brain: th_add: unknown brain");
  assert.equal(statusText(400, { error: "th_add: content is required" }), "Refused: th_add: content is required");
  assert.equal(statusText(502, { error: "th_add: brain says no" }), "The brain refused: th_add: brain says no");
  assert.equal(statusText(0, null), "The gateway did not answer: check the address in the options.");
  assert.equal(statusText(500, { error: "key file of team/acme: broken" }), "Error 500: key file of team/acme: broken");
  assert.equal(statusText(500, "not json"), "Error 500: not json");
});

// The settings: both fields trimmed and required, the gateway without a
// trailing slash.
test("settingsOf trims, requires both fields and drops the trailing slash", () => {
  assert.deepEqual(settingsOf({ gateway: " http://gw:8787/ ", key: " k " }), { gateway: "http://gw:8787", key: "k" });
  assert.equal(settingsOf({ gateway: "", key: "k" }), null);
  assert.equal(settingsOf({ gateway: "http://gw", key: " " }), null);
  assert.equal(settingsOf({}), null);
});

// The sign-in address of the provider, for the implicit id_token flow of
// the extension: every parameter the provider needs, the fragment mode.
test("authorizeURL carries the id_token request", () => {
  const u = new URL(authorizeURL({ endpoint: "https://login.example/authorize?tenant=t", clientID: "thalamus-ext", redirect: "https://abc.chromiumapp.org/", state: "s1", nonce: "n1" }));
  assert.equal(u.origin + u.pathname, "https://login.example/authorize");
  assert.equal(u.searchParams.get("tenant"), "t");
  assert.equal(u.searchParams.get("response_type"), "id_token");
  assert.equal(u.searchParams.get("response_mode"), "fragment");
  assert.equal(u.searchParams.get("scope"), "openid email");
  assert.equal(u.searchParams.get("client_id"), "thalamus-ext");
  assert.equal(u.searchParams.get("redirect_uri"), "https://abc.chromiumapp.org/");
  assert.equal(u.searchParams.get("state"), "s1");
  assert.equal(u.searchParams.get("nonce"), "n1");
});

// The id_token comes back in the fragment: the state must be ours, the
// nonce must be ours, a refusal is said.
test("tokenOf reads the id_token from the fragment and checks state and nonce", () => {
  const tok = "h." + Buffer.from(JSON.stringify({ nonce: "n1" })).toString("base64url") + ".s";
  assert.equal(tokenOf("https://abc.chromiumapp.org/#id_token=" + tok + "&state=s1", "s1", "n1"), tok);
  assert.throws(() => tokenOf("https://abc.chromiumapp.org/#id_token=" + tok + "&state=other", "s1", "n1"), /state/);
  assert.throws(() => tokenOf("https://abc.chromiumapp.org/#id_token=" + tok + "&state=s1", "s1", "n2"), /nonce/);
  assert.throws(() => tokenOf("https://abc.chromiumapp.org/#error=access_denied&state=s1", "s1", "n1"), /refused: access_denied/);
  assert.throws(() => tokenOf("https://abc.chromiumapp.org/#state=s1", "s1", "n1"), /no id_token/);
  assert.throws(() => tokenOf("https://abc.chromiumapp.org/#id_token=not-a-jwt&state=s1", "s1", "n1"), /nonce/);
  assert.throws(() => tokenOf(undefined, "s1", "n1"), /no answer/);
});

// One sentence after the sign-in, by status.
test("signInText says what happened, by status", () => {
  assert.equal(signInText(200, { login: "lfoletti", brains: [{}, {}] }), "Signed in as lfoletti: 2 brain(s). Your key is saved.");
  assert.equal(signInText(404, { error: "no identity provider is configured" }), "This gateway has no identity provider: paste your key instead.");
  assert.equal(signInText(401, { error: "token: expired" }), "The gateway refused the sign-in: token: expired");
  assert.equal(signInText(403, { error: "x has no brain" }), "You have no brain in the rights: ask the administrator.");
  assert.equal(signInText(0, null), "The gateway did not answer: check the address.");
  assert.equal(signInText(500, { error: "key file: broken" }), "Error 500: key file: broken");
});

// A random string is long, URL-safe and new each time.
test("randomString is long and URL-safe", () => {
  const a = randomString(), b = randomString();
  assert.match(a, /^[A-Za-z0-9_-]{32,}$/);
  assert.notEqual(a, b);
});
