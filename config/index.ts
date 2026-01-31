import { Effect } from "effect";
import os from "os";
import { readFile, mkdir, writeFile } from "fs/promises";
import path from "path";

export type Project = {
  alias: string;
  dir: string;
  compose?: string;
  migrations?: string;
};

export type Config = {
  projects: Project[];
};

const EMPTY_CONFIG: Readonly<Config> = {
  projects: [],
};

const DEFAULT_UNIX_CONFIG_DIR = "~/.config";

export function LoadConfig(): Effect.Effect<Config, Error, never> {
  let configDir = "";
  switch (os.type()) {
    case "Darwin":
    case "Linux":
      configDir = process.env.XDG_CONFIG_HOME ?? DEFAULT_UNIX_CONFIG_DIR;
      break;

    default:
      return Effect.dieMessage(`OS ${os.type()} isn't supported yet.`);
  }

  configDir = path.join(configDir, "ws");

  return Effect.gen(function* () {
    const configPath = path.join(configDir, "config.json");
    const readConfig = Effect.promise(() =>
      readFile(configPath, { encoding: "utf-8" }),
    );

    const config = yield* Effect.orElse(readConfig, () =>
      Effect.gen(function* () {
        const serializedConfig = yield* Effect.try(() =>
          JSON.stringify(EMPTY_CONFIG),
        );
        yield* Effect.promise(() => mkdir(configDir, { recursive: true }));
        yield* Effect.promise(() => writeFile(configPath, serializedConfig));
        return serializedConfig;
      }),
    );

    return JSON.parse(config) as Config;
  });
}
