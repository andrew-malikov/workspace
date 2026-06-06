# WS

No hassle DX to manage dozen of microservices at work.

- [WS](#ws)
  - [Install](#install)
  - [Project Tracking](#project-tracking)
  - [Running Containers](#running-containers)
  - [Run Tests](#run-tests)
  - [Git Cleanup](#git-cleanup)
  - [Worktrees and DB Migrations](#worktrees-and-db-migrations)
  - [Development](#development)
  - [Versioning](#versioning)

## Install

Homebrew:

```sh
brew tap andrew-malikov/tap
brew install --cask andrew-malikov/tap/ws
```

## Project Tracking

First thing to do is to set up the microservice you intend to run locally or to check whether integration/ component tests are indeed green. It gets handy when you do, see the in the next chapter.

Add a new directory with a microservice into the tracker:

```bash
ws track <alias> [directory]
```

Remove already tracked directory from the tracker:

```bash
ws untrack [directory | alias]
```

## Running Containers

List tracked projects with currently running docker compose services:

```bash
ws ctrs
```

Start docker compose for the current tracked project:

```bash
ws up
```

Start docker compose for a tracked project by alias:

```bash
ws up <alias>
```

By default, `ws up` stops any other tracked project with running docker compose services before starting the target project. Use `--alongside` to keep other compose projects active:

```bash
ws up <alias> --alongside
```

Use `--blank` or `-b` to remove the target project's docker compose volumes before starting it:

```bash
ws up <alias> --blank
ws up <alias> -b
```

## Run Tests

There are three types of tests

- unit
- integration
- component

Unit tests don't require any prior setup in opposite to the rest that do require to up docker compose. Usually it's no big deal BUT gotcha bangs you quick when you try to up several composes, ports are the same (of some services unfortunately) so running it all doesn't work out. So what you do is dancing with previous composes you touched.

OR you can use this one:

```bash
ws test
ws test -u
ws test -ic
```

It would check whether any other compose is up (across tracked ones) and down it.

Configure global test discovery rules in `~/.config/ws/config.toml` and let `ws track` resolve the concrete test project paths for each tracked project:

```toml
[test.unit]
project_patterns = ["(^|/)tests/.+UnitTests\\.csproj$"]
filter = "Category=Unit"

[test.integration]
project_patterns = ["(^|/)tests/.+IntegrationTests\\.csproj$"]
filter = "Category=Integration"

[test.component]
project_patterns = ["(^|/)tests/.+ComponentTests\\.csproj$"]
filter = "Category=Component"
```

## Git Cleanup

Lots of people make lots of code and thus lots of stale branches (me too by the way). Prints the stale branches to clean up (only branches owned by the current git user based on branch history):

```bash
ws clear
```

## Worktrees and DB Migrations

Worktree support isn't in the current roadmap BUT I definitely would work on it as it would allow tracking mismatched DB migrations. Basically migrations get messed up when several people work on the same mircoservice and have their own new migrations, the apply order guaranteed only by `who deploys first`. The idea is to check across all the branches that the order is preserved against the current work branch.

## Development

Build:

```sh
mise build
```

Run:

```sh
./binaries/ws
```

Debug:

```sh
dlv exec binaries/ws
```

## Versioning

Use the command:

```sh
mise tag
```

It checks the latest tag and appends a new one in format `yyyy.mm.dd-n` where `n` is an incrementing build number per a day.
