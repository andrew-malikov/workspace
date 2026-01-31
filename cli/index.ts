import yargs from "yargs";
import { hideBin } from "yargs/helpers";
import { TrackCommand } from "@/commands/track";

yargs()
  .scriptName("ws")
  .command(TrackCommand)
  .demandCommand()
  .help()
  .parse(hideBin(process.argv));
