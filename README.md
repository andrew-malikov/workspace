# WS

No hassle DX to manage dozen of microservices at work.

- [WS](#ws)
  - [Install](#install)
  - [Project Tracking](#project-tracking)
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

## Run Tests

There are three types of tests

- unit
- integration
- component

Unit tests don't require any prior setup in opposite to the rest that do require to up docker compose. Usually it's no big deal BUT gotcha bangs you quick when you try to up several composes, ports are the same (of some services unfortunately) so running it all doesn't work out. So what you do is dancing with previous composes you touched.

OR you can use this one:

```bash
ws tests [-a alias]
```

It would check whether any other compose is up (across tracked ones) and down it.

## Git Cleanup

Lots of people make lots of code and thus lots of stale branches (me too by the way). Prints the stale branches to clean up (only branches owned by the current git user based on branch history):

```bash
ws clear
```

```md
| [ ] | branch           | author | updated at          | status |
| --- | ---------------- | ------ | ------------------- | ------ |
| [ ] | feat/glancy-list | you    | 2026-08-02 11:30 AM | synced |
| [x] | feat/remove-me   | you    | 2026-02-01 14:15 PM | local  |
| ... | ...              | ...    | ...                 | ...    |
| [x] | story/mocks      | you    | 2024-31-12 23:59 PM | remote |
```

You can count your teammates:

```bash
ws clear
```

```md
| [ ] | branch           | author       | updated at          | status |
| --- | ---------------- | ------------ | ------------------- | ------ |
| [ ] | feat/glancy-list | you          | 2026-08-02 11:30 AM | synced |
| [x] | feat/remove-me   | you          | 2026-02-01 14:15 PM | local  |
| ... | ...              | ...          | ...                 | ...    |
| [x] | hotfix/your-pc   | mr. anderson | 2021-01-01 00:00 AM | local  |
```

Running the command opens an interactive UI for review first, then lets you delete the selected branches:

```bash
ws clear
```

The good part it spit out when the last commit was made and who's the author, which makes easier to communicate whether the branch is indeed stale.

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
