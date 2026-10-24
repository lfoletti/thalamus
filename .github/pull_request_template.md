## What

<!-- What the change does now, in a few short sentences. -->

## Why

<!-- The need, or the issue it closes (Closes #N). -->

## Tests

<!-- For each new rule: the test that covers it, and the hand mutation in gateway/mutate.py that makes it fail. -->

| Rule | Test | Mutation |
| :- | :- | :- |
|  |  |  |

## Failing first

<!-- The command, and the FAIL lines it prints without the change. -->

## Checklist

- [ ] The tests were written before the code, and fail without it.
- [ ] `go test -race ./...` in `gateway/` and `node --test` in `extension/` pass.
- [ ] No key, private data or private term in the diff or the commit messages.
- [ ] The documentation says what changed (README, gateway/README.md, plugin/README.md).
