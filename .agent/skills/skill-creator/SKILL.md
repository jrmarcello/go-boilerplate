---
name: skill-creator
description: Guide for creating new skills — structure, metadata, formatting, and validation
---

# Skill Creator

## Skill Structure

```text
.agent/skills/<skill-name>/
├── SKILL.md              # Main definition (required)
├── references/           # Supporting docs (optional)
│   └── additional.md
└── scripts/              # Automation scripts (optional)
    └── check.sh
```

## SKILL.md Template

```markdown
---
name: skill-name
description: One-line description of what this skill enables
---

# Skill Name

## Purpose
[What this skill does and when to load it]

## Core Concepts
[Key knowledge organized by topic]

## Examples
[Code examples demonstrating usage]

## Rules
[Do's and don'ts]
```

## Naming Conventions

- Directory: `kebab-case` (e.g., `go-patterns`)
- SKILL.md: Always uppercase
- References: `kebab-case.md`
- Scripts: `kebab-case.sh` or `snake_case.py`

## Quality Criteria

- [ ] Has YAML front matter with `name` and `description`
- [ ] Includes at least one code example
- [ ] References are project-specific (not generic)
- [ ] No placeholder content
- [ ] Less than 300 lines (split into references if larger)
