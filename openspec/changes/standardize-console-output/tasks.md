## 1. Console Boundary

- [x] 1.1 Add application console value with injected input, stdout, stderr, terminal capability, and `NO_COLOR` handling.
- [x] 1.2 Extract testable application runner that configures urfave streams, suppresses library-owned process exit, and returns final exit code.
- [x] 1.3 Make `main` construct OS-backed console and perform sole process exit.

## 2. Result Rendering

- [x] 2.1 Replace global-output `view.Render` with injected rich/plain renderer and propagate template, Glamour, and writer errors.
- [x] 2.2 Add renderer behavior coverage for rich terminal output, redirected plain output, `NO_COLOR`, and rejecting writers.
- [x] 2.3 Render command table/prose dynamic values literally, including Markdown delimiters and newlines.
- [x] 2.4 Migrate list, containers, track, untrack, up, scaffold, and shared view callsites to injected renderer without compatibility wrappers.

## 3. Failure Semantics

- [x] 3.1 Represent untracked-directory and untrack domain failures as returned errors instead of rendered success.
- [x] 3.2 Centralize known and unknown failure presentation on stderr with nonzero status.
- [x] 3.3 Add application-level coverage proving domain, render, and command failures use stderr/nonzero while successful results use stdout/zero.

## 4. Terminal Integrations

- [x] 4.1 Inject dotnet runner streams, move `ws` headings/previews to stderr, preserve direct child stream pass-through, and cover success/failure behavior.
- [x] 4.2 Inject Bubble Tea input/output and gate `ws clear` alternate-screen startup on interactive terminal capabilities.
- [x] 4.3 Add `ws clear` coverage for interactive launch and noninteractive stderr/nonzero rejection without terminal control sequences.

## 5. Verification and Dependency Cleanup

- [x] 5.1 Smoke-test `ws list` in terminal, redirected, and `NO_COLOR` modes; verify ANSI/padding and stream/status contracts.
- [x] 5.2 Smoke-test untracked `ws test` and noninteractive `ws clear`; verify diagnostics and nonzero exits.
- [x] 5.3 Run full Go test suite after behavior smoke checks pass.
- [x] 5.4 Run module tidy, remove stale Bubble Tea/Bubbles v1 dependencies, and verify Charm v2 direct dependency classification.
