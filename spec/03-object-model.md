# Object Model

## Overview

The infrastructure model is represented as a graph.

The graph consists of Objects.

Objects are either Entities or Relations.

The graph is the canonical representation of the infrastructure model.

All serialization formats MUST represent the same graph.

---

## Objects

Objects are identifiable elements within a graph.

Every Object has:

- identity
- extensions
- lifecycle

Objects are divided into two categories.

- Entity
- Relation

No other Object categories are defined by the core specification.

---

## Graph

A Graph is a collection of Objects.

The Graph represents the complete infrastructure model.

Graphs are independent of serialization format.

Graphs are independent of rendering.

Graphs are independent of providers.

---

## Ownership Tree

Ownership is declared in the child Entity's `owner` property.

The root Entity does not specify an owner.

Ownership forms exactly one tree.

Every Entity except the root MUST have exactly one owner specified.

The ownership tree exists for navigation only.

Ownership MUST NOT imply execution, communication, dependency, or connectivity.

---

## Semantic Graph

All peer-to-peer Relations form the semantic graph.

The semantic graph expresses:

- connectivity
- dependency
- execution
- membership

The semantic graph may contain cycles.

Cycles are valid unless prohibited by validation rules.

---

## Paths

Every Entity has a canonical path.

The canonical path is constructed from ownership.

Example

/region01/rack01/pve01/eno1

Canonical paths are stable as long as ownership remains unchanged.

---

## References

Objects may reference other Objects.

References MUST use canonical identifiers.

Serialization formats MAY provide shorthand references.

The object model always resolves references to canonical Objects.

---

## Graph Integrity

The graph MUST satisfy the following conditions.

- Every Object has an identifier.
- Every Entity except the root has exactly one owner specified.
- The `owner` property references an existing Entity.
- Relations reference existing Objects.
- Object identifiers are unique.
- Ownership forms a tree.

Everything else is determined by validation profiles.

---

## Incomplete Graphs

Graphs may be incomplete.

Unknown Objects may be omitted.

Unknown Relations may be omitted.

The graph remains valid.

Validation determines completeness for specific use cases.

---

## Graph Evolution

Graphs are expected to evolve over time.

Objects may be added.

Objects may be removed.

Relations may change.

Stable identifiers preserve continuity across revisions.

---

## Serialization Independence

The object model does not define YAML.

The object model does not define JSON.

The object model defines only concepts.

Serialization formats represent those concepts.
