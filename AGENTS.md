# Codex Project Instructions

## Basic Rules
- Respond in Japanese.
- Ask concise questions when requirements are unclear.
- After work, summarize what changed and what the user can do next.
- Do not modify source code outside `doc/` unless the user explicitly asks for implementation.
- For commands that install libraries or mutate the development environment, explain the command first and ask before running it.

## Project Context
- Environment: WSL2 Ubuntu 24.04.3 in a VS Code dev container.
- The project is a learning/catch-up workspace.
- Frontend: Next.js.
- Backend: Gin.
- Database access: PostgreSQL via sqlc.
- Markdown files under `doc/` are project documentation and learning logs.

## Codex Agent Workflow
- The main learning mode is the custom Codex subagent `tdd_mentor`.
- Use `tdd_mentor` when the user asks to develop with TDD, asks for mentoring, or names `tdd-mentor` / `tdd_mentor`.
- `tdd_mentor` is intentionally read-only for source files. It should mentor, inspect, explain, and provide code snippets for the user to copy, but it should not patch implementation code.
- When the user asks for `ログ記録`, `学習ログを保存`, or `学習記録を作成`, delegate to `studylog_writer`.
- When the user asks for `タスク更新`, `task.md更新`, `設計書作成`, or `ドキュメント更新`, delegate to `doc_editor`.
- When the user asks to proceed from a GitHub Issue, delegate issue reading and requirement structuring to `tdd_issue_mentor`.

## Development Triggers
- `作業開始`: read the latest `doc/memo/` file, check whether there are follow-up tasks, and update `doc/task.md` through `doc_editor` if needed.
- `ログ記録`: finalize the current journal and write/update `doc/memo/YYYY/MM/DD_learning.md` through `studylog_writer`.

## Documentation Editing
- Documentation updates are allowed only under `doc/`.
- Preserve existing document structure and use minimal diffs.
- Never rewrite an existing learning log when appending a new session.
