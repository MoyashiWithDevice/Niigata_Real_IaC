# Query Model

## Overview

The Query Model defines how Objects are selected from a Graph.

Queries never modify the Graph.

Queries produce Views of the Graph.

---

## Purpose

Queries provide a consistent mechanism for:

- Validation
- Rendering
- Export
- Analysis
- Search

The same Graph may be queried in many different ways.

---

## Object Selection

Queries operate on Objects.

Selections MAY include:

- Individual Objects
- Entity kinds
- Relation types
- Ownership paths
- Metadata
- Tags
- Labels

---

## Traversal

Queries MAY traverse:

- Ownership
- Relations
- Incoming Relations
- Outgoing Relations

Traversal direction is explicit.

---

## Projection

Queries MAY return:

- Objects
- Properties
- Relations
- Paths

The underlying Graph is never modified.

---

## Filtering

Queries MAY filter by:

- kind
- type
- status
- tags
- labels
- metadata

Future specifications MAY define additional filters.

---

## Composition

Queries MAY be combined.

The output of one query MAY become the input of another.

---

## Determinism

Executing the same query on the same Graph MUST always produce the same result.

---

## Side Effects

Queries never change the Graph.

Queries are pure operations.

---

## Serialization

The core specification does not define a query language.

Implementations MAY provide:

- CLI
- API
- GraphQL
- DSL

All query mechanisms operate on the same Query Model.
