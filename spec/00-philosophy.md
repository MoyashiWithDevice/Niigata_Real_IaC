# Philosophy

## Mission

Infrastructure is knowledge.

This project defines a portable, human-readable, and extensible model for describing infrastructure.

The goal of this project is not to draw diagrams.

The goal is to describe infrastructure once and generate every possible representation from that single source of truth.

---

## Source of Truth

The infrastructure model is the only source of truth.

Diagrams, documentation, reports, inventories, validation results, and exports are all generated artifacts.

They must never become the source of truth.

---

## Object Model First

The object model is the core of the project.

YAML, JSON, APIs, and GUI editors are simply different representations of the same model.

The model must never be designed around a specific serialization format.

---

## Everything is an Entity.

Every object that exists in infrastructure is represented as an Entity.

Examples include:

- Region
- Rack
- Server
- Interface
- Cable
- Network
- VM
- Container
- Application
- ACL
- Open Port

---

## Everything is a Relation.

Peer-to-peer relationships between entities are explicit.

Relations have their own identity and metadata.

Examples include:

- connects
- hosts
- depends_on
- applies_to
- listens_on

---

## Ownership

Every Entity has exactly one Owner.

Ownership is declared in the child Entity's manifest by specifying the parent identifier.

Ownership forms a tree.

Peer-to-peer relationships are represented as Relations.

This separation allows hierarchical navigation while preserving graph semantics.

---

## Human Readable

The primary author of infrastructure data is a human.

The model must remain readable in plain text.

Generated data must never replace hand-written data.

---

## Extensibility

Everything except the core object model is extensible.

Examples include:

- Entity kinds
- Relation types
- Views
- Validation rules
- Providers
- Renderers

---

## Vendor Neutrality

The core object model does not understand vendor-specific concepts.

Vendor-specific information belongs to Providers.

The model represents infrastructure concepts, not implementations.

---

## View Independence

The model never stores presentation information.

Coordinates, colors, layout hints, and rendering details belong to Views.

The model only represents knowledge.

---

## Progressive Validation

The model should allow incomplete designs.

Validation is responsible for determining whether a model satisfies a particular profile or deployment target.

Incomplete infrastructure is still valid infrastructure.

---

## Long-term Stability

The object model is expected to remain stable for many years.

Implementations may evolve.

Views may evolve.

Providers may evolve.

The model should change only when absolutely necessary.
