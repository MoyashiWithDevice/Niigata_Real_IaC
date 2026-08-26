# Relation

## Overview

A Relation represents a peer-to-peer semantic connection between two or more Entities.

Relations describe how Entities are related in a non-hierarchical manner.

Ownership hierarchy is declared in the child Entity's `owner` property.

Relations form the graph structure of the infrastructure model.

Every Relation is an independent object.

---

## Identity

Every Relation MUST have a stable identifier.

Identifiers are unique within the model.

Relations MAY be referenced by other objects.

Changing a Relation identifier creates a different Relation.

---

## Relation Types

Every Relation MUST define a relation type.

The relation type defines the meaning of the relationship.

The core specification defines several standard relation types.

Implementations MAY introduce additional relation types.

---

## Standard Relation Types

The following relation types are defined by the core specification.

### connects

Represents a connection.

This relation is symmetric.

Typical examples include:

- Interface ↔ Cable
- Cable ↔ Interface

---

### hosts

Represents an execution or hosting relationship.

Examples include:

- Server hosts VM
- VM hosts Application

---

### depends_on

Represents dependency.

Dependencies are directional.

If A depends_on B,

A requires B,

but B does not require A.

---

### belongs_to

Represents logical membership.

Examples include:

- Interface belongs_to Network
- VM belongs_to Cluster

Membership does not imply ownership.

---

## Direction

Every Relation defines its directionality.

Relations are one of the following:

- Directed
- Symmetric

Directed relations have a source and a target.

Symmetric relations have two or more equal participants.

---

## Participants

A Relation connects two or more participants.

Participants are usually Entities.

Future specifications MAY allow Relations to participate in Relations.

---

## Common Properties

Every Relation shares the following properties.

Required

- id
- type

Optional

- description
- status
- tags
- labels
- extensions

---

## Description

Descriptions provide human-readable documentation.

Descriptions MAY contain Markdown.

Descriptions MUST NOT affect semantics.

---

## Status

Status represents the lifecycle state of a Relation.

The core specification recommends:

- planned
- active
- maintenance
- deprecated
- offline
- standby

Additional statuses MAY be introduced by extensions.

---

## Extensions

Extensions are ignored by the core specification.

Extensions MAY interpret extensions.

---

## Relation Semantics

The meaning of a Relation is defined exclusively by its relation type.

Entity kinds MUST NOT redefine the semantics of a Relation.

For example,

depends_on always means dependency,

regardless of the participating Entity kinds.

---

## User-defined Relation Types

Implementations MAY define additional relation types.

User-defined relation types MUST NOT redefine the semantics of standard relation types.

Examples include:

- replicates_to
- backs_up
- monitors
- managed_by
- mounted_on
- applies_to
- listens_on

---

## Validation

Validation rules MAY restrict

- allowed relation types
- participant kinds
- relation cardinality
- cycles
- required relations

The object model itself imposes no such restrictions unless explicitly defined by the core specification.

---

## Equality

Relations are uniquely identified by their identifier.

Changing participants modifies an existing Relation.

Changing the identifier creates a different Relation.
