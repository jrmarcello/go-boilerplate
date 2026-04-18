# Harness self-steering

This guide describes **how the gopherplate harness evolves over time**. The harness inventory in
[docs/harness.md](../harness.md) is a snapshot; this document is the process that keeps it
useful.

The term "self-steering" comes from Martin Fowler's article
["Harness Engineering for Coding Agents"](https://martinfowler.com/articles/harness-engineering.html):

> When issues recur, humans iterate on the harness itself — improving feedforward and feedback
> controls so the class of problem stops coming back. Coding agents can help build the new
> controls (tests, linters, rules, docs).

The three loops below turn that principle into something executable in this repo.

## When to open a harness gap note

Open a gap note whenever **the harness failed to prevent or detect a problem it should have**.
Concrete triggers:

1. **A bug escaped to production** that a sensor could reasonably have caught (e.g., wrong HTTP
   status returned, response field missing, N+1 query, race condition on a write path).
2. **The Stop hook (`stop-validate.sh`) failed the same class of error three times in a week**
   — that is a signal the class deserves a pre-edit guide or an on-edit sensor, not just a late
   gate.
3. **A human reviewer caught something a harness could check** (unused receiver, handler not
   using `httpgin.SendSuccess`, missing span classification). Human review should not be the
   first line of defense for mechanical rules.
4. **A business metric (SLO, error rate, tail latency) degraded** without any sensor firing.
   Either a sensor is missing, or it exists but is not wired to an alert.
5. **A contributor asks "where should this rule live?" and the answer is unclear** — the harness
   has a coherence gap; pick a home and document it.
6. **A guide (CLAUDE.md, `.claude/rules/*.md`, ADR) is out of sync with the code** — either the
   guide lies or the code drifted. Either way, add a sensor that would have caught the drift.

## Gap note template

Store gap notes in a team-agreed location (issue tracker, Confluence page, or a
`docs/harness-gaps/` directory if we want them in-repo). Use this shape:

```markdown
---
date: YYYY-MM-DD
author: <who>
status: open | in-progress | resolved | wontfix
---

### Symptom

What was observed. One paragraph. Link to the incident, PR, or Stop-hook log if applicable.

### Category

One of: `maint` / `arch-fitness` / `behavior` / `meta`.

### Proposed control

Describe the guide or sensor. Be concrete:

- Kind: `guide` (feedforward) or `sensor` (feedback)?
- Execution: `computational` (linter, test, script) or `inferential` (skill, subagent)?
- Stage: `pre-commit` / `on-edit` / `stop-hook` / `CI` / `post-integration` / `continuous` /
  `scaffold-time` / `review-time`.
- Home: `.claude/rules/*.md` / `.claude/skills/*/` / `.claude/hooks/*.sh` /
  `.golangci.yml` / `.semgrep/*.yml` / `.github/workflows/*.yml` / `cmd/cli/*` /
  `docs/*`.

### Alternatives considered

Why this control and not another. E.g., "linter vs. semgrep vs. architecture test" — note
trade-offs.

### Cost estimate

Rough time to implement + ongoing cost (CPU, runner minutes, dev feedback latency).

### References

- Incident link / PR link / Stop-hook log snippet.
- Related ADR or harness spec if any.
```

## Where new controls live

Match the control to the closest existing home. Only invent a new home when nothing fits.

| Control type | Home |
| --- | --- |
| Project-wide narrative convention (when to use X pattern) | [CLAUDE.md](../../CLAUDE.md) |
| Targeted rule auto-applied by file pattern | [.claude/rules/](../../.claude/rules/) |
| Architectural decision with long-term context | [docs/adr/](../adr/) |
| Workflow that an agent runs on demand | [.claude/skills/](../.claude/skills/) |
| Specialized inferential reviewer | [.claude/agents/](../../.claude/agents/) |
| Deterministic check on edit | [.claude/hooks/](../../.claude/hooks/) |
| Deterministic check at commit/push | [lefthook.yml](../../lefthook.yml) |
| Go static analysis | [.golangci.yml](../../.golangci.yml) |
| Team-specific code pattern check | `.semgrep/` (planned, see behavior-harness spec) |
| Continuous pre-production check | `.github/workflows/` |
| Performance fitness function | `tests/load/` + `.github/workflows/` |
| Architectural fitness (breaking changes, etc.) | contract tools (buf breaking, etc.) |
| Scaffolder / new-service bootstrapping | [cmd/cli/](../../cmd/cli/) |

If the answer is "it could go in two places", open the gap note as **meta** and pick one home
with rationale.

## Coherence check — monthly review

The harness grows over time. Without periodic review, guides drift out of sync with sensors,
and duplicate or contradictory rules accumulate. Run this checklist monthly (≈30 min).

- [ ] **Inventory accuracy**: walk [docs/harness.md](../harness.md) and confirm every row still
      maps to an existing artifact. Remove rows for deleted skills/hooks/linters.
- [ ] **Missing rows**: grep the repo for new skills, hooks, linters, workflows added since the
      last review (`git log --since="1 month ago" --name-only -- .claude/ .github/
      .golangci.yml lefthook.yml`). Add them to the inventory.
- [ ] **Open gap notes**: review all `open` and `in-progress` gap notes. Has the cost estimate
      held? Is the control still the right one?
- [ ] **Duplicate controls**: any rule expressed in two places with a risk of diverging?
      Typical example: a check in both `golangci-lint` and a `.claude/rules/` rule. Consolidate.
- [ ] **Dead controls**: any rule or sensor that hasn't fired in 3+ months? Either codebase
      outgrew it, or it is redundant. Either way, note for removal.
- [ ] **Coverage vs. category**: does any category (maint / arch-fitness / behavior / meta)
      feel under-instrumented? If all our sensors are about maintainability and none about
      behavior, we have a blind spot.
- [ ] **Feedback loop latency**: are we catching classes of bugs too late (post-integration) that
      could be caught earlier (on-edit)? Push sensors left where feasible.

Log the review outcome in the same place you keep gap notes, so future reviews can see what was
changed and why.

## The self-steering loop in one diagram

```text
  ┌────────────────────────────────────────────┐
  │  Symptom observed                          │
  │  (bug in prod / recurrent hook failure /   │
  │   human review catch / SLO drift)          │
  └───────────────────┬────────────────────────┘
                      │
                      ▼
  ┌────────────────────────────────────────────┐
  │  Open gap note (template above)            │
  │  Classify: guide vs. sensor, category,     │
  │  execution type, home.                     │
  └───────────────────┬────────────────────────┘
                      │
                      ▼
  ┌────────────────────────────────────────────┐
  │  Implement the control                     │
  │  (often via /spec + /ralph-loop for        │
  │   non-trivial additions)                   │
  └───────────────────┬────────────────────────┘
                      │
                      ▼
  ┌────────────────────────────────────────────┐
  │  Update docs/harness.md inventory          │
  │  Close the gap note                        │
  └───────────────────┬────────────────────────┘
                      │
                      ▼
  ┌────────────────────────────────────────────┐
  │  Monthly coherence check                   │
  │  Prune, consolidate, shift-left            │
  └────────────────────────────────────────────┘
```

## Non-goals

- **This is not a ticketing system.** Gap notes are lightweight notes about harness evolution,
  not a replacement for Jira/Linear tracking of features or bugs.
- **Not every bug becomes a gap note.** Most production bugs get a fix and a regression test —
  that is already harness reinforcement via the existing test sensor. Open a gap note only when
  the class of bug reveals a missing control, not an isolated slip.
- **Not every new control needs its own spec.** A new golangci-lint rule or a one-line addition
  to a `.claude/rules/` file is a PR, not a spec. Use `/spec` only when the control spans
  multiple files or touches infrastructure.

## References

- Fowler, ["Harness Engineering for Coding Agents"](https://martinfowler.com/articles/harness-engineering.html)
  — in particular the "Self-Steering Loop" subsection.
- [docs/harness.md](../harness.md) — the current inventory this loop keeps up to date.
- [CLAUDE.md § Execution Directives](../../CLAUDE.md) — the mandatory
  Plan → Implement → Review → Test → Validate cycle that every harness addition follows.
