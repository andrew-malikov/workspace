## ADDED Requirements

### Requirement: Live category output
The system SHALL identify each selected test category and stream that category's `dotnet test` stdout and stderr to their corresponding terminal streams while the process is running.

#### Scenario: Category starts
- **WHEN** `ws test` begins an eligible unit, integration, or component category
- **THEN** the system prints the category identity before that category's `dotnet` output

#### Scenario: Dotnet emits output
- **WHEN** the active `dotnet test` process writes to stdout or stderr
- **THEN** the system makes that output visible on the corresponding terminal stream without waiting for the category to finish

### Requirement: Durable per-category logs
The system SHALL persist the complete process output for each executed category beneath `.logs/tests/<run-id>/<category>/output.log` in the resolved project directory, where each invocation has a unique, sortable run identifier and category logs from the same invocation share one run directory.

#### Scenario: Multiple categories execute
- **WHEN** one `ws test` invocation executes more than one category
- **THEN** each category has its own `output.log` under the same unique run directory

#### Scenario: A later invocation executes
- **WHEN** `ws test` is invoked again
- **THEN** the system creates a different run directory and does not overwrite logs from the earlier invocation

#### Scenario: Dotnet writes to either stream
- **WHEN** the category's `dotnet test` process writes to stdout or stderr
- **THEN** the system appends that output to the category's `output.log` while also streaming it live

#### Scenario: Log setup fails
- **WHEN** the system cannot create or open the category log
- **THEN** it returns an error before starting that category's `dotnet test` process

### Requirement: Structured category summary
The system SHALL derive `total`, `passed`, `failed`, and `skipped` counts from all structured test result files produced for a category and print exactly one concise summary after that category's process exits. It SHALL NOT derive counts by parsing localized console output.

#### Scenario: Successful category completes
- **WHEN** a category finishes and its structured result files are valid
- **THEN** the system prints the aggregate total, passed, failed, and skipped counts for that category

#### Scenario: Multiple result files are produced
- **WHEN** a category produces structured result files for multiple projects or target frameworks
- **THEN** the summary aggregates counters from every result file under that category's results directory

#### Scenario: Results cannot be summarized
- **WHEN** no structured result file exists or a result file lacks valid counters
- **THEN** the system reports that the category summary is unavailable and returns a reporting error without fabricating counts

### Requirement: Discoverable log location
The system SHALL print the absolute native filesystem path of the category's `output.log` immediately after the category summary or summary-unavailable diagnostic.

#### Scenario: Category reporting completes
- **WHEN** the system has finished summarizing an executed category
- **THEN** it prints a labeled absolute path to that category's log as plain text

### Requirement: Failure semantics remain intact
The system SHALL report the completed category's summary status and log path before returning its command or reporting error, and SHALL preserve sequential category execution with stop-on-failure behavior.

#### Scenario: Tests fail
- **WHEN** `dotnet test` exits unsuccessfully after producing valid structured results
- **THEN** the system prints the failing category's counts and log path and then returns a non-success result without starting a later category

#### Scenario: Dotnet infrastructure fails
- **WHEN** `dotnet test` exits unsuccessfully without valid structured results
- **THEN** the system prints the summary-unavailable diagnostic and log path and then returns a non-success result without starting a later category

#### Scenario: Category succeeds
- **WHEN** a category completes successfully with valid structured results and another category was selected
- **THEN** the system reports the completed category and then starts the next category in configured order
