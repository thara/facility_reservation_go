---
layout: default
title: Architecture Decision Records
nav_order: 4
has_children: true
permalink: /adr
---

# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records for the Facility Reservation System.

## ADR Index

| Number | Title | Status |
|--------|-------|--------|
| [0001](0001-use-typespec-over-openapi-yaml.md) | Use TypeSpec over OpenAPI YAML | Accepted |
| [0002](0002-use-sqlc-over-orm.md) | Use sqlc over ORM | Accepted |
| [0003](0003-avoid-database-triggers.md) | Avoid Database Triggers | Accepted |

## What is an ADR?

Architecture Decision Records (ADRs) document important architectural decisions made during the development of this system. Each ADR captures:

- The context and forces that led to the decision
- The decision itself
- The consequences (positive and negative) of the decision

## ADR Format

We use the format popularized by Michael Nygard:

```markdown
# [Number]. [Title]

Date: [YYYY-MM-DD]

## Status

[Proposed | Accepted | Deprecated | Superseded by [link]]

## Context

[Describe the forces at play, including technological, political, social, and project factors]

## Decision

[State the architecture decision and explain why this decision was made]

## Consequences

[Describe the positive and negative consequences of this decision]
```

## Creating a New ADR

1. Copy the template format above
2. Use the next sequential number
3. Write a clear, descriptive title
4. Fill in all sections thoroughly
5. Start with status "Proposed"
6. Change to "Accepted" once the team agrees

## Principles

- **One decision per ADR**: Each ADR should document exactly one architectural decision
- **Immutable**: Once accepted, ADRs should not be changed (create a new superseding ADR instead)
- **Context matters**: Include enough background so future readers understand why the decision was made
- **Honest about tradeoffs**: Document both positive and negative consequences