## ADDED Requirements

### Requirement: Compose actions stream native process output
The CLI SHALL attach its configured input, stdout, and stderr directly to every mutating Docker Compose subprocess started by `ws up` or `ws down`. It MUST NOT capture, buffer, replay, merge, or prefix the subprocess output.

#### Scenario: Up streams native output
- **WHEN** `ws up` starts, stops, or cleans a Compose project
- **THEN** each action's Docker stdout and stderr stream live to the corresponding configured CLI destinations

#### Scenario: Down streams native output
- **WHEN** `ws down` stops or cleans a Compose project
- **THEN** the action's Docker stdout and stderr stream live to the corresponding configured CLI destinations

### Requirement: Compose actions are identified before execution
Before each mutating Compose subprocess, the CLI SHALL write a single-line heading containing the tracked project alias and semantic operation, followed by an exact safely escaped command preview. CLI-owned headings and previews SHALL use stderr.

#### Scenario: Multiple projects are stopped by up
- **WHEN** `ws up` stops two or more other tracked Compose projects
- **THEN** each project's native output is preceded by a heading that identifies that project and its stop operation

#### Scenario: Blank up has two target actions
- **WHEN** `ws up --blank` cleans and then starts the target
- **THEN** cleanup and startup receive separate headings and command previews for the same target alias

#### Scenario: Blank down identifies cleanup
- **WHEN** `ws down --blank` executes `docker compose down -v`
- **THEN** its heading identifies the target and volume-cleanup operation before native output begins

#### Scenario: Dynamic target data is displayed safely
- **WHEN** a project alias, directory, or Compose filename contains whitespace, a newline, quotes, or control characters
- **THEN** the heading and command preview preserve it as escaped single-line data without creating a forged output block or terminal control sequence

### Requirement: Multi-project up order is deterministic
When `ws up` considers other tracked projects, the CLI SHALL process candidates in ascending alias order, SHALL execute each required stop serially in that order, and SHALL execute target cleanup and startup after all required stops. Running-state probes MUST remain captured and silent.

#### Scenario: Several other projects are running
- **WHEN** projects `worker`, `api`, and `database` are running while another target is brought up
- **THEN** their stop actions and reported stopped-project list use the order `api`, `database`, `worker`, followed by target actions

#### Scenario: Running-state probes produce data
- **WHEN** `ws up` invokes `docker compose ps` to determine whether another project is running
- **THEN** probe output is consumed internally and does not appear in configured CLI output

### Requirement: Compose action failures stop orchestration
The CLI SHALL stop the current lifecycle sequence on the first heading-write, preview-write, or Compose execution failure. It MUST NOT launch later Compose actions or render the command's final success result after such a failure.

#### Scenario: Stopping another project fails
- **WHEN** a stop action for another project fails during `ws up`
- **THEN** the CLI starts no later stop, cleanup, or startup action and exits nonzero

#### Scenario: Target cleanup fails
- **WHEN** target cleanup fails during blank `ws up`
- **THEN** the CLI does not start the target and exits nonzero

#### Scenario: Operational heading cannot be written
- **WHEN** stderr rejects a Compose action heading or command preview
- **THEN** the CLI does not launch that subprocess and exits nonzero
