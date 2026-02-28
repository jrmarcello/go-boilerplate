---
name: context7
description: Using Context7 MCP to fetch up-to-date library documentation and code examples
---

# Context7

## Overview

Context7 is an MCP (Model Context Protocol) tool that fetches current, version-specific documentation and code examples for any library.

## Usage

### Step 1: Resolve Library ID

Use `mcp_upstash_conte_resolve-library-id` to find the correct library identifier.

### Step 2: Query Documentation

Use `mcp_upstash_conte_query-docs` with the resolved library ID to get documentation.

## When to Use

- Before implementing with an unfamiliar library
- When documentation might have changed
- To get version-specific API references
- For library-specific code patterns

## Common Libraries

| Library | Usage in Project |
| --- | --- |
| gin-gonic/gin | HTTP framework |
| jmoiron/sqlx | Database queries |
| redis/go-redis | Cache client |
| oklog/ulid | ID generation |
| pressly/goose | DB migrations |
| swaggo/swag | Swagger docs |
| stretchr/testify | Test assertions |
| otel | OpenTelemetry SDK |

## Integration with Skills

Load this skill when you need to:

- Verify API signatures before coding
- Check for breaking changes in new versions
- Find best practices for a specific library
