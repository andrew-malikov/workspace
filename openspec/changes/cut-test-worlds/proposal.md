## Why

Tests reconstruct a whole machine — home directory, cwd, compose files, PATH docker, TRX XML, alt-screen — to assert one decision. That cost comes from production functions that close over process globals and filesystem instead of taking values. Helpers would only hide the world; the APIs have to shrink.

## What Changes

- Split compose up/down into a pure action plan plus a thin executor. Policy tests assert `[]Action` from project state and flags, with no directories, compose files, or spies.
- Treat “compose configured” as the Compose filename pointer. Keep “compose file exists” as a separate filesystem check used only when that is the question.
- Load and save workspace from an explicit config path. `HOME` / `XDG` resolution stays in one path helper, not in every persistence call.
- Construct CLI commands with injected workspace, Compose adapter, and cwd. Command actions MUST NOT call `LoadWorkspace`, `os.Getwd`, or `NewDockerCompose` themselves.
- Split `TestRunner` so command-arg construction and summary reporting are pure. `Run` only sequences injected artifacts, presentation, command execution, summary loading, and reporting.
- Rewrite the heavy tests to match those seams. Delete `installFakeDocker`, `useIsolatedWorkspace` chdir, cloned workspace maps in policy tests, and runner tests that invent TRX or poke unexported presentation factories.
- Keep honest-world tests only where the job *is* the filesystem or a child process: TOML load/save, compose-file existence, TRX parse, test discovery, git history, and one helper-process suite per exec wrapper.

No user-facing CLI flags, output, or Compose semantics change.

## Capabilities

### New Capabilities
- `compose-action-plan`: Plan exclusive-up, alongside, blank, and down as a list of compose actions from in-memory project state.
- `command-injection`: Commands and workspace persistence take explicit workspace, Compose, cwd, and config-path values instead of process globals.
- `test-runner-parts`: Dotnet test command args and category reports are separable from artifact, process, and presentation sequencing.

### Modified Capabilities

None. Existing compose-down, compose-operation-output, and test-output-reporting behavior stays the same.

## Impact

- Touches `workspaces` orchestration and persistence, `commands/up`, `commands/down`, other command constructors that load workspace or cwd, `dotnet.TestRunner`, and the corresponding tests.
- `main` / command construction becomes the I/O edge: resolve config path, load workspace, read cwd, build `DockerCompose`, pass them in.
- Application-level `main_test.go` stops faking docker and changing directories. It keeps exit-code and stdout/stderr routing coverage that cannot live lower.
- No persisted config format, dependency, or README command-behavior changes.
