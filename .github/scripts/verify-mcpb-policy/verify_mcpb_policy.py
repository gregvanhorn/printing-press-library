#!/usr/bin/env python3
"""PR-time MCPB release policy checks.

This intentionally checks the cheap adversarial cases before signing ever
touches a release artifact:
- do not accept newly committed MCPB/native binary payloads under library/
- keep manifest.json pinned to the generated binary entrypoint shape

The manifest contract is also covered by verify-manifest; this script exists as
the release-signing policy guard and focuses on contributor-supplied payloads.
"""
from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]

BINARY_SUFFIXES = {
    ".mcpb",
    ".dylib",
    ".so",
    ".dll",
    ".exe",
}

MAGIC_PREFIXES = {
    b"\x7fELF": "ELF executable/shared object",
    b"MZ": "Windows PE executable",
    b"\xca\xfe\xba\xbe": "Mach-O universal binary",
    b"\xbe\xba\xfe\xca": "Mach-O universal binary",
    b"\xfe\xed\xfa\xce": "Mach-O executable",
    b"\xfe\xed\xfa\xcf": "Mach-O executable",
    b"\xce\xfa\xed\xfe": "Mach-O executable",
    b"\xcf\xfa\xed\xfe": "Mach-O executable",
}


def run(args: list[str]) -> str:
    return subprocess.check_output(args, cwd=REPO_ROOT, text=True, stderr=subprocess.DEVNULL)


def changed_library_paths() -> list[Path]:
    event_name = os.environ.get("GITHUB_EVENT_NAME")
    base_ref = os.environ.get("GITHUB_BASE_REF")

    if event_name == "pull_request" and base_ref:
        try:
            run(["git", "fetch", "--no-tags", "--depth=1", "origin", base_ref])
        except subprocess.CalledProcessError:
            pass
        try:
            base = run(["git", "merge-base", f"origin/{base_ref}", "HEAD"]).strip()
            diff = run(
                [
                    "git",
                    "diff",
                    "--name-only",
                    "--diff-filter=AMR",
                    f"{base}...HEAD",
                    "--",
                    "library/",
                ]
            )
            return [REPO_ROOT / line for line in diff.splitlines() if line]
        except subprocess.CalledProcessError:
            return []

    paths: set[Path] = set()
    for cmd in (
        ["git", "diff", "--name-only", "--diff-filter=AMR", "--", "library/"],
        ["git", "diff", "--cached", "--name-only", "--diff-filter=AMR", "--", "library/"],
        ["git", "ls-files", "--others", "--exclude-standard", "--", "library/"],
    ):
        try:
            paths.update(REPO_ROOT / line for line in run(cmd).splitlines() if line)
        except subprocess.CalledProcessError:
            continue
    return sorted(paths)


def binary_kind(path: Path) -> str | None:
    if path.suffix.lower() in BINARY_SUFFIXES:
        return f"{path.suffix} artifact"
    try:
        prefix = path.read_bytes()[:4]
    except OSError:
        return None
    for magic, label in MAGIC_PREFIXES.items():
        if prefix.startswith(magic):
            return label
    if prefix.startswith(b"PK\x03\x04") and path.name.endswith(".mcpb"):
        return "MCPB zip bundle"
    return None


def main() -> int:
    problems: list[str] = []
    for path in changed_library_paths():
        if not path.is_file():
            continue
        kind = binary_kind(path)
        if kind:
            rel = path.relative_to(REPO_ROOT)
            problems.append(
                f"::error file={rel}::Do not commit {kind} payloads under library/. "
                "MCPB and native binaries must be built from source in GitHub Actions before signing."
            )

    if problems:
        for problem in problems:
            print(problem)
        return 1

    print("No newly added or modified MCPB/native binary payloads under library/.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
