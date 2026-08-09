## 1. Shared Terminal Capability

- [x] 1.1 Extend `console.Console` construction with an explicit shared-interactive-output capability and migrate every constructor callsite without changing existing color or input-terminal behavior.
- [x] 1.2 Implement OS detection that requires terminal stdout and stderr on the same device, with uncertain and injected configurations defaulting to plain output.
- [x] 1.3 Cover shared-terminal, redirected-stream, distinct-device, and existing console-construction behavior with focused tests.

## 2. Category Presentation Lifecycle

- [x] 2.1 Add a category presentation abstraction with plain and alternate-screen implementations, using centralized terminal control constants rather than inline escape literals.
- [x] 2.2 Make alternate-screen entry failure prevent execution and make restoration idempotent so explicit and deferred cleanup can safely coexist.
- [x] 2.3 Verify exact entry and exit bytes, plain no-op behavior, repeated restoration, and writer-error propagation with injected writers.

## 3. Test Runner Integration

- [x] 3.1 Select presentation from the console capability when constructing `dotnet.TestRunner`, retaining plain behavior for every non-shared-terminal configuration.
- [x] 3.2 Reorder `TestRunner.Run` so presentation starts before the heading, restoration completes before reporting, and the existing stdout/stderr tee path and durable log remain unchanged.
- [x] 3.3 Join process, log-close, summary, restoration, and report errors without suppressing the category report or changing stop-on-failure semantics.
- [x] 3.4 Cover successful, failed, summary-unavailable, entry-failure, restoration-failure, and multi-category lifecycle ordering, including complete durable logs and absence of control sequences in plain mode.

## 4. Controlled Interruption

- [x] 4.1 Run `ws test` categories with an interrupt-aware context that cancels the active child and releases signal handling after the command returns.
- [x] 4.2 Verify a handled interrupt terminates the child, restores the normal screen before diagnostics, retains available log evidence, and returns failure.

## 5. End-to-End Verification

- [x] 5.1 Run the focused console, dotnet runner, and test-command suites covering the changed contract.
- [x] 5.2 Exercise `ws test` through a pseudo-terminal and confirm raw category output appears only on the alternate screen while the restored normal output retains per-category summaries and log paths.
- [x] 5.3 Exercise redirected stdout and stderr combinations and confirm plain live output contains no alternate-screen or cursor-control sequences.
