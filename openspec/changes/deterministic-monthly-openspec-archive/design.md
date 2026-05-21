# Design

## Context

The monthly archive workflow already has deterministic candidate detection and post-archive assertions. The weak point is the base archive execution: it relies on Codex, Claude Code, or GitHub Copilot CLI to interpret the prompt and run `lina-auto-archive`. In the failed GitHub Actions run, GitHub Copilot CLI returned success but completed active changes remained under `openspec/changes/`.

A local replay at the failed commit showed that direct `openspec archive -y` can archive most candidates, while `remove-sqlite-support` fails because one REMOVED requirement header does not exist in the current baseline spec. This is a good fit for deterministic automation: the workflow can archive everything that OpenSpec can apply, report exact failures, open a PR for successful archives, and then fail to force maintainers to address remaining blockers.

## Approach

Add a shared composite action `.github/actions/monthly-openspec-auto-archive` that:

- Runs `openspec list --json` with the pinned CLI version.
- Selects active changes with status `complete`, `completed`, or `done`.
- Treats mismatched `completedTasks` / `totalTasks` counts as an archive failure when OpenSpec reports task counts.
- Runs `openspec archive -y <change>` for each candidate in a stable order.
- Rechecks `openspec list --json` after each archive command so a candidate is only considered archived after it leaves the active change list.
- Records successful and failed changes in JSON outputs and the job summary.
- Outputs `had-failures=true` when any archive command fails or when a candidate remains active after the command returns.

The tool-specific reusable workflows will call this deterministic action before AI runtime setup. They will then detect archive diffs and only prepare/run Codex, Claude Code, or GitHub Copilot CLI for consolidation when deterministic archiving produced changes. Final PR creation should happen before a final failure gate, so partially successful archive runs still write an archive PR while the job fails if any completed changes were not archived.

The existing `monthly-openspec-assert-archive-complete` action stays as the final gate, but it moves after PR finalization and runs only when deterministic archiving reported failures. That preserves a failing job for unresolved blockers and keeps the summary of remaining completed active changes.

## Scope

This is CI/OpenSpec governance only. It does not change runtime product behavior, HTTP APIs, backend Go production code, data permissions, runtime i18n, or cache behavior.

## Known Blocker Fix

`remove-sqlite-support` currently contains a REMOVED delta for a requirement title that is not present in `openspec/specs/cluster-coordination-config/spec.md`. The current baseline already expresses the PostgreSQL-only behavior in `Requirement: 非 PostgreSQL 数据库链接必须在 coordination 启动前失败`, so the delta should be changed from REMOVED to MODIFIED against that existing requirement.
