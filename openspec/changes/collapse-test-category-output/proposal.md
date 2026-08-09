## Why

`ws test` permanently writes every category's live `dotnet test` output into normal terminal scrollback, so verbose runs can evict earlier category summaries and other useful terminal history. Interactive runs need transient raw output while a category is active, followed by only its concise summary and durable log location after completion.

## What Changes

- Display each active test category's unmodified live output in a transient alternate-screen terminal session.
- Restore the normal terminal screen after every category completes, fails, or is interrupted, then print only that category's existing summary and log path.
- Keep complete per-category output in `.ws/tests/<run-id>/<category>/output.log` regardless of terminal presentation.
- Preserve plain live streaming without terminal-control sequences when output is non-interactive.
- Preserve sequential category order, TRX-derived summaries, and stop-on-failure behavior.

## Capabilities

### New Capabilities

- `transient-test-output`: Interactive lifecycle, restoration, fallback, and failure behavior for ephemeral live test-category output.

### Modified Capabilities

None.

## Impact

- Affects `commands/test`, `dotnet.TestRunner`, and the console capability information used to select interactive presentation.
- Reuses the existing Bubble Tea terminal dependency or equivalent centralized terminal lifecycle handling; no new external dependency is expected.
- Changes interactive terminal presentation only. Durable logs, structured result files, category selection, and non-interactive output remain compatible.
