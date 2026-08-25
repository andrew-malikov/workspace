## Purpose

Separate dotnet test argument construction and category reporting from artifact creation, child execution, and terminal presentation so each contract can be checked with values.

## ADDED Requirements

### Requirement: Test command arguments are a pure mapping
The arguments passed to `dotnet test` for a category SHALL be determined only by the category kind, optional filter, and results directory.

#### Scenario: Unfiltered category
- **WHEN** the category is unit and no filter is set
- **THEN** the arguments are `test`, a TRX logger prefixed with `unit`, and the given results directory

#### Scenario: Filtered category
- **WHEN** a filter is set
- **THEN** the arguments include `--filter` and that filter after the logger and results directory

### Requirement: Category report is a pure mapping
The category completion text SHALL be determined only by the kind, the loaded summary or summary error, and the log path.

#### Scenario: Summary present
- **WHEN** a summary of total 4, passed 2, failed 1, skipped 1 is reported for unit at a log path
- **THEN** the report contains `unit summary: total 4, passed 2, failed 1, skipped 1` and `Log: ` followed by that path

#### Scenario: Summary missing
- **WHEN** summary loading fails for component
- **THEN** the report contains `component summary unavailable:` and `Log: ` followed by the log path

### Requirement: Sequencing stays independent of those mappings
Running a category SHALL still stream output, persist the log, restore presentation before reporting, and retain command, cancellation, and presentation errors. Those behaviors MUST remain testable with injected command execution, summary loading, and presentation, without writing TRX files or constructing a project tree.

#### Scenario: Report follows restoration
- **WHEN** a category run finishes, including after cancellation or a presentation restore error
- **THEN** the report is emitted after presentation ends, and the command or restore error is still returned

#### Scenario: Command is not started after setup failure
- **WHEN** artifact, log, heading, or presentation setup fails
- **THEN** no test command is executed

### Requirement: Child-process execution stays at one boundary
Actual child-process stream attachment and cancellation SHALL remain covered only at the command-execution adapter, not by reconstructing that adapter inside category-sequencing tests.

#### Scenario: Adapter owns the process
- **WHEN** the standard command runner executes a child
- **THEN** it still attaches the configured streams and terminates the child on cancellation
