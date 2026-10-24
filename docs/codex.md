# Developing GoHP with Codex

Use the existing GoHP checkout in the isolated Codex task. Do not create an
additional Git worktree unless requested. This setup is for GoHP; Gossip has a
separate development workflow.

The development environment is defined by the repository's `Dockerfile`,
`compose.yaml`, and `scripts/check.sh`. They pin Go 1.23.12 and the PHP 8.3.30 /
7.4.33 reference images by digest. A host Go or PHP installation is not required.
Docker Engine and Docker Compose v2 must be available to the Codex environment.

## Setup command

Run from the checkout, or use the following as the Codex setup script when the
repository is mounted at `/workspace/GoHP`:

```sh
set -eu
cd /workspace/GoHP
sh scripts/codex-setup.sh
```

The script builds both pinned images and runs the existing full validation for
both PHP versions. It uses the `gohp-codex` Compose project name, preserves the
underlying command's failure status, and removes its one-off containers after
execution. It does not edit repository files or initialize a new project.
No background application service needs to be started.
Buildx metadata defaults to the ignored `.cache/docker-buildx` directory so a
read-only home directory works; an existing `BUILDX_CONFIG` is respected and
Docker authentication configuration is preserved.

Image downloads require the standard Docker registry HTTPS endpoints. No
application credentials, database, Composer packages, or production services
are needed. Tests run with networking disabled, a read-only root filesystem,
temporary build/cache storage, and the restrictions in `compose.yaml`.

## Validation during development

After changing compiler/runtime code or fixtures, rerun:

```sh
sh scripts/codex-setup.sh
```

Rebuilding is necessary because the images contain a copy of the source tree;
running an old image alone does not test current edits. Docker reuses unchanged
image layers. The PHP 8.3 and 7.4 jobs run sequentially.

Each job runs `go vet`, all existing Go tests, CLI compilation, PHP-to-Go source
generation, native compilation, and a byte-for-byte comparison with reference
PHP output. Positive fixtures cover functions/classes/arrays and version-aware
operator precedence. Unsupported-source and runtime-failure checks must also
execute. A successful compile alone is not a passing differential test.

The generated binary is also tested without PHP in its PATH. These checks cover
the documented Phase 0 subset, not arbitrary PHP, Composer packages, or
WordPress. See [the compatibility matrix](compatibility.md) and
[the implementation report](phase0-report.md).

Keep changes on a feature branch and follow [CONTRIBUTING.md](../CONTRIBUTING.md).
Do not merge to `main`, deploy, release, or finalize licensing without the
owner's approval. Preserve existing user work and pinned dependency files.

## 日本語概要

GoHP のチェックアウトで `sh scripts/codex-setup.sh` を実行すると、固定済みの
Go / PHP イメージをビルドし、PHP 8.3 と 7.4 の両方で既存の差分テストを実行します。
Docker Engine と Compose v2 が必要です。コード変更後は同じコマンドを再実行して、
最新コードを含むイメージで検証してください。Gossip は対象に含みません。
