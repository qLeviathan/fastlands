# Pizza Dispatch System

A lattice-routed agentic workflow built on exact arithmetic.

## Quick Start

```bash
# From any terminal, including mobile
go run ./cmd/pizza next       # what to do right now
go run ./cmd/pizza express    # 48-hour express plan
go run ./cmd/pizza plan       # full dispatch with all agents
go run ./cmd/pizza repl       # interactive shell
go run ./cmd/pizza legend     # decoder ring
```

## Commands

| Command | What it does |
|---------|-------------|
| `pizza plan` | Full dispatch plan — checker validates, creative finds cascades, logger tracks |
| `pizza next` | Single best next slice based on lattice coupling |
| `pizza express` | 48-hour express order only |
| `pizza start <id>` | Mark a slice as in-progress |
| `pizza done <id>` | Mark done, holographic memory learns, routes you to the next cascade |
| `pizza oven <id>` | Mark submitted / waiting |
| `pizza burnt <id> [reason]` | Mark failed |
| `pizza status` | Grid overview with all slices |
| `pizza holo` | Holographic memory visualization |
| `pizza log` | Full build log with learnings |
| `pizza legend` | The decoder ring — what every topping means |
| `pizza repl` | Interactive dispatch shell |

## How Routing Works

Every task sits at position **(n, m)** on the lattice.

- **n** (Dough) = preparation style / skill axis
- **m** (Sauce) = kitchen station / platform axis
- **Temperature** = n + m (total effort shell)
- **Flavor** = F(n) · F(m) (basis coupling strength)

When you complete a slice, the router finds the next one that **shares an axis**:

| Coupling | Condition | Effect |
|----------|-----------|--------|
| **Cascade** | Shared axis, gap = 1 | Output feeds directly into next slice |
| **Resonance** | Shared axis, gap = 2 | Harmonic reinforcement |
| **Coupled** | Shared axis, gap > 2 | Standard skill/platform transfer |
| **Parallel** | Same diagonal, no shared axis | Independent, no interaction |

## Three Agents

1. **Checker** — Validates readiness, finds blockers, scores 0-100
2. **Creative** — Spots cascade combos, then grounds them into actionable steps
3. **Logger** — Records everything, generates learnings

## Architecture

```
lattice/    (n,m) grid, coupling rules, menu presets
dispatch/   three-agent system + orchestrator
holo/       holographic memory with interference-pattern recall
cmd/pizza/  CLI entry point
```

Built on top of the caslo inference engine (`ar/`, `basis/`, `memory/`, `swarm/`).

## Phone Workflow

The system is designed to run from `claude code` on mobile:

```
pizza next          # see what's next
pizza start tc-mar-001   # begin it
pizza oven tc-mar-001    # submitted, waiting
pizza done tc-mar-001    # completed → auto-routes next
```

Each `done` feeds holographic memory, which strengthens future routing decisions through interference-pattern recall along shared axes.
