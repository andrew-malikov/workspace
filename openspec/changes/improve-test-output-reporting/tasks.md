## 1. Result and Path Foundations

- [x] 1.1 Add a TRX result model and XML aggregation logic for total, passed, failed, and skipped counters across every result file in a category directory.
- [x] 1.2 Cover single-file, multi-file, missing-file, malformed-file, and missing-counter TRX behavior with focused `dotnet` package tests.
- [x] 1.3 Add project-local `.logs/tests/<run-id>/<category>` path creation with sortable unique run identifiers, shared run roots, and absolute output paths.

## 2. Live Output and Artifact Capture

- [x] 2.1 Evolve the command execution boundary to support category-scoped stdout and stderr sinks while preserving stdin, working directory, argument handling, and wrapped process errors.
- [x] 2.2 Implement synchronized streaming tees that send stdout and stderr to their original terminal streams and the category `output.log` without buffering the complete run in memory.
- [x] 2.3 Configure `dotnet test` to emit TRX files into the category results directory while preserving existing category filters and readable command diagnostics.
- [x] 2.4 Guarantee log setup happens before process launch and log closure happens after all process writes, including unsuccessful process exits.

## 3. Category Reporting

- [x] 3.1 Create one run root per `ws test` invocation and pass the corresponding category paths through sequential unit, integration, and component execution.
- [x] 3.2 Print each category heading before live output, then print exactly one concise aggregate summary and the absolute plain-text `output.log` path after process completion.
- [x] 3.3 Report unavailable summaries without fabricated counts, preserve the original process failure alongside reporting failures, and retain stop-on-failure category semantics.
- [x] 3.4 Add `.logs/` to repository ignore rules so generated local test evidence does not enter version control.

## 4. Verification

- [x] 4.1 Extend runner tests to prove output is visible before process completion, stdout and stderr reach the durable log, TRX counts aggregate correctly, and failure reports still expose the log path.
- [x] 4.2 Extend `ws test` command tests to prove categories share one run root, retain configured order, advance after success, and stop only after reporting a failed category.
- [x] 4.3 Run the focused Go test packages and smoke `ws test` with a controlled `dotnet` executable to verify the terminal sequence, summary counts, exit status, and traversable on-disk log hierarchy end to end.
