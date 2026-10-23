import test from "node:test";
import assert from "node:assert/strict";
import { writableBrains, documentID, buildAdd, statusText, settingsOf, authorizeURL, tokenOf, signInText, randomString, mainText } from "./lib.js";

// Only the brains the user may write to, with a key, sorted by name.
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
  // How the page text was extracted is said in the metadata; a selection
  // was not extracted.
  assert.deepEqual(buildAdd({ ...page, extract: "landmark" }).metadata, { via: "chrome-extension", kind: "page", extract: "landmark" });
  assert.deepEqual(buildAdd({ ...page, extract: "landmark", selection: "words" }).metadata, { via: "chrome-extension", kind: "selection" });
});

// The page as the collector in the popup reads it: the landmarks, the text
// blocks with the index of their parent (group) and grandparent (outer), the
// length of their link text and the class and id of the parent (hint), and
// the whole body text.
const long = (word, n) => Array.from({ length: n }, (_, i) => `${word} ${i}.`).join(" ");
const block = (text, group, outer, hint = "", links = 0) => ({ text, group, outer, hint, links });

// An article or a main landmark with enough text is the page's content.
test("mainText takes an article or main landmark with enough text", () => {
  const body = "Menu Home Products " + long("Article sentence", 30) + " Footer legal";
  let got = mainText({ landmarks: [{ kind: "main", text: long("Main sentence", 40) }, { kind: "article", text: long("Article sentence", 30) }], blocks: [], body });
  assert.deepEqual(got, { text: long("Article sentence", 30), how: "landmark" });
  // Of two articles, the longer; a main when there is no article.
  got = mainText({ landmarks: [{ kind: "article", text: long("Short", 25) }, { kind: "article", text: long("Longer one", 40) }], blocks: [], body });
  assert.equal(got.text, long("Longer one", 40));
  got = mainText({ landmarks: [{ kind: "main", text: long("Main sentence", 40) }], blocks: [], body });
  assert.deepEqual(got, { text: long("Main sentence", 40), how: "landmark" });
  // A landmark too short to be the content is passed over.
  got = mainText({ landmarks: [{ kind: "article", text: "Share this" }], blocks: [], body });
  assert.deepEqual(got, { text: body, how: "body" });
});

// Without a landmark, the group of blocks with the most text that is not
// links wins; a sibling group (same grandparent) stays in, in page order,
// as the sections of one article do; the navigation, the footer, the
// comments and the link lists stay out.
test("mainText keeps the groups of text blocks and leaves the chrome out", () => {
  const blocks = [
    block("Home Products Pricing Contact", 0, 0, "nav menu", 29),
    block("Cutover plan", 1, 1, "post-content"),
    block(long("The rehearsal", 20), 1, 1, "post-content"),
    block(long("Second section", 8), 2, 1, "section"),
    block(long("Great post, thanks", 30), 3, 1, "comments"),
    block(long("Related link", 10), 4, 1, "", 300),
    block("Copyright 2026 Acme", 5, 0, "site-footer"),
  ];
  const got = mainText({ landmarks: [], blocks, body: "whole body" });
  assert.equal(got.how, "blocks");
  assert.equal(got.text, ["Cutover plan", long("The rehearsal", 20), long("Second section", 8)].join("\n\n"));
  // A group under another grandparent stays out, however close in size.
  const small = mainText({ landmarks: [], blocks: [blocks[1], blocks[2], block(long("Teaser elsewhere", 18), 6, 7, "teaser")], body: "x" });
  assert.equal(small.text, ["Cutover plan", long("The rehearsal", 20)].join("\n\n"));
});

// A content hint in the class or id wins between two groups of close size.
test("mainText prefers the group whose class says content", () => {
  const blocks = [block(long("Teaser text", 22), 0, 0, "teaser"), block(long("Story text", 18), 1, 1, "article-body")];
  assert.equal(mainText({ landmarks: [], blocks, body: "x" }).text, long("Story text", 18));
});

// Too little text anywhere: the whole body, as before; nothing at all is
// an empty text.
test("mainText falls back to the body text", () => {
  assert.deepEqual(mainText({ landmarks: [], blocks: [block("Hello", 0, 0)], body: "Hello there" }), { text: "Hello there", how: "body" });
  assert.deepEqual(mainText({}), { text: "", how: "body" });
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
