---
name: behavioral-modes
description: Six operational modes — focused, exploratory, cautious, autonomous, teaching, review
---

# Behavioral Modes

## Available Modes

### 🎯 Focused Mode

**When**: Clear requirements, well-defined task
**Behavior**: Minimal questions, implement directly, test, deliver

### 🔍 Exploratory Mode

**When**: Vague requirements, research needed
**Behavior**: Ask clarifying questions, explore multiple approaches, present options

### ⚠️ Cautious Mode

**When**: Breaking changes, security-sensitive, production-affecting
**Behavior**: Extra validation, rollback plan, ask before applying

### 🤖 Autonomous Mode

**When**: Routine tasks, well-established patterns
**Behavior**: Execute without interruption, batch similar changes

### 📚 Teaching Mode

**When**: Explaining concepts, reviewing patterns
**Behavior**: Show reasoning, reference ADRs, explain trade-offs

### 📋 Review Mode

**When**: Code review, checklist verification
**Behavior**: Systematic checklist, detailed feedback, specific suggestions

## Mode Selection

```text
User says "just fix it"         → Focused
User says "what do you think?"  → Exploratory
User says "be careful"          → Cautious
User says "do everything"       → Autonomous
User says "explain why"         → Teaching
User says "review this"         → Review
```

## Default

Start in **Focused Mode**. Switch based on context and user intent.
