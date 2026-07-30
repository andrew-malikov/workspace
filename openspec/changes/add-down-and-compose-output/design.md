## Context

`ws up` orchestrates multiple tracked projects through `Workspace.UpProject`: it probes every non-target project, stops those with running containers, optionally cleans the target, and starts the target. Project storage is a map, so the current probe, stop, and result order is unstable. `Workspace.DownProject` already resolves and stops a target but has no CLI command or blank-cleanup selection.

`containers.DockerCompose` currently uses `exec.Cmd.Output` for both queries and actions. That is appropriate for the JSON inspected by `HasRunning`, but it captures action stdout and suppresses live Docker progress. The application already owns an injectable `console.Console`; the dotnet runner establishes the local convention of direct child streams plus CLI-owned operational headings on stderr.

The change crosses command registration, workspace orchestration, project-to-container boundaries, subprocess execution, and output behavior. It must preserve native Docker stream behavior while giving serial multi-project output an unambiguous owner.

## Goals / Non-Goals

**Goals:**

- Add `ws down [alias]` with safe default and optional volume cleanup.
- Stream native Docker Compose output for all mutating actions invoked by `ws up` and `ws down`.
- Identify every action by project and operation without transforming child output.
- Keep Compose probes quiet and make multi-project up order deterministic.
- Preserve injected stream behavior, literal dynamic data, failure identity, and final-result channel conventions.

**Non-Goals:**

- Prefixing individual Docker output lines or normalizing Docker ANSI/progress rendering.
- Running Compose actions concurrently.
- Adding arbitrary Compose argument passthrough, service selection, or a general process-event framework.
- Changing `ws up` target-selection, `--alongside`, or cleanup semantics.
- Rolling back already completed stops if a later action fails.

## Decisions

### 1. Pass complete target identity to the Compose boundary

Introduce an application-owned Compose target value containing the tracked alias, working directory, and Compose filename. Change the `containers.Compose` methods and project wrappers to pass that value rather than separate directory and filename strings.

The alias belongs at this boundary because action output must identify the tracked project, while directory and filename define execution. One value keeps these fields coherent and avoids adding presentation callbacks or disconnected label parameters throughout orchestration.

Alternative: derive labels from the directory or Compose filename inside `DockerCompose`. Rejected because either can be shared or unclear, while the tracked alias is the CLI's stable user-facing identity.

### 2. Separate captured queries from streamed actions

Construct `DockerCompose` with the application's configured input, stdout, and stderr. Its query path continues to use captured stdout for `docker compose ps --status running --format json` and returns bytes only to `HasRunning`. Its action path writes operational metadata, attaches the configured streams directly to `exec.Cmd`, and uses `Run`.

`Up`, `Down`, and `Cleanup` use the action path. Direct attachment preserves Docker colors, carriage-return progress, prompts, stdout/stderr separation, and failure behavior without buffering or extra copies.

Alternative: capture output and replay it after completion. Rejected because it is not live and alters ordering and terminal behavior. Alternative: scan and prefix every line. Rejected because line scanning cannot faithfully preserve Docker's terminal control sequences and partial writes.

### 3. Delimit each serial action with stderr metadata

Before launching a mutating subprocess, the action path writes:

1. a single-line heading containing the safely escaped alias and semantic operation;
2. an exact command preview with each dynamic argument safely quoted.

The semantic operations are stop, clean volumes, and start. Headings and previews use configured stderr, matching existing subprocess reporting. Docker stdout and stderr remain untouched on their corresponding destinations. Either metadata write failure prevents process launch and propagates to the application boundary.

Because workspace orchestration is serial, one heading delimits the native output until the next heading; no per-line ownership marker is needed. Blank up therefore produces separate clean-volumes and start blocks for the target, while blank down produces one stop-and-clean-volumes block.

Alternative: send metadata to stdout. Rejected because it would mix CLI operational diagnostics with child stdout and successful result content.

### 4. Sort non-target projects before up orchestration

Build the non-target candidate list in ascending alias order before probing it. Probe and stop each candidate serially, append stopped aliases in that same order, then clean and start the target. Probes remain captured and produce no heading.

Sorting makes action blocks, side effects, and the rendered stopped-project list reproducible. It does not change which projects qualify or the stop-on-first-error behavior.

Alternative: retain map iteration and rely on headings alone. Rejected because labels solve ownership but not unstable execution or result order.

### 5. Make down blank mode select one terminal action

Extend workspace down orchestration with a blank choice. After the existing resolution and Compose-file validation:

- normal mode calls the project's down operation once;
- blank mode calls cleanup once, mapping to `docker compose down -v`.

Blank mode does not call down before cleanup because cleanup already stops the stack. The command renders its result only after the selected action succeeds and includes a volume-removal statement only in blank mode.

Alternative: call down and then cleanup. Rejected as duplicate work that produces two misleading action blocks and increases failure surface.

### 6. Inject the console into Compose-using commands

Update command construction so `ws up`, `ws down`, and Compose status listing receive or construct a stream-configured Docker Compose adapter from the same `console.Console` used by the application. Avoid a usable zero-value action adapter that could silently discard child output.

Register the new down command beside up and use the existing renderer for its final stdout result. Command actions continue to return all domain, process, metadata-write, and render errors to the application boundary.

Alternative: access `os.Stdin`, `os.Stdout`, and `os.Stderr` in the adapter. Rejected because it breaks the established injectable-console contract and isolated CLI tests.

## Risks / Trade-offs

- [Docker writes progress primarily to stderr] → Preserve stdout/stderr exactly and put CLI metadata on stderr so terminal users see coherent serial blocks without stream merging.
- [An action may emit partial output before failing] → Preserve it as diagnostic evidence, stop orchestration immediately, and omit only the final success result.
- [Earlier projects may already be stopped when a later up action fails] → Keep current stop-on-error behavior and perform no implicit rollback; rollback could itself fail and is outside scope.
- [Dynamic aliases or paths could forge headings or terminal controls] → Use single-line escaping for labels and shell-safe quoting for every preview argument; never interpolate raw dynamic data.
- [Changing the Compose interface touches spies and all callers] → Perform a clean cutover and keep target construction centralized in project wrappers.
- [Operational stderr changes existing `ws up` observations] → Treat this as the intended behavior change and cover exact channel routing with injected buffers and fake executors.

## Migration Plan

1. Introduce target identity and the injected query/action Compose adapter, then migrate all callers and spies in one cutover.
2. Make workspace up ordering deterministic and extend down orchestration with blank selection.
3. Add and register the down command and update up construction for live output.
4. Verify default down, blank down, multi-project up labels/order, quiet probes, direct streams, and failure short-circuit behavior.
5. Update user-facing command documentation after behavior passes.

Rollback is a source revert. The change adds no persisted data, configuration fields, or dependency migration; volume deletion occurs only after an explicit blank flag.

## Open Questions

None.
