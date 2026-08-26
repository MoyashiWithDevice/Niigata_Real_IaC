# YAML Grammar

## Overview

YAML is the first standardized serialization format for the Object Model.

This specification defines the canonical YAML representation.

Other authoring formats MAY exist.

---

## Document Structure

A YAML document consists of a single Graph.

The document root MUST contain:

objects:

Additional sections MAY be defined by future specifications.

---

## Objects

Every Object MUST appear exactly once.

Objects are distinguished by their object type.

Entity

Relation

---

## Entity

An Entity object MUST define:

id

kind

name

An Entity object MAY contain an `attributes:` mapping.

Attributes hold optional common properties shared across all kinds (e.g. `owner`, `tags`).

An Entity object MAY contain a `spec:` mapping.

Spec holds kind-specific properties defined by the Entity kind.

---

## Relation

A Relation object MUST define:

id

type

participants

A Relation object MAY contain an `attributes:` mapping.

Attributes hold optional common properties shared across all relation types.

A Relation object MAY contain a `spec:` mapping.

Spec holds relation-type-specific properties defined by the relation type.

---

## References

References use canonical identifiers.

A parser MUST resolve every reference.

Unknown references are validation errors.

---

## Ownership

Ownership is declared in the child Entity's `owner` property.

The root Entity does not specify an owner.

Hierarchical nesting MUST NOT change semantics.

Nested syntax MAY exist in authoring formats.

The canonical format always represents ownership explicitly.

---

## Unknown Properties

Unknown properties MUST be preserved.

The core specification ignores unknown properties.

Extensions MAY interpret them.

---

## Ordering

Object order has no semantic meaning.

Implementations SHOULD preserve ordering when possible.

---

## Comments

Comments are not part of the Object Model.

Mappings SHOULD preserve comments whenever possible.

---

## Round-trip

Parsing and serializing a canonical YAML document SHOULD preserve semantic equivalence.

Formatting differences are acceptable.

Semantic differences are not.
