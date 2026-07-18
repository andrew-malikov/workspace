## Context

`ws` output currently has four owners: urfave/cli, `view.Render`, Bubble Tea, and `dotnet.StdCommandRunner`. Application paths use process-global streams, `view.Render` forces Glamour's `dark` style, domain failures sometimes render then return `nil`, and urfave's default exit handler can terminate before a testable application boundary regains control.

Change crosses entrypoint, commands, rendering, subprocesses, and interactive UI. Existing rich terminal experience remains valuable; redirected and automated use requires deterministic stream/status behavior.

## Goals / Non-Goals

**Goals:**

- One console contract across command results, diagnostics, failures, child processes, and TUI launch.
- Injectable input/output/error streams and terminal capability decisions.
- Rich human output on capable terminals; plain deterministic output elsewhere.
- Domain failures retain failure identity through process exit.
- All render/write errors propagate.
- Dynamic values render literally.
- Existing command vocabulary and successful business behavior remain unchanged.

**Non-Goals:**

- JSON or other machine-readable result schema.
- Redesign of `ws clear` interaction or visual styling.
- Capturing, buffering, or transforming child-process output.
- General logging/telemetry framework.
- Refactor of unrelated workspace/domain code.

## Decisions

### 1. Add one injected console boundary

Create a small application-owned console value containing `io.Reader` input, stdout/stderr `io.Writer`s, and terminal/color capability state. Construct it from `os.Stdin`, `os.Stdout`, and `os.Stderr` in `main`; pass dependencies into command constructors, renderers, subprocess runners, and TUI launch.

Keep presentation in `view`; console value supplies channels/capabilities rather than owning command-specific text. Configure urfave `Reader`, `Writer`, and `ErrWriter` from same value.

Alternative: continue globals plus capture descriptors in tests. Rejected: global mutation prevents isolated parallel tests and preserves split policy.

### 2. Make application runner own exit translation

Extract testable runner returning exit code. Disable urfave's default process-exiting error handler; command actions return errors without rendering them as success. Application boundary maps known domain failures to user diagnostics, writes them to stderr, and returns nonzero. `main` performs sole `os.Exit` call.

Keep success rendering inside commands because result data and templates remain command-owned. Failure rendering moves to boundary or returns an error after any required presentation; no failure path returns `nil`.

Alternative: render failures inside every command and wrap sentinel errors. Rejected: duplicates mapping/channel/status logic.

### 3. Select renderer from destination capability

For human success output:

- stdout terminal + color enabled → existing Glamour-rich rendering.
- redirected/non-terminal stdout or `NO_COLOR` present → plain rendering without ANSI or terminal-width padding.

Terminal detection occurs once at application construction and is injectable for tests. `NO_COLOR` overrides detected color support. Plain output remains human-readable; stable machine schema stays out of scope.

Use explicit renderer configuration rather than `glamour.Render(..., "dark")` convenience defaults. Preserve dark style initially for rich mode; style selection can evolve independently.

Alternative: always strip ANSI after rich rendering. Rejected: retains Glamour padding/layout artifacts and needless work.

### 4. Separate templates from dynamic literal values

Structured tables must render cells as literal data, not raw Markdown syntax. Prefer direct table/text construction where markup provides no value; where Markdown remains useful for prose, escape dynamic values before template execution.

All template execution, render, and writer errors return to caller. Writer APIs use `io.Writer`; no direct `fmt.Print*` remains in application presentation paths.

Alternative: trust aliases/paths as benign. Rejected: valid path and alias characters can alter Markdown structure.

### 5. Preserve child pass-through with explicit streams

`dotnet.StdCommandRunner` receives stdin/stdout/stderr from console boundary and attaches them directly to `exec.Cmd`. `ws`-generated test headings and command previews go to stderr, leaving child stdout unmodified. Runner returns both heading-write and process errors.

No capture or replay: direct pass-through preserves streaming, colors, prompts, and child exit behavior without extra allocations.

### 6. Gate alternate-screen UI on interactivity

Launch `ws clear` Bubble Tea UI only when required input and output streams are terminals. Unsupported noninteractive invocation returns a clear stderr diagnostic and nonzero status before alternate-screen control sequences are emitted. Bubble Tea receives injected streams/options rather than process globals.

Alternative: silently degrade to printed branch list. Rejected: deletion semantics need explicit noninteractive UX and flags not included in this change.

### 7. Clean dependency state after behavior works

Run module tidy after console behavior passes smoke checks. Bubble Tea/Bubbles v1 disappear; Charm v2 modules become direct; Glamour-required Lip Gloss v1 remains indirect. No library replacement in this change.

## Risks / Trade-offs

- Existing scripts may rely on exit `0` or ANSI output for domain failures → document breaking behavior in proposal; lock new contract with process-level tests.
- Terminal detection differs across platforms → isolate detector behind injectable capability state; cover terminal/nonterminal branches without requiring real TTY in unit tests.
- Broken-pipe handling can emit recursive diagnostics → return write failure once; process boundary avoids retrying same failed destination.
- Markdown escaping can miss edge cases → prefer direct text/table rendering for structured data.
- Constructor signatures will change across commands → clean cutover; migrate every caller without compatibility wrappers.
- Rich and plain renderers can drift → shared semantic result/template inputs plus paired behavior tests.

## Migration Plan

1. Introduce console capability/value and testable application runner.
2. Migrate static command rendering and failure mapping.
3. Migrate dotnet and Bubble Tea streams.
4. Add process/renderer/runner behavior coverage and smoke-test TTY plus redirected paths.
5. Tidy modules after behavior passes.

Rollback: revert change as one unit. No persisted data or configuration migration.

## Open Questions

None. JSON output and noninteractive branch deletion remain separate future changes.
