---
description: Propose a file or a text into a shared brain, for the curators' review
argument-hint: <brain> <file or text>
---

Add the document named by the arguments to the shared brain, through the `gc_add` tool of the granular-cortex MCP server.

1. The first argument is the brain (`team/acme`, for example). Call `gc_brains` first when it is missing, and ask which brain to use among those with `write` or `curate` access.
2. The rest is a file path or a text. Read the file when it is a path. Use the file name as `title` and its path as `source_url`.
3. Call `gc_add` with `brain`, `content`, `title`, `source_url` and a stable `document_id` derived from the file name.
4. Report the engine's answer in one line. Say that the document waits for a curator's review, unless the caller is a curator, in which case it is already served.

Arguments: $ARGUMENTS
