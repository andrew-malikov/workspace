import zod from "zod";

export const options = zod.object({
  name: zod.string().describe("Your name"),
});

type Props = {
  options: zod.infer<typeof options>;
};

export function Index({ options }: Props) {}

export const TrackCommand = {
  command: "track [dir]",
  describe: "start tracking directory",
  builder: {
    compose: {
      alias: "c",
      describe: "filepath to docker-compose.yaml",
      default: "docker-compose.yaml",
    },
    migrations: {
      alias: "m",
      describe: "directory with db migrations",
      default: "migrations",
    },
  },
  handler: (argv: any) => {},
};

function LoadConfig() {
  
}