import { execFileRunner, type Runner } from "./process.js";

export interface GoDetection {
  installed: boolean;
  version?: string;
}

export async function detectGo(runner: Runner = execFileRunner): Promise<GoDetection> {
  const result = await runner("go", ["version"]);
  if (result.code !== 0) {
    return { installed: false };
  }

  const match = result.stdout.match(/\bgo version go(\S+)/);
  return {
    installed: true,
    version: match?.[1],
  };
}

export interface GoInstallOptions {
  ref?: string;
  env?: NodeJS.ProcessEnv;
  runner?: Runner;
}

export async function goInstall(modulePath: string, options: GoInstallOptions = {}) {
  const runner = options.runner ?? execFileRunner;
  const ref = options.ref ?? "latest";
  return runner("go", ["install", `${modulePath}@${ref}`], { env: options.env });
}
