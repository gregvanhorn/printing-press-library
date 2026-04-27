import test from "node:test";
import assert from "node:assert/strict";
import { detectGo, goInstall } from "../src/go.js";
import type { Runner } from "../src/process.js";

test("detectGo returns installed version when go is available", async () => {
  const runner: Runner = async () => ({
    code: 0,
    stdout: "go version go1.23.4 darwin/arm64\n",
    stderr: "",
  });

  assert.deepEqual(await detectGo(runner), { installed: true, version: "1.23.4" });
});

test("detectGo returns installed false when go is unavailable", async () => {
  const runner: Runner = async () => ({
    code: 127,
    stdout: "",
    stderr: "go: command not found",
  });

  assert.deepEqual(await detectGo(runner), { installed: false });
});

test("goInstall invokes go install with the requested ref", async () => {
  const calls: Array<{ command: string; args: string[] }> = [];
  const runner: Runner = async (command, args) => {
    calls.push({ command, args });
    return { code: 0, stdout: "", stderr: "" };
  };

  await goInstall("github.com/mvanhorn/printing-press-library/library/sports/espn/cmd/espn-pp-cli", {
    ref: "main",
    runner,
  });

  assert.deepEqual(calls, [
    {
      command: "go",
      args: [
        "install",
        "github.com/mvanhorn/printing-press-library/library/sports/espn/cmd/espn-pp-cli@main",
      ],
    },
  ]);
});
