## Purpose

Give commands and workspace persistence explicit workspace, Compose, working-directory, and config-path values so CLI behavior can be exercised without process globals or fake binaries.

## ADDED Requirements

### Requirement: Persistence uses an explicit config path
Loading and saving the workspace SHALL take the config file path as an input. Home-directory and XDG resolution SHALL exist only as a path-resolution step, not inside load or save.

#### Scenario: Load from a given path
- **WHEN** the workspace is loaded from a provided file path
- **THEN** the result is that file's workspace and no home-directory environment is consulted

#### Scenario: Save to a given path
- **WHEN** the workspace is saved to a provided file path
- **THEN** that file receives the workspace and no home-directory environment is consulted

#### Scenario: Default path resolution
- **WHEN** the default config path is resolved
- **THEN** the result is the existing user config location derived from home and XDG rules

### Requirement: Commands receive session values
Every CLI command that needs a workspace, current directory, or Compose adapter SHALL receive those as injected session values. Command actions MUST NOT load the workspace from the default path, read the process working directory, or construct the real Docker Compose adapter themselves.

#### Scenario: Up uses injected session
- **WHEN** `ws up` runs with an injected workspace, cwd, and Compose adapter
- **THEN** it resolves and orchestrates against those values and does not search `PATH` for `docker`

#### Scenario: Down uses injected session
- **WHEN** `ws down` runs with an injected workspace, cwd, and Compose adapter
- **THEN** it resolves and stops against those values and does not search `PATH` for `docker`

#### Scenario: Non-compose commands use injected workspace and cwd
- **WHEN** a command such as `ws list` or `ws test` runs with an injected workspace and cwd
- **THEN** it uses those values and does not read home-directory config or the process working directory itself

### Requirement: Application edge supplies defaults
The process entrypoint SHALL resolve the default config path, load the workspace, read the process working directory, and construct the real Compose adapter, then pass those session values into commands.

#### Scenario: Production wiring
- **WHEN** the CLI is started normally
- **THEN** commands still operate on the user's real workspace, cwd, and Docker Compose

#### Scenario: Isolated application routing
- **WHEN** the application runner is invoked with an injected session
- **THEN** success and failure still follow existing stdout, stderr, and exit-code rules without requiring a fake `docker` binary or a changed process working directory
