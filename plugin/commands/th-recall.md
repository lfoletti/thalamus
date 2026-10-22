---
description: Choose the brains of the automatic recall, local and shared
argument-hint: [brains | on <brain> | off <brain> | only <brain>... | only --clear]
---

Show or change the brains that the automatic recall searches at each prompt. Local brains have the prefix `local:`.

1. Run `"${CLAUDE_PLUGIN_ROOT}/bin/th-recall" brains` when there are no arguments, else `"${CLAUDE_PLUGIN_ROOT}/bin/th-recall" $ARGUMENTS`.
2. After a change, run `"${CLAUDE_PLUGIN_ROOT}/bin/th-recall" brains` and show the list.
3. Say that the choice applies to the automatic recall only. The `th_recall` tool still reads every brain the person may read.

Arguments: $ARGUMENTS
