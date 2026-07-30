## Context

`ws test` resolves configured test categories and runs them sequentially through `dotnet.TestRunner`. The runner currently attaches `dotnet test` directly to the terminal, so output is live but ephemeral; it neither captures a per-category log nor exposes structured counts. The change crosses command orchestration, process I/O, result parsing, and filesystem management while preserving the existing category selection, ordering, and failure behavior.

## Goals / Non-Goals

**Goals:**
- Keep stdout and stderr visible while each `dotnet test` process runs.
- Persist the same category output in a stable, human-navigable run directory.
- Produce reliable passed, failed, skipped, and total counts after each completed category.
- Print an absolute log path after the summary so link-aware terminals can open it.
- Retain the underlying command error and existing stop-on-failure semantics after reporting the completed category.

**Non-Goals:**
- Running categories concurrently.
- Replacing or reformatting `dotnet`'s live output.
- Adding log retention, cleanup, upload, or cross-run aggregation.
- Rendering terminal-specific hyperlink escape sequences.
- Changing test discovery, category filters, or category selection flags.

## Decisions

### Use one project-local run directory with category subdirectories

Create one run root for each `ws test` invocation at `.logs/tests/<run-id>/`, beneath the resolved project directory. `<run-id>` uses a sortable UTC timestamp plus a collision-resistant suffix. Each selected category receives `<category>/output.log` and a `<category>/results/` directory for structured results.

This groups a multi-category invocation, keeps paths traversable by time and category, and prevents later runs from overwriting earlier evidence. A single flat log was rejected because categories and concurrent invocations would be difficult to distinguish.

### Tee process output instead of buffering it

Open the category log before starting `dotnet test`, then route stdout and stderr through synchronized multi-writers that write to both their original terminal streams and the same log. Close the log only after the process exits.

This preserves live feedback and bounds memory regardless of output volume. Capturing output in memory and replaying it was rejected because it removes live progress and can allocate without bound. The combined log preserves writes as received, although ordering between independently emitted stdout and stderr cannot be made stronger than the process-stream interface provides.

### Request TRX results and aggregate them with the standard library

Invoke `dotnet test` with a TRX logger and a category-specific results directory. After process completion, discover all TRX files in that directory and aggregate their result counters with Go's XML decoder. This supports solutions with multiple projects or target frameworks without parsing localized console prose and adds no external dependency.

Parsing the console summary was rejected because formatting, verbosity, SDK versions, and localization can change. Treat absent or malformed TRX as a reporting error: preserve the full log path, report that counts are unavailable, and return an error rather than inventing a summary.

### Report before propagating the test command result

After `dotnet test` exits, parse available results, print one concise category summary (`total`, `passed`, `failed`, `skipped`), then print the absolute `output.log` path. If tests or infrastructure failed, return the original command failure after reporting; if result parsing also failed, retain both causes in the returned diagnostic.

This preserves current stop-on-failure behavior while ensuring the failed category remains diagnosable. Continuing with later categories after a failure was rejected because it changes established command semantics and was not requested.

### Keep terminal linking portable

Print a labeled absolute native filesystem path as plain text. Modern terminals detect such paths without terminal-specific control sequences; plain output also remains readable in redirected output and CI logs. OSC 8 links were rejected because support varies and escape sequences would pollute non-interactive output.

## Risks / Trade-offs

- [TRX shape differs across SDK or runner versions] → Decode only the stable result-summary counters, cover representative fixtures, and fail explicitly when counters are unavailable.
- [Multiple target frameworks produce several result files] → Recursively discover and aggregate every TRX file under the category results directory.
- [stdout and stderr writes race into one file] → Serialize writes to the shared log sink while preserving separate terminal destinations.
- [`.logs` grows indefinitely] → Use isolated sortable run directories and document that retention is intentionally user-managed; automated deletion is outside this change.
- [Project-local logs appear as untracked files] → Add `.logs/` to this repository's ignore rules and make the directory contract explicit for projects using `ws`.
- [A log cannot be created] → Fail before launching tests so the command never claims durable logging that did not occur.
