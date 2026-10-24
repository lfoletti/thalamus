# Contributing to Thalamus

Thank you for your interest in Thalamus. This file gives the rules for a change to this repository. The rules for a change to CortexDB are in the [CortexDB repository](https://github.com/liliang-cn/cortexdb).

## Before you start

- For a change larger than a fix, open an issue first. Describe the need, not only the solution.
- For a security defect, do not open an issue. Read [SECURITY.md](SECURITY.md).
- Thalamus stays thin. A mechanism that serves every CortexDB user goes to CortexDB as a pull request. Thalamus keeps the policy of an organisation: the rights, the groups, the curators and the federation between servers.

## Set up

1. Install Go (the version in `gateway/go.mod`) and Node.js 22.
2. Clone CortexDB at the version in `gateway/go.mod` into `cortexdb/`. The end-to-end benches build its server from there.
3. Install the hook: `git config core.hooksPath tools/hooks`.
4. Write your private terms in `.git/info/private-terms`, one extended regular expression per line. The hook refuses a push without this file. Git does not track it.

## Rules for a change

1. **Write the test before the code.** Each new test fails without the change. In the pull request, name the test that covers each new rule, and show that it fails without the code.
2. **One hand mutation per rule.** Add a line to `gateway/mutate.py` that removes or inverts the rule. Then `python3 -I mutate.py`, run in `gateway/`, must show that a named test fails.
3. **Full coverage.** The patch covers all its reachable lines. The body of a `main` function stays a dispatch to a function that returns an exit code.
4. **No test reaches the network,** an embedder or a brain of the machine. The benches use local servers and the fake identity provider `tests/fake_idp.mjs`.
5. **One change per pull request,** about 400 lines outside tests.
6. **Commit messages** use the conventional-commit style: `feat(gateway): …`, `fix(relay): …`, `docs(readme): …`. The subject says what the code does now.
7. **Plain text.** Short sentences, one idea per sentence. A text about CortexDB says what Thalamus adds. It does not judge the engine.

## Checks

Run these before you push. The hook runs the first three when the push touches `gateway/` or `extension/`.

```sh
cd gateway && test -z "$(gofmt -l .)" && go vet ./... && go test -race ./...
python3 -I mutate.py
cd ../extension && node --test
cd .. && sh tests/e2e_gateway.sh
sh tests/e2e_extension.sh   # needs `npm install` in extension/ once
```

## Pull requests

- Every change to `main` goes through a pull request. Nobody pushes to `main` directly.
- The CI (`.github/workflows/ci.yml`) must pass: gateway checks, extension tests, the gateway bench and a secret scan.
- A maintainer reviews and approves the pull request before the merge.
- In the description, give what changed, why, and the test for each rule. Add the commands that show the failing tests.

## Licence

By contributing, you agree that your contribution is under the MIT licence of this repository.
