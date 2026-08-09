## Context

`dotnet.TestRunner.Run` currently opens a per-category log, writes the category heading and command preview to configured stderr, and executes `dotnet test` with stdout and stderr tee writers targeting both their original terminal streams and the shared log. It parses TRX results and writes the summary only after the process and log close. Because every live write lands on the normal screen, a verbose category permanently consumes scrollback before the concise report appears.

The console boundary reports whether stdin and stdout are terminals but does not currently establish that stdout and stderr address the same interactive screen. The implementation must preserve injected streams, plain redirected output, exact child bytes, durable logging, joined error identity, and sequential stop-on-failure behavior. See `specs/transient-test-output/spec.md` for the observable lifecycle.

## Goals / Non-Goals

**Goals:**

- Isolate each interactive category's existing raw output in one alternate-screen lifecycle.
- Restore normal screen contents before category reporting on every controlled return path.
- Keep the subprocess-to-terminal and subprocess-to-log data path unchanged after presentation selection.
- Make interactive selection explicit and independently testable through the console boundary.
- Preserve every process, log, result-parsing, presentation, and report error with deterministic lifecycle ordering.

**Non-Goals:**

- Building a viewport, spinner, progress parser, or managed test-output model.
- Keeping earlier summaries visible while the active category occupies the alternate screen; the normal screen is restored after completion.
- Buffering or truncating process output in memory.
- Interpreting, sanitizing, prefixing, or reformatting child ANSI/control sequences.
- Changing summary wording, log layout, category selection, category order, or failure short-circuiting.
- Recovering the terminal after an uncatchable process termination such as `SIGKILL`.

## Decisions

### Use the alternate screen as a category transaction

Create a small category presentation lifecycle with `Begin` and idempotent `End` operations. Interactive `Begin` enters the terminal's alternate screen before the heading is written; `End` restores the saved normal screen. Plain presentation implements the same lifecycle without emitting control bytes.

`TestRunner.Run` explicitly ends the presentation after the process has exited, the log has closed, and structured results have been loaded, but before it writes the summary or summary-unavailable diagnostic. It also installs a guarded deferred end immediately after successful entry so early returns cannot strand a controlled invocation in the alternate screen. Idempotent end makes the explicit reporting boundary and deferred safety path compatible.

This treats the entire heading, command preview, and raw child stream as transient without counting or erasing lines. Attempting to erase normal-screen output afterward was rejected because wrapped or scrolled lines cannot be identified or removed reliably. Keeping one alternate screen for the whole `ws test` invocation was rejected because completed category reports would not reach normal scrollback until every category finished.

### Preserve raw tee streaming instead of introducing a TUI renderer

After presentation begins, retain the existing tee writers: stdout still goes to configured stdout plus the category log, and stderr still goes to configured stderr plus the same synchronized log. No output is routed through a Bubble Tea model or retained in a viewport buffer.

A managed viewport was rejected because it would merge or parse streams, bound visible history, reinterpret carriage returns and ANSI sequences, and create resize behavior unrelated to the requested collapse. A full Bubble Tea renderer was also rejected for this raw-output option because its redraw ownership would conflict with a subprocess writing directly to the same screen. Use centralized ANSI alternate-screen sequences from the terminal dependency already present in the module rather than scattering literals.

### Select transient presentation only for one shared interactive screen

Extend the console capability boundary with an explicit fact indicating that configured stdout and stderr are terminals attached to the same screen. The OS console derives this from both file descriptors and their terminal device identity; injected consoles supply the fact directly. `commands/test` selects transient presentation only when this fact is true. Every other stream arrangement selects plain presentation.

Checking stdout alone was rejected because stderr carries headings, many `dotnet` diagnostics, and the final category report. Entering an alternate screen on stdout while stderr targets another destination would not provide the specified lifecycle. Assuming two terminal booleans imply the same terminal was rejected because distinct terminal devices are possible.

### Separate restoration from reporting and retain all errors

The category lifecycle order is:

1. create artifacts and open the log;
2. enter presentation;
3. write transient heading and command preview;
4. execute the process through the existing tee writers;
5. close the log and parse structured results;
6. restore normal presentation;
7. write the concise report and log path;
8. return the join of process, close, summary, restoration, and report errors.

An entry failure prevents the process from starting, matching the existing heading-write failure rule. A restoration failure does not suppress the report attempt, and its identity remains in the joined return value. Reports remain on configured stderr.

Reporting from a deferred callback was rejected because it obscures ordering and makes it easy to print the summary before restoration. Returning immediately on the first post-launch error was rejected because it could lose the log path, restoration attempt, or original process failure.

### Convert handled interrupts into context cancellation

Create an interrupt-aware context for the `ws test` action and pass it through category orchestration to `exec.CommandContext`. A handled interrupt cancels the active child, lets `Run` complete its close/parse/restore/report sequence, and then returns failure. Stop signal notification after the action so later process behavior is unchanged.

Relying on an unhandled signal was rejected because immediate process termination does not execute deferred terminal restoration. This guarantee intentionally excludes `SIGKILL` and equivalent uncatchable termination.

## Risks / Trade-offs

- [Earlier summaries are hidden while a category is active] → Restore the untouched normal screen immediately after that category; retaining summaries inside the active view would require the rejected managed viewport.
- [A child emits terminal controls that alter the alternate screen] → Preserve them as raw output; alternate-screen restoration still returns to the saved normal screen, and the durable log retains the exact bytes.
- [Terminal identity detection differs across supported operating systems] → Centralize detection in the OS console adapter and default uncertain or injected configurations to plain presentation.
- [Entering succeeds but the process is forcibly killed] → Handle cancellable interrupts and document that uncatchable termination cannot guarantee cleanup.
- [Presentation writes fail] → Prevent launch on entry failure, attempt restoration after successful entry, attempt reporting afterward, and join every observed error.
- [stdout and stderr writes race in the shared log] → Keep the existing synchronized log writer; this change does not claim stronger cross-stream ordering.

## Migration Plan

1. Add shared-terminal capability detection and injectable presentation selection without changing the default plain runner path.
2. Introduce the idempotent plain and alternate-screen category lifecycles.
3. Reorder category-run reporting around explicit restoration while preserving tee logging and joined errors.
4. Add handled-interrupt cancellation for `ws test`.
5. Verify interactive success, process failure, missing results, interruption, restoration failure, sequential categories, and redirected-output fallback through pseudo-terminal and injected-stream scenarios.

Rollback removes transient selection and the lifecycle wrapper, returning all categories to the existing plain tee path. There is no persisted-data or configuration migration; existing logs and TRX artifacts remain compatible.
