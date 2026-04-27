import { detectGo, goInstall, type GoDetection } from "../go.js";
import { commandOnPath, type RunResult } from "../process.js";
import {
  cliBinaryName,
  cliSkillName,
  DEFAULT_REGISTRY_URL,
  fetchGoModulePath,
  fetchRegistry,
  lookupByName,
  type Registry,
} from "../registry.js";
import { installSkill } from "../skill.js";

interface InstallOptions {
  agents: string[];
  json: boolean;
  registryUrl: string;
}

interface InstallDeps {
  fetchRegistry: (url: string) => Promise<Registry>;
  resolveModulePath: (entryPath: string, registryUrl: string) => Promise<string | null>;
  detectGo: () => Promise<GoDetection>;
  goInstall: (modulePath: string, ref: string, env?: NodeJS.ProcessEnv) => Promise<RunResult>;
  commandOnPath: (binary: string) => Promise<string | null>;
  installSkill: (skillName: string, agents: string[]) => Promise<RunResult>;
  stdout: (message: string) => void;
  stderr: (message: string) => void;
  platform: NodeJS.Platform;
}

interface InstallResult {
  name: string;
  binary: string;
  modulePath: string;
  skill: string;
  binaryPath: string;
  authEnvVars: string[];
}

export function createInstallCommand(overrides: Partial<InstallDeps> = {}) {
  const deps: InstallDeps = {
    fetchRegistry: (url) => fetchRegistry(url),
    resolveModulePath: (entryPath, registryUrl) => fetchGoModulePath(entryPath, registryUrl),
    detectGo: () => detectGo(),
    goInstall: (modulePath, ref, env) => goInstall(modulePath, { ref, env }),
    commandOnPath: (binary) => commandOnPath(binary),
    installSkill: (skillName, agents) => installSkill(skillName, { agents }),
    stdout: (message) => console.log(message),
    stderr: (message) => console.error(message),
    platform: process.platform,
    ...overrides,
  };

  return async function installCommandWithDeps(args: string[]): Promise<number> {
    const parsed = parseInstallArgs(args);
    if ("error" in parsed) {
      deps.stderr(parsed.error);
      deps.stderr("Usage: pp install <name> [--agent <agent>...] [--json]");
      return 1;
    }

    const { name, options } = parsed;

    try {
      const registry = await deps.fetchRegistry(options.registryUrl);
      const entry = lookupByName(registry, name);
      if (!entry) {
        deps.stderr(`No Printing Press CLI found for "${name}". Try \`pp search ${name}\`.`);
        return 1;
      }

      const go = await deps.detectGo();
      if (!go.installed) {
        deps.stderr(goMissingMessage(deps.platform));
        return 1;
      }

      const binary = cliBinaryName(entry);
      const moduleRoot =
        (await deps.resolveModulePath(entry.path, options.registryUrl)) ??
        `github.com/mvanhorn/printing-press-library/${entry.path}`;
      const modulePath = `${moduleRoot}/cmd/${binary}`;
      const skillName = cliSkillName(entry);

      const install = await installGoWithFallback(deps, modulePath);
      if (install.code !== 0) {
        deps.stderr(`go install failed for ${modulePath}`);
        if (install.stderr.trim()) {
          deps.stderr(install.stderr.trim());
        }
        return 1;
      }

      const binaryPath = await deps.commandOnPath(binary);
      if (!binaryPath) {
        deps.stderr(pathMessage(binary));
        return 1;
      }

      const skill = await deps.installSkill(skillName, options.agents);
      if (skill.code !== 0) {
        deps.stderr(`Skill install failed for ${skillName}. The binary remains installed at ${binaryPath}.`);
        if (skill.stderr.trim()) {
          deps.stderr(skill.stderr.trim());
        }
        return 1;
      }

      const result: InstallResult = {
        name: entry.name,
        binary,
        modulePath,
        skill: skillName,
        binaryPath,
        authEnvVars: entry.mcp?.env_vars ?? [],
      };

      if (options.json) {
        deps.stdout(JSON.stringify({ ok: true, ...result }, null, 2));
      } else {
        deps.stdout(`Installed ${entry.name}`);
        deps.stdout(`  binary: ${binaryPath}`);
        deps.stdout(`  skill: ${skillName}`);
        if (result.authEnvVars.length > 0) {
          deps.stdout(`  auth env vars: ${result.authEnvVars.join(", ")}`);
        }
      }

      return 0;
    } catch (error) {
      deps.stderr(error instanceof Error ? error.message : String(error));
      return 1;
    }
  };
}

export const installCommand = createInstallCommand();

function parseInstallArgs(args: string[]):
  | { name: string; options: InstallOptions }
  | { error: string } {
  const options: InstallOptions = {
    agents: [],
    json: false,
    registryUrl: DEFAULT_REGISTRY_URL,
  };
  let name: string | undefined;

  for (let i = 0; i < args.length; i++) {
    const arg = args[i]!;
    if (arg === "--json") {
      options.json = true;
    } else if (arg === "--agent" || arg === "-a") {
      const agent = args[++i];
      if (!agent) {
        return { error: "Missing value for --agent" };
      }
      options.agents.push(agent);
    } else if (arg === "--registry-url") {
      const registryUrl = args[++i];
      if (!registryUrl) {
        return { error: "Missing value for --registry-url" };
      }
      options.registryUrl = registryUrl;
    } else if (arg.startsWith("-")) {
      return { error: `Unknown install option: ${arg}` };
    } else if (!name) {
      name = arg;
    } else {
      return { error: `Unexpected argument: ${arg}` };
    }
  }

  if (!name) {
    return { error: "Missing CLI name" };
  }

  return { name, options };
}

async function installGoWithFallback(deps: InstallDeps, modulePath: string): Promise<RunResult> {
  const latest = await deps.goInstall(modulePath, "latest");
  if (latest.code === 0) {
    return latest;
  }

  const env = {
    GOPRIVATE: "github.com/mvanhorn/*",
    GOFLAGS: "-mod=mod",
  };
  const main = await deps.goInstall(modulePath, "main", env);
  return main.code === 0 ? main : combineFailures(latest, main);
}

function combineFailures(latest: RunResult, main: RunResult): RunResult {
  return {
    code: main.code || latest.code || 1,
    stdout: [latest.stdout, main.stdout].filter(Boolean).join("\n"),
    stderr: [
      "go install @latest failed:",
      latest.stderr.trim(),
      "go install @main fallback failed:",
      main.stderr.trim(),
    ]
      .filter(Boolean)
      .join("\n"),
  };
}

function goMissingMessage(platform: NodeJS.Platform): string {
  const installHint =
    platform === "darwin"
      ? "Install Go with: brew install go"
      : platform === "win32"
        ? "Install Go with: winget install GoLang.Go"
        : "Install Go from your package manager or https://go.dev/dl/";
  return `Go is required to install Printing Press CLIs. ${installHint}`;
}

function pathMessage(binary: string): string {
  return `${binary} was installed, but it is not on PATH. Add $(go env GOPATH)/bin, usually $HOME/go/bin, to PATH and retry.`;
}
