## 1. Compose Action Plan

- [x] 1.1 Add a compose action value and a pure planner for exclusive, alongside, blank, and down from in-memory target/other state (configured + running flags). Verify with table tests that need no directories, compose files, or spies.
- [x] 1.2 Add a configured-filename check distinct from `DoesComposeExist`, and keep the existing missing-file error message as the `UpProject`/`DownProject` gate. Verify configured is true without creating files, and that a missing file still fails before any plan runs.
- [x] 1.3 Point `UpProject` and `DownProject` at probe → plan → execute, with execute applying mutating actions in order and stopping on first error. Verify remaining spy tests cover only execution/fail-fast, not policy.
- [x] 1.4 Delete cloned workspace maps, `trackedProjectDir`, and exact `ps`/`down`/`up` literals from policy tests. Verify `workspaces` tests still pass and policy cases are planner tables.

## 2. Path Persistence

- [x] 2.1 Split workspace load/save so they take an explicit config path, and isolate HOME/XDG lookup in `ConfigPath`. Verify load/save tests write and read a temp file without setting `HOME`.
- [x] 2.2 Cover default path resolution in one `ConfigPath` test. Verify existing default location rules still hold.

## 3. Test Runner Parts

- [x] 3.1 Extract `dotnet test` argument construction and category report text as pure mappings. Verify with string-table tests for filtered/unfiltered args and present/missing summaries.
- [x] 3.2 Inject summary loading and presentation construction into `TestRunner` so `Run` only sequences artifacts, streams, execution, restore, and report. Verify sequencer tests use fakes and do not write TRX or assign unexported fields.
- [x] 3.3 Remove `writeTRXFile` and helper-process usage from runner sequencing tests. Leave `StdCommandRunner` helper-process coverage as the only child-process tests for that adapter. Verify `dotnet` tests still pass.

## 4. Command Session Injection

- [x] 4.1 Introduce one session struct (load, cwd, compose factory) and a production default that uses `ConfigPath`+`Load`, `os.Getwd`, and `NewDockerCompose`.
- [x] 4.2 Thread the session through `run`/`newCommand` and every command that currently calls `LoadWorkspace`, `os.Getwd`, or `NewDockerCompose` (`up`, `down`, `list`, `track`, `untrack`, `test`, `scaffold`, `containers`, `git/clear`). Verify command actions contain none of those calls.
- [x] 4.3 Rewrite `main_test.go` compose coverage to inject a workspace, cwd, and Compose spy. Delete `installFakeDocker`, `chdir` isolation, and unused compose-file setup. Verify up/down still assert exit codes, stream routing, and success-template suppression without a `PATH` docker.

## 5. Cutover Verification

- [x] 5.1 Remove leftover world-building helpers (`useIsolatedWorkspace` if unused, `setupComposeWorkspace` if unused) and any test that rebuilds HOME+PATH+chdir to assert policy or argv. Verify no test binary still looks for `WS_DOCKER_*`.
- [x] 5.2 Run the full Go test suite and confirm user-facing up/down/test output and exclusive-up semantics are unchanged.
