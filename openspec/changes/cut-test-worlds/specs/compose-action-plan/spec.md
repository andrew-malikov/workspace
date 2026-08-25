## Purpose

Describe exclusive-up, alongside, blank, and down as a list of compose actions derived from in-memory project state so policy can be checked without a filesystem or process.

## ADDED Requirements

### Requirement: Plan uses in-memory project state
The system SHALL produce a compose action plan from a target identity, a list of other projects with configured and running flags, and the alongside and blank choices. The plan MUST NOT consult the filesystem or invoke Compose.

#### Scenario: Exclusive up stops other running projects
- **WHEN** alongside is false, blank is false, and other projects include running `api` and `billing` plus stopped `worker`
- **THEN** the plan is stop `api`, stop `billing`, start the target, in ascending alias order among the stopped projects

#### Scenario: Alongside up ignores other projects
- **WHEN** alongside is true and other projects are running
- **THEN** the plan is only start the target

#### Scenario: Blank exclusive up cleans the target after stops
- **WHEN** alongside is false, blank is true, and another project is running
- **THEN** the plan is stop each running other project in alias order, clean the target volumes, then start the target

#### Scenario: Blank alongside up skips other projects
- **WHEN** alongside is true and blank is true
- **THEN** the plan is clean the target volumes, then start the target

#### Scenario: Unconfigured others are omitted
- **WHEN** an other project has no compose configuration
- **THEN** the plan includes no probe or stop for that project

### Requirement: Down plan selects one action
The system SHALL plan default down as one stop of the target and blank down as one volume-cleanup of the target. The plan MUST NOT include a stop before cleanup.

#### Scenario: Default down
- **WHEN** blank is false
- **THEN** the plan is one stop of the target

#### Scenario: Blank down
- **WHEN** blank is true
- **THEN** the plan is one volume-cleanup of the target

### Requirement: Execution follows the plan and stops on first failure
The system SHALL apply planned actions in order through the Compose adapter and SHALL stop at the first action error, leaving later actions uninvoked.

#### Scenario: Executed exclusive up
- **WHEN** a plan of stop `api`, stop `billing`, start `orders` is executed
- **THEN** Compose receives those three mutating actions in that order

#### Scenario: Failure skips remaining actions
- **WHEN** the second planned action fails
- **THEN** the system returns that error and does not invoke later planned actions

### Requirement: Configuration is distinct from file presence
A project SHALL be treated as compose-configured when it has a compose filename. File presence SHALL be a separate check and MUST NOT be required to produce a plan.

#### Scenario: Configured without touching disk
- **WHEN** a project has a compose filename and no directory is created
- **THEN** it is compose-configured and can appear in a plan

#### Scenario: Missing file is not a planning input
- **WHEN** a project has a compose filename whose file is absent
- **THEN** planning still accepts it as configured; the existence check is a separate result
