# Project agent guide

<!-- One source of truth for every AI coding agent. AGENTS.md is read by Codex,
     Cursor, Gemini CLI, Copilot and others; CLAUDE.md is a symlink to this file. -->

## What this project is

(One or two lines: what it does, the stack, anything an agent must know first.)

## Conventions

- (Commit format, branch rules, how to run tests, etc.)

## Skills

Reusable skills live in `.agents/skills/` and are shared across all agents.
Workflow packages live in `.agents/workflows/`.
Basic setup does not create `docs/`; docs are owned by installed workflows.
