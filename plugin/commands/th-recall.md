---
description: Choose the brains of the automatic recall, local and shared
argument-hint: [brains | on <brain> | off <brain> | only <brain>... | only --clear]
---

Show or change the brains that the automatic recall searches at each prompt. Local brains have the prefix `local:`.

With arguments:

1. Run `"${CLAUDE_PLUGIN_ROOT}/bin/th-recall" $ARGUMENTS`.
2. Run `"${CLAUDE_PLUGIN_ROOT}/bin/th-recall" brains` and show the list.

Without arguments, show a checklist. The `/th-brains` command shows the same list in a pane, where Enter turns a brain on or off; mention it once.

1. Run `"${CLAUDE_PLUGIN_ROOT}/bin/th-recall" brains --json`. Each brain has `access`, `on`, and `ok` with a `reason` when it does not answer.
2. Ask the person with the AskUserQuestion tool, with `multiSelect: true`. Put 4 brains at most in one question, and use more questions for more brains. Group the brains by the part of the name before `/`, and give each question a short header from that group. The label of each option is the name of a brain. Its description is "now on" or "now off", and the reason when the brain does not answer.
3. If more than 16 brains are listed, do not ask. Show the list, and say that `/th-recall on <brain>` and `/th-recall off <brain>` change it.
4. If the person chooses no brain, keep the choice of brains as it is, skip steps 5 and 6, and say so.
5. If the person chooses every brain, run `"${CLAUDE_PLUGIN_ROOT}/bin/th-recall" only --clear`. Else run `"${CLAUDE_PLUGIN_ROOT}/bin/th-recall" only` with the chosen names.
6. Run `"${CLAUDE_PLUGIN_ROOT}/bin/th-recall" on` with the chosen names, which takes them out of the skipped brains.
7. Run `"${CLAUDE_PLUGIN_ROOT}/bin/th-recall" brains` and show the list.

In both cases, say that the choice applies to the automatic recall only. The `th_recall` tool still reads every brain the person may read.

Arguments: $ARGUMENTS
