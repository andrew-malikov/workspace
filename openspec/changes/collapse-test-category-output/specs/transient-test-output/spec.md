## Purpose

Keep verbose test-process output visible while useful, then remove it from interactive terminal history while retaining complete durable diagnostics.

## ADDED Requirements

### Requirement: Active category output is transient on a shared interactive terminal
When `ws test` runs with stdout and stderr attached to the same interactive terminal, the system SHALL enter a transient terminal screen before printing the active category heading or launching its test process. It SHALL forward the category's stdout and stderr live without prefixing, reformatting, or waiting for process completion.

#### Scenario: Interactive category starts
- **WHEN** an eligible test category starts with both terminal streams attached to the same interactive terminal
- **THEN** the category heading, command preview, and subsequent process output appear on a transient screen rather than in normal terminal scrollback

#### Scenario: Active process emits output
- **WHEN** the active test process writes to stdout or stderr
- **THEN** the bytes become visible on the transient screen while the process is running

### Requirement: Completed category output collapses to its report
The system SHALL restore the normal terminal screen after every started category finishes and before printing that category's summary diagnostic and log path. The completed category's raw output, heading, and command preview SHALL NOT remain in normal terminal scrollback.

#### Scenario: Category succeeds
- **WHEN** an interactive category exits successfully and its structured results are valid
- **THEN** the system restores the prior normal screen and prints only the category summary and durable log path for that category

#### Scenario: Category fails
- **WHEN** an interactive category exits unsuccessfully
- **THEN** the system restores the prior normal screen before printing the category report and returning the failure

#### Scenario: Summary cannot be produced
- **WHEN** an interactive category finishes but its structured results cannot be summarized
- **THEN** the system restores the prior normal screen before printing the summary-unavailable diagnostic and durable log path

### Requirement: Terminal restoration covers interruption and presentation errors
After entering a transient screen, the system SHALL attempt to restore the normal screen on every controlled completion path, including context cancellation, process failure, result-processing failure, and output-close failure. A terminal restoration failure SHALL be returned without discarding other category failures.

#### Scenario: User interrupts an active category
- **WHEN** the running invocation receives a handled interrupt
- **THEN** the system cancels the active test process, restores the normal screen, and returns a non-success result

#### Scenario: Restoration fails
- **WHEN** restoring the normal screen returns an error
- **THEN** the system retains that error together with any process or reporting error

### Requirement: Non-interactive output remains plain
When stdout and stderr do not share one interactive terminal, the system SHALL NOT emit alternate-screen or other cursor-control sequences and SHALL retain plain live category streaming followed by the category report.

#### Scenario: Output is redirected
- **WHEN** either process output stream is redirected or the streams do not share one interactive terminal
- **THEN** the system streams category output normally and emits no transient-screen control sequences

### Requirement: Complete category evidence remains durable
Transient terminal presentation SHALL NOT reduce or truncate the existing per-category output log. Every stdout and stderr write forwarded to the live presentation SHALL also be written to the category's `output.log`.

#### Scenario: Interactive category completes
- **WHEN** a category displayed on a transient screen produces process output
- **THEN** its durable `output.log` contains the complete output after the transient screen has been removed

#### Scenario: Sequential categories execute
- **WHEN** more than one selected category runs successfully
- **THEN** each category uses a separate transient lifecycle and leaves its concise report on the restored normal screen before the next category starts
