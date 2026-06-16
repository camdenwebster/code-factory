---
module: code-factory
date: 2026-06-16
problem_type: tooling_decision
component: compound-cli
severity: high
tags: [go, swift, cli, cross-platform, state-machine, distribution]
applies_when: Choosing the implementation language for the `compound` orchestration CLI, or any subprocess-coupled developer tool that drives Swift/Apple builds but does not link Swift libraries in-process.
---

# CLI language: Go over Swift for the `compound` orchestrator

## Decision

The `compound` CLI is implemented in **Go**, not Swift.

## Context

`compound` is the executable form of the CompoundEngine state machine. Its job is:
a pure `(state, event) → state` transition, a per-phase tool-policy check, a
verification gate that **shells out** to `swift test` / `xcodebuild`, a
file-backed knowledge store, JSON checkpointing, and a headless driver that
**shells out** to `claude -p` / `codex exec`.

The tempting premise — "it builds Swift apps, so write it in Swift" — does not
hold here: every Swift/Apple touchpoint is a **subprocess boundary**, not a
library linkage. There is no in-process dependency on SwiftPM, Xcode, or
SourceKit. Go spawns and parses those subprocesses as well as Swift does.

## Options considered

- **Swift** — best fit for the *core*: `PhaseEvent` carries associated values
  and `Transition` is an exhaustive `(phase, event)` switch, which is the exact
  mechanism behind the "can't skip a stage" guarantee. `Codable` makes the JSON
  artifacts effortless. Opens the door to in-process `swift-syntax` later.
  Costs: painful cross-compilation, a Swift runtime/toolchain on Linux CI, and —
  proven in practice — it would not build in our Linux container at all.
- **Go** — best fit for the *shell*: one static binary, trivial
  `GOOS/GOARCH` cross-compile, fast predictable startup (the firewall runs on
  every tool call), and a mature CLI/exec/json stdlib. Cost: no sum types, so
  the transition's compile-time exhaustiveness is weaker.
- **Rust** — theoretically ideal (sum types *and* a static binary, matches
  Codex's own language) but more than a solo, pragmatic maintainer wants to own.

## Deciding factors

1. **Maintainer is solo / pragmatic** → time-to-ship and frictionless
   distribution dominate language purity.
2. **In-process Swift code analysis is "maybe later", not now** → does not
   justify prepaying Swift's cross-platform tax. If it ever lands, it ships as a
   *separate* small `swift-syntax` helper that `compound` invokes as a
   subprocess — the same shell-out pattern already in use.
3. **The coupling is subprocess-only** → Swift's in-process advantages do not
   apply to this tool.

## Resolution

Implement in Go. Recover the one real loss — compile-time switch
exhaustiveness — with (a) the `exhaustive` linter over const-enum switches in
CI, and (b) the invariant test suite, which proves the determinism guarantees
at runtime regardless of language. Model `PhaseEvent` as a `Kind` const-enum +
payload struct; the Swift `default` arm that refuses illegal transitions ports
verbatim.

## Consequences

- **Carried over untouched** (language-agnostic): `spec/cli.md`,
  `spec/system-spec.html`, the example hook scripts, and both harness configs —
  they speak argv + JSON + exit codes and never knew the binary was Swift.
- **Rewritten**: only `CompoundEngine.swift` (~600 lines), translated directly.
- **Gained**: `crypto/sha256` (stdlib) replaces the CryptoKit→swift-crypto swap;
  the binary builds and tests in the Linux dev container and in CI.
- **Reversible**: if the team later standardizes on Swift, the pure core is the
  only non-trivial port, and the specs/hooks are unaffected.
