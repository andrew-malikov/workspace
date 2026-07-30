## Why

`ws` can start tracked Docker Compose projects but cannot explicitly stop a selected project, forcing users to leave the workspace CLI for a routine lifecycle operation. Compose actions also suppress native progress, making multi-project `ws up` runs opaque and difficult to diagnose.

## What Changes

- Add `ws down [alias]` with the same current-directory and tracked-alias target resolution used by `ws up`.
- Add `--blank` and `-b` to `ws down` to stop the target and remove its Compose volumes through one `docker compose down -v` operation.
- Stream native Docker Compose stdout and stderr for mutating operations triggered by both `ws up` and `ws down`.
- Print an escaped project-and-operation banner and exact command preview before each Compose action so output from multi-project `ws up` runs is distinguishable.
- Process non-target projects in stable alias order, while keeping internal running-state probes captured and silent.
- Stop an operation sequence on the first failure and omit the final success result.

## Capabilities

### New Capabilities
- `compose-down`: Explicitly stop a tracked project's Docker Compose stack, optionally removing its volumes.
- `compose-operation-output`: Stream and distinguish native output from Docker Compose lifecycle actions, including multi-project `ws up` sequences.

### Modified Capabilities

None.

## Impact

- Adds a `commands/down` package and registers it in the root CLI.
- Changes workspace Compose orchestration and project target data passed to the Compose adapter.
- Changes `containers.DockerCompose` from captured action output to injected direct subprocess streams while retaining captured query output.
- Changes observable `ws up` output and makes its multi-project operation order deterministic.
- Extends command, workspace orchestration, Compose adapter, and application-level behavior coverage; no configuration or dependency changes are expected.
