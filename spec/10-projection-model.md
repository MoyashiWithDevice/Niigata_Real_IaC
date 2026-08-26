# Projection Model

> **Implementation status:** The `src/projection/` implementation has been
> removed. Projection functionality is planned to be integrated into the Query
> model in a future release rather than maintained as a standalone package.
> This document describes the design intent; refer to the Query model
> (`spec/07-query-model.md`) for the currently supported behavior.

## Overview

A Projection transforms one Graph into another Graph.

A Projection never modifies the source Graph.

The output Graph represents the same infrastructure knowledge from a different perspective.

Projections are composable.

---

## Purpose

Projections simplify, enrich, or reorganize a Graph without changing its meaning.

Typical uses include:

- Physical topology
- Logical topology
- Network topology
- Rack layout
- Service dependency
- Inventory
- Documentation
- Validation preparation

---

## Graph Transformation

A Projection accepts exactly one input Graph.

A Projection produces exactly one output Graph.

Both Graphs are valid Object Models.

The output Graph MAY contain fewer Objects than the input.

The output Graph MAY introduce derived Objects.

Derived Objects MUST NOT change the semantics of the source Graph.

---

## Operations

A Projection MAY perform one or more of the following operations.

### Selection

Choose a subset of Objects.

### Filtering

Remove Objects that match specific conditions.

### Traversal

Follow Relations to discover additional Objects.

### Aggregation

Represent multiple Objects as a single derived Object.

### Expansion

Replace an abstract Object with more detailed Objects.

### Annotation

Attach computed metadata.

### Grouping

Organize Objects into logical collections.

---

## Derived Objects

A Projection MAY create derived Objects.

Derived Objects exist only within the output Graph.

Derived Objects MUST contain provenance information.

Derived Objects MUST NOT be written back to the canonical Graph.

Examples include:

- Rack summary
- VLAN summary
- Cluster node group
- Service dependency group

---

## Provenance

Every derived Object SHOULD reference the source Objects that produced it.

Provenance enables traceability.

Implementations MAY expose provenance information to users.

---

## Identity

Objects copied from the source Graph preserve their identifiers.

Derived Objects receive new identifiers.

Identifier collisions are not permitted.

---

## Composition

Projections MAY be chained.

Example:

Graph

↓

Physical Projection

↓

Network Projection

↓

Documentation Projection

↓

Renderer

Each Projection operates only on its input Graph.

---

## Determinism

A Projection MUST be deterministic.

Applying the same Projection to the same Graph MUST always produce the same output Graph.

---

## Side Effects

Projections are pure operations.

A Projection MUST NOT modify:

- the source Graph
- external systems
- persistent storage

---

## Extensibility

Implementations MAY define additional Projection operations.

Additional operations MUST preserve semantic equivalence with the source Graph.
