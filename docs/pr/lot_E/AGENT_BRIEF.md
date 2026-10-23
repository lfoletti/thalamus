# Consigne commune des agents de développement du lot E

You develop upstream pull requests for CortexDB (github.com/liliang-cn/cortexdb) on behalf of Lauro Foletti. Read this whole file before you start.

## Hard rules

1. Never push, never run `gh`, never open a PR, never comment anywhere. Never change the branch `proto/granular-cortex` or the main checkout `/Users/mini/Documents/granular_cortex/cortexdb`.
2. Work only in your own worktrees: `git -C /Users/mini/Documents/granular_cortex/cortexdb worktree add -b <branch> /private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/<branch> <base>`. Base `origin/main` (`a591ae1`) unless your task names another branch. Keep the worktrees at the end.
3. Never set `CORTEXDB_TEST_POSTGRES`. Never start PostgreSQL. Lines that only PostgreSQL runs are declared with `DECLARED=<file>` in `make precommit`, and listed in the PR body.
4. One commit per PR, made with `git -c user.name="Lauro Foletti" -c user.email="lauro.foletti@gmail.com" commit`, conventional-commit message, NO `Co-Authored-By` line. No `CHANGELOG.md` change.
5. In `/Users/mini/Documents/granular_cortex` (the Thalamus repo), write only your own files under `docs/pr/lot_E/`. Do not commit there.
6. Do not run more than one `go test -race ./pkg/cortexdb/` at a time from your side. Use `-run` filters while you develop. Run the full package suites of the touched packages once, at the end.

## Read first

- The plan: `docs/pr/lot_E/plan.md`, and the recheck report of your fiche in `docs/pr/lot_E/recheck/`. The report has the evidence, the probes, the named tests and the mutations. Verify its claims; do not trust them blindly.
- `docs/pipeline_pr.md`: sections 1, 3, 3 bis (coding form of a slice: pure core plus one I/O line, class test first, one choke point, about 400 lines) and 4.
- The owner's rules (20 October 2026): about 400 changed lines of non-test code per PR, one rule or one behaviour per PR, tests in the same PR, failing-first tests, a tool-by-tool table for everything a confined key reaches.
- A model body: `docs/pr/round2/GC-11.body.md` (sections `## Description`, `## Type of Change`, `## Testing`, `## Performance Impact`, `## Breaking Changes`, then `## Codecov` when lines stay uncovered). The body names each rule and the test that fails without it ("Failing first" inside Testing, with the command to replay). English, plain short sentences, no judgement of CortexDB, no explanation of the owner's own tools (CI, Codecov).

## Per PR

1. Write the tests first. Run them on the base and see them fail for the expected reason. Keep the FAIL lines.
2. Write the code. Match the surrounding code: naming, comment density, idiom.
3. One hand mutation per rule, each killed by a named test. Use `make handmut` from `/Users/mini/Documents/granular_cortex` (it restores from a copy) or an equivalent copy-and-restore; never `git checkout` to restore. Record each mutation → test → FAIL line in `docs/pr/lot_E/failing_first_<branch>.md`, with the command to replay.
4. `make precommit REPO=<your worktree> BASE=<base>` from `/Users/mini/Documents/granular_cortex` must be green: gofmt, vet, `-race` of the touched packages, and `patchcov` at 100 % of the reachable patch lines. Lines in `func main()` are shown apart; declare them in the body.
5. Write the body in `docs/pr/lot_E/GC-<NN>.body.md` (NN from the branch name). Give the real numbers you measured: lines outside tests (`git diff --stat` on non-test files), tests, mutations.
6. If the rule is unclear, or the owner must decide something, do not invent: write the question in the body under "Questions" and in your report.

## Report

At the end, return a short report: per PR, the branch, the head commit, the lines outside tests, the tests that failed first, the mutations killed, the `precommit` result, and the open questions. State plainly anything you could not finish or verify.
