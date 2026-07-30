## ADDED Requirements

### Requirement: Down resolves a tracked Compose project
The CLI SHALL provide `ws down [alias]` and SHALL resolve its target using the same tracked-project rules as `ws up`: an omitted alias selects the project tracked at the current directory, and a supplied alias selects that tracked project.

#### Scenario: Down current project
- **WHEN** the user runs `ws down` from a tracked project with a configured Compose file
- **THEN** the CLI stops that project's Compose stack

#### Scenario: Down project by alias
- **WHEN** the user runs `ws down api` and `api` is a tracked project with a configured Compose file
- **THEN** the CLI stops the `api` Compose stack regardless of the current directory

#### Scenario: Down target is not tracked
- **WHEN** the user runs `ws down` without a resolvable tracked target or supplies an unknown alias
- **THEN** the CLI reports the resolution failure, invokes no Compose action, and exits nonzero

#### Scenario: Down target has no Compose configuration
- **WHEN** the resolved project has no existing configured Compose file
- **THEN** the CLI reports the missing Compose configuration, invokes no Compose action, and exits nonzero

### Requirement: Default down preserves volumes
The CLI SHALL implement plain `ws down` with one `docker compose down` action and SHALL NOT request volume removal.

#### Scenario: Stop while preserving volumes
- **WHEN** the user runs `ws down` without `--blank`
- **THEN** the CLI invokes `docker compose down` once for the target and leaves its Compose volumes intact

### Requirement: Blank down removes volumes
The CLI SHALL accept `--blank` and `-b` on `ws down` and SHALL implement either form with one `docker compose down -v` action rather than separate stop and cleanup actions.

#### Scenario: Long blank flag
- **WHEN** the user runs `ws down --blank`
- **THEN** the CLI stops the target and removes its Compose volumes with one cleanup action

#### Scenario: Short blank flag
- **WHEN** the user runs `ws down -b`
- **THEN** the CLI performs the same cleanup action as `--blank`

### Requirement: Down reports only completed work
The CLI SHALL render a successful down result after the selected Compose action completes and SHALL identify whether volumes were removed. If the Compose action fails, the CLI MUST return nonzero and MUST NOT render the success result.

#### Scenario: Default down succeeds
- **WHEN** the target's default down action completes successfully
- **THEN** the CLI reports that the target Compose project is down without claiming volume removal

#### Scenario: Blank down succeeds
- **WHEN** the target's blank down action completes successfully
- **THEN** the CLI reports that the target Compose project is down and that its volumes were removed

#### Scenario: Down action fails
- **WHEN** Docker Compose returns an error while stopping or cleaning the target
- **THEN** the CLI preserves the failure, exits nonzero, and emits no successful down result
