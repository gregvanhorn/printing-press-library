# @mvanhorn/printing-press

Installer and catalog CLI for Printing Press-generated CLIs.

## Release status

This package is prepared for v0.1.0, but public use waits on the library PR being merged, the npm package being published, and the catalog repository being public. While the repository is private, set `GITHUB_TOKEN` or `GH_TOKEN` and make sure Go can read private `github.com/mvanhorn/*` modules.

## Install a CLI

```bash
npx -y @mvanhorn/printing-press install espn
```

`pp install <name>` reads the live catalog from `registry.json`, resolves the CLI's Go module path, runs `go install`, and installs the matching skill from `cli-skills/pp-<name>` using `skills@latest`.

Useful commands:

```bash
npx -y @mvanhorn/printing-press search sports
npx -y @mvanhorn/printing-press list
npx -y @mvanhorn/printing-press update espn
npx -y @mvanhorn/printing-press uninstall espn --yes
```

## Options

```bash
npx -y @mvanhorn/printing-press install espn --agent claude-code
npx -y @mvanhorn/printing-press install espn --json
npx -y @mvanhorn/printing-press search sports --registry-url https://example.com/registry.json
```

`--agent` can be repeated to constrain skill installation to specific agents. `--registry-url` is mainly for testing alternate catalogs.
