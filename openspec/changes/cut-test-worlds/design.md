## Context

See proposal.md for motivation.

`Workspace.UpProject` / `DownProject` mix resolve, `os.Stat` via `DoesComposeExist`, running probes, and mutating Compose calls. `LoadWorkspace` / `SaveWorkspace` hide `HOME`/`XDG` inside I/O. Every command action calls those plus `os.Getwd` and, for up/down, `containers.NewDockerCompose`. `TestRunner.Run` owns artifact dirs, log files, presentation, `dotnet` argv, TRX parse, and reporting.

`commands/test` already injects a runner and run factory into `runCategories`; that is the target shape.

## Goals / Non-Goals

**Goals:**

- Make compose policy, persistence, command actions, and test-report mappings take values.
- Put process globals and real Docker construction only at the application edge and exec adapters.
- Cut over tests to those seams and delete world-building helpers that exist only to feed the old APIs.

**Non-Goals:**

- Changing CLI flags, templates, Compose argv, exclusive-up semantics, or test output text.
- A shared `testkit` package, fake docker script, or relocated HOME/chdir harness.
- Splitting files without shrinking inputs.
- Changing config.toml schema.
- Broad rejecting-writer or assertion-helper extractions unless they fall out of a deleted fixture.

## Decisions

### 1. Plan compose actions as data, then execute

Introduce a small action value (operation + target identity) and a pure planner that takes:

- target alias and compose target
- other projects as `{alias, configured, running}`
- alongside and blank

Output is only mutating actions: stop, clean volumes, start. Probes are not plan steps. `UpProject` still probes non-target configured projects in alias order, then calls the planner with the observed running flags, then executes.

`DownProject` plans one stop or one cleanup after resolve.

Execution is a serial loop over the plan that stops on first error.

Alternative: keep orchestration in `UpProject` and only table-drive tests with a fixture helper. Rejected — that still builds a world.

Alternative: include `ps` steps in the plan. Rejected — running state is an input to the decision, not part of the decision.

### 2. Split configured from file exists

`Compose != nil` (non-empty filename) means configured. `DoesComposeExist` stays the `os.Stat` check.

Planning uses configured. `UpProject` / `DownProject` keep today's user-facing gate: missing file still errors with the existing “no docker compose configured” message before any plan execution, so CLI behavior is unchanged.

Policy tests never create directories or write yaml.

### 3. Persistence takes a path

Replace in-body path lookup with:

- `Load(path)` / `Save(path, workspace)`
- `ConfigPath()` for existing HOME / `XDG_CONFIG_HOME` rules

Keep thin wrappers `LoadWorkspace` / `SaveWorkspace` only if the application edge wants them; command packages must not call them.

Alternative: inject an `io/fs` filesystem. Rejected — path-in is enough and matches current TOML files.

### 4. Inject a session into commands and `run`

Add a small session value owned next to command construction (package `main` can define it, or a single shared struct passed into each `NewCommand`):

- `Load func() (*workspaces.Workspace, error)`
- `Cwd func() (string, error)`
- `Compose func(console.Console) containers.Compose` (only needed by compose commands)

Default session: load via `ConfigPath` + `Load`, `os.Getwd`, `containers.NewDockerCompose`.

Every current `LoadWorkspace` / `Getwd` / `NewDockerCompose` inside command actions moves to that session. `run` / `newCommand` take the session so application tests inject an in-memory workspace, a cwd string, and a Compose spy.

Do not add a new abstraction package if a struct beside existing command constructors is enough. Prefer one shared struct over copying three function parameters into eight constructors.

Alternative: load workspace once at process start. Rejected — today's per-command load is correct if the file changes between invocations.

Alternative: keep default session inside each `NewCommand` and only inject in unit tests of command packages. Rejected — `main_test` would still need fake docker for up/down.

### 5. Split TestRunner mappings from sequencing

Extract:

- args from kind, filter, results directory
- report text from kind, summary or summary error, log path

`Run` calls those. Inject summary loading and the presentation factory at construction so sequencer tests pass fakes. Stop writing TRX in runner tests. Stop assigning unexported presentation fields.

Leave `NewTestRun` / artifact directories and `loadTestSummary` as honest filesystem units. Leave `StdCommandRunner` helper-process tests as the only child-process coverage for that adapter.

Alternative: inject an `exec.Cmd` factory into `StdCommandRunner` and drop helper processes. Rejected for this change — one honest process suite is acceptable; the waste is repeating it through `TestRunner` and `main`.

### 6. Rewrite tests to the new seams; delete world builders

| Area | After |
|---|---|
| compose policy | tables over planner inputs/outputs |
| plan execution | spy, two cases: happy path and fail-fast |
| persistence | temp file path; one `ConfigPath` test |
| commands / `run` | injected session; no `PATH` docker, no `chdir` |
| runner mappings | string tables |
| runner sequencing | injected exec, summary, presentation |
| compose adapter | keep existing helper-process tests |
| discovery, TRX, git | keep honest fixtures |

Delete `installFakeDocker`, `setupComposeWorkspace` if unused, `useIsolatedWorkspace` chdir, `trackedProjectDir` from policy tests, and `writeTRXFile` in runner tests.

`DoesComposeExist` keeps a small file-based test. `main_test` keeps stdout/stderr/exit routing, including compose failure suppressing the success template, via injected session.

## Risks / Trade-offs

- [Planner drift from `UpProject`] → `UpProject` must only resolve, gate file existence, probe, plan, execute. No extra policy in the wrapper.
- [Session type sprawl] → one struct at the command-construction boundary; do not invent per-command deps types.
- [User-facing error text] → keep the current missing-compose message even when the internal check is file existence after configured.
- [Application tests that still want real `run`] → inject session; do not fall back to a docker script “just in case.”
- [Large test rewrite in one change] → required for a clean cutover; leaving old helpers would preserve the world.

## Migration Plan

1. Add planner, configured check, path-based load/save, and TestRunner mappings without switching callers.
2. Point `UpProject` / `DownProject` / `Run` at those pieces; prove with new tables.
3. Thread session through command constructors and `run`; delete in-action globals.
4. Rewrite tests and delete world-building helpers.
5. Run the full Go test suite. No README change unless a helper or package name is user-visible (it should not be).

Rollback is a source revert. No persisted data or dependency migration.

## Open Questions

None.
