## Why

Console behavior lacks shared contract: domain failures can exit `0`, redirected output contains ANSI/padding, and application/subprocess output bypasses CLI writers. Standardized semantics make `ws` predictable for terminals, scripts, CI, and tests.

## What Changes

- Define stdout, stderr, and exit-status contracts for success, diagnostics, and failures.
- Route application rendering through injected console I/O instead of process-global streams.
- Render rich output only on capable interactive terminals; emit plain output when redirected or `NO_COLOR` applies.
- Preserve `ws clear` TUI for interactive terminals and reject unsupported noninteractive invocation clearly.
- Preserve child-process stream pass-through while separating `ws` headings/diagnostics from child stdout.
- Propagate template, rendering, and write failures.
- Escape or avoid markup interpretation for dynamic aliases, paths, branch names, and error details.
- Add behavior tests covering streams, status semantics, terminal capabilities, and write failures.
- Remove stale Charm v1 dependencies left by Bubble Tea v2 migration.
- **BREAKING**: rendered domain failures that previously exited `0` will return nonzero status and write diagnostics to stderr.
- **BREAKING**: redirected output will become plain and unpadded instead of ANSI-styled terminal output.

## Capabilities

### New Capabilities

- `console-output`: CLI output channels, failure status, terminal capability handling, interactive fallback, and subprocess stream behavior.

### Modified Capabilities

None.

## Impact

- Entrypoint: `main.go`.
- Rendering and failure presentation: `view/`, `failure/`, command actions.
- Subprocess execution: `dotnet/runner.go`.
- Interactive UI boundary: `commands/git/clear/`.
- Tests: command, renderer, runner, and entrypoint-level behavior coverage.
- Dependencies: `go.mod`, `go.sum`; Bubble Tea/Bubbles v1 cleanup while Glamour's required Lip Gloss v1 remains indirect.
