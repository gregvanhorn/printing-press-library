import { readFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { installCommand } from "./commands/install.js";
import { listCommand } from "./commands/list.js";
import { searchCommand } from "./commands/search.js";
import { uninstallCommand } from "./commands/uninstall.js";
import { updateCommand } from "./commands/update.js";

type CommandHandler = (args: string[]) => Promise<number>;

const COMMANDS: Record<string, CommandHandler> = {
  install: installCommand,
  update: updateCommand,
  list: listCommand,
  search: searchCommand,
  uninstall: uninstallCommand,
};

export async function main(args = process.argv.slice(2)): Promise<void> {
  const code = await run(args);
  if (code !== 0) {
    process.exitCode = code;
  }
}

export async function run(args: string[]): Promise<number> {
  const [command, ...rest] = args;

  if (!command || command === "-h" || command === "--help") {
    printHelp();
    return 0;
  }

  if (command === "-v" || command === "--version") {
    console.log(await packageVersion());
    return 0;
  }

  const handler = COMMANDS[command];
  if (!handler) {
    console.error(`Unknown command: ${command}`);
    printHelp();
    return 1;
  }

  return handler(rest);
}

function printHelp(): void {
  console.log(`Printing Press CLI installer

Usage:
  pp <command> [options]

Commands:
  install <name>     Install a Printing Press CLI and skill
  update [name]      Refresh one installed CLI, or all installed CLIs
  list               List installed Printing Press CLIs
  search <query>     Search the Printing Press catalog
  uninstall <name>   Remove a Printing Press CLI and skill

Options:
  -h, --help         Show help
  -v, --version      Show version`);
}

async function packageVersion(): Promise<string> {
  let dir = dirname(fileURLToPath(import.meta.url));
  for (let i = 0; i < 5; i++) {
    try {
      const data = await readFile(join(dir, "package.json"), "utf8");
      const parsed = JSON.parse(data) as { version?: string };
      if (parsed.version) {
        return parsed.version;
      }
    } catch {
      // Walk up until we find the package root.
    }
    dir = dirname(dir);
  }
  return "0.0.0";
}
