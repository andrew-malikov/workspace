## ADDED Requirements

### Requirement: Successful results use stdout
The CLI SHALL write successful command result content to stdout and SHALL return exit status `0` when command work and result rendering complete successfully.

#### Scenario: Successful redirected command
- **WHEN** user runs a noninteractive command with stdout redirected
- **THEN** command writes its result to stdout, writes no success content to stderr, and exits `0`

### Requirement: Failures use stderr and nonzero status
The CLI SHALL write command, domain, rendering, and execution failure diagnostics to stderr and SHALL return nonzero exit status. Rendering a diagnostic MUST NOT convert a failure into success.

#### Scenario: Command run outside tracked project
- **WHEN** user runs a command requiring a tracked project from an untracked directory
- **THEN** CLI writes the untracked-directory diagnostic to stderr and exits nonzero

#### Scenario: Result rendering fails
- **WHEN** template execution, terminal rendering, or result writing returns an error
- **THEN** CLI reports failure through stderr when possible and exits nonzero

### Requirement: Nonterminal output is plain
The CLI SHALL emit plain, unpadded result text when output destination is not an interactive terminal. Plain output MUST NOT contain ANSI control sequences introduced by `ws`.

#### Scenario: List output is redirected
- **WHEN** user redirects `ws list` stdout to a file or pipe
- **THEN** result contains readable table or empty-state text without ANSI sequences or terminal-width padding

### Requirement: Color opt-out is honored
The CLI SHALL disable color and ANSI styling introduced by `ws` when `NO_COLOR` is present, regardless of terminal capability.

#### Scenario: NO_COLOR in interactive terminal
- **WHEN** stdout is a terminal and environment contains `NO_COLOR`
- **THEN** successful result contains no `ws`-generated ANSI styling

### Requirement: Interactive terminals retain rich rendering
The CLI SHALL retain rich human rendering for supported commands when output is an interactive color-capable terminal and color is not disabled.

#### Scenario: Rich list output
- **WHEN** user runs `ws list` in a color-capable terminal without `NO_COLOR`
- **THEN** CLI renders the human-oriented styled list to stdout and exits `0`

### Requirement: Dynamic values render literally
The CLI MUST render aliases, paths, branch names, and diagnostic details as literal user data rather than interpreting their Markdown or terminal-control syntax.

#### Scenario: Table value contains Markdown delimiter
- **WHEN** a displayed alias or path contains `|`, `*`, backticks, or a newline
- **THEN** output preserves the value without changing surrounding table or prose structure

### Requirement: Console streams are injectable
Application commands, renderers, subprocess runners, and interactive programs SHALL use caller-supplied input, stdout, and stderr streams instead of directly accessing process-global streams.

#### Scenario: Command runs with in-memory streams
- **WHEN** application runner receives in-memory input, stdout, and stderr streams
- **THEN** all `ws`-owned input and output for that run uses only those streams

### Requirement: Subprocess streams remain direct
The CLI SHALL pass configured input, stdout, and stderr directly to child processes. `ws`-generated headings and command previews SHALL use stderr and MUST NOT be mixed into child stdout.

#### Scenario: Dotnet test execution
- **WHEN** `ws test` launches `dotnet test`
- **THEN** child stdout and stderr stream directly to their configured destinations while `ws` headings appear on stderr

#### Scenario: Child process fails
- **WHEN** child process exits unsuccessfully
- **THEN** CLI preserves streamed child output, reports command failure, and exits nonzero

### Requirement: Alternate-screen UI requires interactivity
The CLI SHALL start alternate-screen interactive UI only when required input and output streams are interactive terminals.

#### Scenario: Clear command in terminal
- **WHEN** user runs `ws clear` with interactive terminal input and output
- **THEN** CLI launches existing Bubble Tea alternate-screen UI

#### Scenario: Clear command without terminal
- **WHEN** user runs `ws clear` with redirected or noninteractive required streams
- **THEN** CLI emits no alternate-screen control sequence, writes clear diagnostic to stderr, and exits nonzero

### Requirement: Output failures propagate
The CLI SHALL check and propagate template execution, renderer, and writer errors. It MUST NOT retry a failed write to the same destination.

#### Scenario: Result writer rejects output
- **WHEN** configured result writer returns an error
- **THEN** command stops rendering and application returns nonzero status without retrying stdout write
