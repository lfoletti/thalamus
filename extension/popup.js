import { writableBrains, buildAdd, statusText, settingsOf, mainText } from "./lib.js";

const $ = (id) => document.getElementById(id);
const say = (text, state) => {
  $("status").textContent = text;
  $("status").dataset.status = state;
};

// The page to propose: the active tab of the window, never an extension
// page (a test opens this popup as a tab of its own).
async function pageTab() {
  const tabs = await chrome.tabs.query({ currentWindow: true });
  const web = tabs.filter((t) => /^https?:/.test(t.url || ""));
  return web.find((t) => t.active) || web[web.length - 1] || null;
}

const settings = settingsOf(await chrome.storage.local.get(["gateway", "key"]));
if (!settings) {
  say("Set the gateway address and your key in the options first.", "no-settings");
  const a = document.createElement("a");
  a.href = chrome.runtime.getURL("options.html");
  a.target = "_blank";
  a.textContent = "Open the options";
  $("status").append(" ", a);
} else {
  const tab = await pageTab();
  if (!tab) {
    say("No web page to propose in this window.", "no-page");
  } else {
    $("title").value = tab.title || "";
    try {
      // The page only collects; mainText (lib.js) decides what the content
      // is. The function runs in the page, so it is self-contained.
      const [{ result }] = await chrome.scripting.executeScript({
        target: { tabId: tab.id },
        func: () => {
          const text = (el) => (el && el.innerText) || "";
          const ids = new Map();
          const idOf = (el) => {
            if (!el) return -1;
            if (!ids.has(el)) ids.set(el, ids.size);
            return ids.get(el);
          };
          const hintOf = (el) => (el ? [el.tagName, el.getAttribute("class"), el.id].join(" ") : "");
          const blockTags = "p, pre, blockquote, li, h1, h2, h3, h4, h5, h6";
          const landmarks = [...document.querySelectorAll("article")].map((el) => ({ kind: "article", text: text(el) }))
            .concat([...document.querySelectorAll("main, [role=main]")].map((el) => ({ kind: "main", text: text(el) })));
          const blocks = [...document.querySelectorAll(blockTags)]
            .filter((el) => !el.parentElement.closest(blockTags))
            .map((el) => {
              const parent = el.parentElement, grand = parent.parentElement;
              return {
                text: text(el),
                group: idOf(parent),
                outer: idOf(grand),
                hint: hintOf(parent) + " " + hintOf(grand),
                links: [...el.querySelectorAll("a")].reduce((n, a) => n + text(a).length, 0),
              };
            });
          return { selection: String(getSelection()), page: { landmarks, blocks, body: text(document.body) } };
        },
      });
      if ((result.selection || "").trim()) {
        $("content").value = result.selection.trim();
        $("content").dataset.kind = "selection";
      } else {
        const main = mainText(result.page);
        $("content").value = main.text;
        $("content").dataset.kind = "page";
        $("content").dataset.extract = main.how;
      }
    } catch (e) {
      say(`Cannot read this page: ${e.message}`, "no-text");
    }
    let brains = [];
    try {
      const res = await fetch(settings.gateway + "/v1/brains", { headers: { Authorization: "Bearer " + settings.key } });
      const body = await res.json().catch(() => null);
      if (!res.ok) {
        say(statusText(res.status, body), "brains-refused");
      } else {
        brains = writableBrains(body);
      }
    } catch (e) {
      say(statusText(0, null), "no-gateway");
    }
    for (const b of brains) {
      const o = document.createElement("option");
      o.value = b.brain;
      o.textContent = `${b.brain} (${b.access})`;
      o.dataset.access = b.access;
      $("brain").append(o);
    }
    if (brains.length === 0 && !$("status").dataset.status) {
      say("You may not write to any brain.", "no-brain");
    }
    $("propose").disabled = brains.length === 0;
    $("propose").addEventListener("click", async () => {
      $("propose").disabled = true;
      let body;
      try {
        body = buildAdd({
          brain: $("brain").value,
          title: $("title").value,
          content: $("content").value,
          url: tab.url,
          selection: $("content").dataset.kind === "selection" ? $("content").value : "",
          extract: $("content").dataset.extract,
        });
      } catch (e) {
        say(e.message, "invalid");
        $("propose").disabled = false;
        return;
      }
      const access = $("brain").selectedOptions[0]?.dataset.access;
      try {
        const res = await fetch(settings.gateway + "/v1/add", {
          method: "POST",
          headers: { Authorization: "Bearer " + settings.key, "Content-Type": "application/json" },
          body: JSON.stringify(body),
        });
        const answer = await res.json().catch(() => null);
        say(statusText(res.status, answer, body.brain, access), res.ok ? "sent" : "refused");
      } catch (e) {
        say(statusText(0, null), "no-gateway");
      }
      $("propose").disabled = false;
    });
  }
}
