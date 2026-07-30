## 1. Compose Execution Boundary

- [x] 1.1 Add a Compose target value carrying alias, working directory, and Compose filename; migrate the Compose interface, project wrappers, callers, and workspace spies to use it.
- [x] 1.2 Replace the zero-value Docker Compose adapter with console-stream construction and split captured `ps` queries from directly streamed mutating actions.
- [x] 1.3 Add safely escaped single-line action headings and exact command previews on stderr, preventing process launch when either metadata write fails.
- [x] 1.4 Cover quiet query capture, direct stdin/stdout/stderr routing, action argument selection, safe metadata formatting, process failures, and metadata-write failures in Compose adapter tests.

## 2. Workspace Lifecycle Orchestration

- [x] 2.1 Sort non-target projects by alias before `ws up` probes and stops them, preserving serial execution and the same order in the stopped-project result.
- [x] 2.2 Extend down orchestration with a blank choice that selects exactly one normal down or cleanup action after existing target and Compose-file validation.
- [x] 2.3 Extend workspace spy tests for deterministic multi-project order, quiet probes, default down, blank down, missing targets/configuration, and stop-on-first-failure sequencing.

## 3. CLI Commands and Wiring

- [x] 3.1 Add `commands/down` with optional alias, `--blank`/`-b`, current-directory resolution, success rendering, and error propagation.
- [x] 3.2 Inject the application console into Compose-using command construction, enable live actions in `ws up`, register `ws down`, and keep Compose status queries silent.
- [x] 3.3 Add command and application-level coverage for down registration and flags, stdout success summaries, stderr operation metadata, blank-result wording, nonzero failures, and suppression of success output after failure.

## 4. Verification and Documentation

- [x] 4.1 Run focused container, workspace, command, and application tests covering both capability specs, then run the full Go test suite.
- [x] 4.2 Build the CLI and smoke-test help plus controlled multi-project up, default down, blank down, direct stream routing, stable labels/order, and failed-action short-circuit behavior.
- [x] 4.3 Document `ws down`, its alias and blank forms, and live labeled Compose output for up and down in the existing README command section.
