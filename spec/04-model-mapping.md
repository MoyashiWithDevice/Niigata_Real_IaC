# Model Mapping

## Overview

The Object Model is independent of any serialization format.

A Model Mapping defines how Objects are represented in external formats.

The mapping preserves semantics.

It does not change the meaning of the model.

---

## Purpose

A Model Mapping exists to transform the canonical Object Model into another representation.

Examples include:

- YAML
- JSON
- REST
- GraphQL
- Binary formats

Every mapping MUST preserve the same graph.

---

## Canonical Model

The Object Model is canonical.

Mappings are projections of the Object Model.

No mapping may introduce concepts that do not exist in the Object Model.

---

## Information Preservation

A valid mapping MUST preserve:

- Object identity
- Object kind
- Relation semantics
- Ownership
- References
- Metadata

Mappings MUST NOT lose semantic information.

---

## Representation Independence

Mappings MAY represent the same information differently.

For example,

A Relation MAY appear

- inline
- as a reference
- as a separate collection

All representations are equivalent if they produce the same Object Model.

---

## Canonical References

Mappings MAY provide shorthand syntax.

Internally,

all references are resolved to canonical Object identifiers.

Applications MUST operate on canonical references.

---

## Ordering

The Object Model has no inherent ordering.

Mappings MAY preserve ordering for readability.

Ordering MUST NOT change semantics.

---

## Optional Information

Mappings MAY omit properties whose values are unknown.

Omitted information represents an unknown value.

It does not imply a default value.

---

## Extensions

Mappings MAY define syntax extensions.

Extensions MUST NOT alter the semantics of the Object Model.

Unknown extensions SHOULD be ignored whenever possible.

---

## Round-trip

A Mapping SHOULD support round-trip conversion.

Object Model

↓

Serialization

↓

Object Model

The resulting Object Model SHOULD be semantically identical.
