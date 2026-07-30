## Why

`ws test` currently streams raw `dotnet test` output without leaving a durable, category-oriented result, making multi-category runs difficult to review after completion. Each category should remain observable while it runs, then conclude with a concise result summary and a discoverable path to its full log.

## What Changes

- Stream `dotnet test` output live for each selected test category: unit, integration, and component.
- Print a per-category completion summary with passed, failed, and other relevant test counts.
- Persist each category's full output under a structured, traversable `.logs` hierarchy scoped to the test run.
- Print the resulting log path in terminal output so terminals can render it as a navigable file link.
- Preserve category ordering and stop/failure semantics while ensuring the completed category's summary and log location remain available.

## Capabilities

### New Capabilities
- `test-output-reporting`: Live category output, per-category result summaries, and durable discoverable test logs for `ws test`.

### Modified Capabilities

None.

## Impact

- Affects `commands/test` orchestration and the `dotnet` test runner abstraction.
- Adds filesystem output under a project-local `.logs` directory.
- May require structured `dotnet test` result output or parsing to obtain reliable counts while retaining live console streaming.
- Requires updates to command and runner tests for output, summaries, log paths, and failure behavior.
