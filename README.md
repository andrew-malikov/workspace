# WS

No hassle DX to manage dozen of microservices at work.

- [WS](#ws)
  - [Project Tracking](#project-tracking)
  - [Run Tests](#run-tests)
  - [Git Cleanup](#git-cleanup)
  - [Worktrees and DB Migrations](#worktrees-and-db-migrations)
  - [Development](#development)

## Project Tracking

First thing to do is to set up the microservice you intend to run locally or to check whether integration/ component tests are indeed green. It gets handy when you do, see the in the next chapter.

Add a new directory with a microservice into the tracker:

```bash
ws track <directory> [-a alias]
```

Or simply track the current one:

```bash
ws track
```

Remove already tracked directory from the tracker:

```bash
ws untrack <directory>
```

## Run Tests

There are three types of tests

- unit
- integration
- component

Unit tests doesn't require any prior setup while the last two do require to up docker compose. Usually it's no big deal BUT gotcha bangs you quick when you try to up several composes, ports are the same (of some services unfortunately) so running it all doesn't work out. So what you do is dancing with previous composes you touched.

OR you can use this one:

```bash
ws tests [-a alias]
```

It would check whether any other compose is up (across tracked ones) and down it.

## Git Cleanup

Lots of people make lots of code and thus lots of stale branches (me too by the way). Prints the stale branches to clean up (only author's branches):

```bash
ws git cleanup --dry-run
```

You can count your teammates:

```bash
ws git cleanup --dry-run --team
```

Running without `--dry` forces the show up plan first for you to review and then if you're okay execute:

```bash
ws git cleanup
```

The good part it spit out when the last commit was made and who's the author, which makes easier to communicate whether the branch is indeed stale.

## Worktrees and DB Migrations

Worktree support isn't in the current roadmap BUT I definitely would work on it as it would allow tracking mismatched DB migrations. Basically migrations get messed up when several people work on the same mircoservice and have their own new migrations, the apply order guaranteed only by `who deploys first`. The idea is to check across all the branches that the order is preserved against the current work branch.

## Development

To install dependencies:

```bash
bun install
```

To run:

```bash
bun run index.ts
```

This project was created using `bun init` in bun v1.2.15. [Bun](https://bun.sh) is a fast all-in-one JavaScript runtime.
