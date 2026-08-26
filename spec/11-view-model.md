# View Model

## Overview

A View defines how a projected Graph is presented to a consumer.

A View does not modify the Graph.

A View does not define rendering.

A View defines perspective.

Views are ephemeral.

They are generated from the Graph.

Views are never stored as the canonical model.

---

## Purpose

A View provides a meaningful interpretation of infrastructure knowledge.

Different Views expose different aspects of the same Graph.

Examples include:

- Physical Infrastructure
- Logical Infrastructure
- Network Topology
- Storage Topology
- Virtualization
- Security Zones
- Service Dependencies
- Inventory
- Rack Layout
- Application Dependency
- Documentation

---

## View Pipeline

Every View consists of three conceptual stages.

Graph

↓

Query

↓

Projection

↓

Rendering

The Query determines what is included.

The Projection determines how Objects are interpreted.

The View determines how that Graph should be interpreted.

Rendering determines presentation.

---

## Relationship to Projection

A View consumes the output of a Projection.

The Projection determines the available Graph.

The View determines how that Graph should be interpreted.

The View never changes the Graph.

---

## View Definition

A View MAY define:

- title
- description
- intended audience
- grouping rules
- visibility rules
- annotation rules
- layout hints

View definitions contain no rendering information.

---

## Visibility

A View MAY hide Objects.

Hidden Objects remain part of the underlying Graph.

Visibility affects presentation only.

---

## Relation Lifting

When a View hides Objects, Relations anchored on hidden Objects MUST NOT be
silently lost: their connectivity knowledge is lifted onto visible Objects.

A Relation participant that is not visible is mapped to visible anchor
Objects:

1. If the participant itself is visible, it is its own anchor.
2. Otherwise every visible Object inside its ownership subtree becomes an
   anchor (a hidden node maps to the applications it hosts; a hidden cluster
   maps to all applications beneath it).
3. If neither applies, the nearest visible ancestor is used (e.g. a hidden
   port maps to the application owning it).

The two sides of a Relation are then collapsed so that one Relation yields at
most one derived edge:

- A single anchor is used directly.
- Multiple anchors sharing a common hidden ancestor whose visible subtree is
  exactly the anchor set collapse onto a structural group keyed by that
  ancestor.
- Otherwise the first anchor (ordered by ID) acts as a deterministic
  representative.

Derived edges:

- inherit the type and direction of the source Relation,
- are marked as derived and record the source Relation identifiers,
- never form self loops,
- are suppressed when a direct Relation between the same endpoints is already
  visible.

Lifting is deterministic and never modifies the Graph. It produces no output
when all Objects are visible.

---

## Grouping

A View MAY organize Objects into logical groups.

Examples include:

- Rack
- VLAN
- Cluster
- Availability Zone
- Application Stack

Grouping does not alter ownership.

Grouping does not alter Relations.

---

## Annotation

Views MAY attach annotations.

Examples include:

- Calculated utilization
- Interface speed
- Host counts
- Warning indicators

Annotations are ephemeral.

Annotations never modify the Graph.

---

## Audience

Views MAY specify their intended audience.

Examples include:

- Network Engineers
- Infrastructure Engineers
- Security Engineers
- Developers
- Operators
- Management

Audience information is descriptive only.

---

## Rendering Independence

Views do not define rendering.

The same View may be rendered as:

- SVG
- PNG
- PDF
- Mermaid
- D2
- Graphviz DOT
- Markdown
- HTML
- JSON

---

## Composition

Multiple Views MAY consume the same Projection.

A View MAY also consume different Projection outputs.

Views are reusable.

---

## Persistence

Views are not canonical data.

Views SHOULD be regenerated whenever possible.

Persisting Views is implementation-specific.

---

## Extensibility

Implementations MAY define custom View types.

Custom Views MUST preserve the semantics of the underlying Graph.
