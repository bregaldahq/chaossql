---
name: chaossql-skill-maintenance
description: Procedure for keeping the ChaosSQL agent skills, docs/harness/flow-map.md and CLAUDE.md in sync with the code. Use at the end of every change, when adding or renaming files, when creating a new flow or skill, or when make check-harness reports a skill problem.
---

# Skill Maintenance

The skills under `.claude/skills/` are part of the product's harness. They are
only useful if they describe the code as it is. `CLAUDE.md` makes updating them
mandatory in the same commit as the change that affects them.

## When to use

- Before finishing any task that modified code, config, workflows or docs.
- When `make check-harness` prints `❌ [SKILL] ...`.
- When a new flow appears (new command, package, protocol, workflow, surface).

## Sync procedure (every change)

1. List the paths you touched (`git diff --name-only main...`).
2. Find the owning skills in `docs/harness/flow-map.md` (primary paths) and by
   searching the skills: `grep -rl "<path or symbol>" .claude/skills`.
3. For each affected skill, update:
   - the behavior description (steps, defaults, precedence, formats);
   - **Gotchas** — add newly discovered pitfalls, delete fixed ones;
   - **Change checklist** — new places that must move together;
   - **Tests** — new or renamed test files;
   - **Source map** — added, renamed or deleted files.
4. If the change creates a new flow, create a skill (template below) and add it
   to `docs/harness/flow-map.md` **and** the index in `CLAUDE.md`.
5. If the change removes a flow, delete the skill directory and both index entries.
6. Run `make check-harness`.

Behavior claims in a skill must be verified against the code, never copied from
older docs, specs, or planning ledgers — those can be stale.

## What `make check-harness` enforces

Implemented in `tools/harness_check.go` (`checkSkills`):

- each `.claude/skills/<dir>/SKILL.md` starts with YAML frontmatter whose
  `name:` equals `<dir>` and whose `description:` is non-empty;
- each skill has a `## Source map` section;
- every backticked token in `## Source map` that looks like a repository path
  exists (tokens containing spaces, `*`, `$`, `:`, `(`, `=`, quotes, or starting
  with `/` or `-` are ignored, so keep one plain path per backtick);
- every skill directory is referenced as `` `chaossql-...` `` in both
  `docs/harness/flow-map.md` and `CLAUDE.md`, and every reference there exists.

It does **not** verify prose accuracy — that is on you.

## Skill template

```markdown
---
name: chaossql-<flow>
description: <What the flow does>. Use when <concrete triggers>.
---

# <Title>

<One paragraph: what this flow is and why it exists.>

## When to use
## How it works            (numbered steps, defaults, precedence, formats)
## Contracts and invariants
## Gotchas                 (verified, non-obvious facts)
## Change checklist        (what must move together)
## Tests                   (files and commands)
## Source map              (one backticked repository path per bullet)
## Related skills
```

Naming: `chaossql-` prefix, kebab-case, one flow per skill. Prefer splitting a
skill over letting it cover unrelated flows.

## Writing rules

- English only (the purity gate does not scan `.claude/`, but `CLAUDE.md`
  requires it).
- Name concrete functions, flags, env vars, file formats and defaults.
- State defaults exactly as the code does (for example "workers default to 4
  in the runner but 2 in `chaossql engine`").
- Link sibling skills instead of duplicating their content.
- Commits never carry `Co-Authored-By` or AI attribution lines.

## Source map

- `.claude/skills/`
- `docs/harness/flow-map.md`
- `CLAUDE.md`
- `tools/harness_check.go`
- `Makefile`

## Related skills

- `chaossql-harness-overview`
- `chaossql-quality-gate`
