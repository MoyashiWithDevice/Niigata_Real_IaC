# Extension Model

## Overview

The Extension Model defines how the Infrastructure Model may be extended without modifying the Core Specification.

Extensions allow implementations to introduce new capabilities while preserving interoperability.

The Core Specification remains intentionally small and stable.

---

## Purpose

Extensions enable domain-specific functionality.

Typical extensions include:

- Entity kinds
- Relation types
- Schemas
- Validation rules
- Projections
- Views
- Layout engines
- Renderers
- Providers

Extensions are optional.

Implementations may support any combination of extensions.

---

## Core Compatibility

An Extension MUST NOT redefine the semantics of the Core Specification.

Core concepts retain their original meaning.

Extensions may introduce additional concepts.

They may not change existing concepts.

---

## Namespaces

Every Extension SHOULD define a namespace.

Namespaces prevent naming conflicts.

Example:

network.switch

proxmox.vm

kubernetes.pod

aws.vpc

---

## Discovery

Extensions SHOULD expose machine-readable metadata.

Typical metadata includes:

- identifier
- version
- author
- description
- supported specification version
- dependencies

---

## Dependencies

Extensions MAY depend on other Extensions.

Circular dependencies SHOULD be avoided.

Dependency resolution is implementation-specific.

---

## Compatibility

Extensions SHOULD declare compatibility with:

- Core Specification version
- Schema version
- Language version

Implementations MAY reject incompatible Extensions.

---

## Isolation

Extensions SHOULD operate independently whenever possible.

One Extension SHOULD NOT require knowledge of another Extension unless explicitly declared.

---

## Unknown Extensions

Implementations SHOULD ignore unsupported Extensions whenever possible.

Ignoring an Extension MUST NOT corrupt the underlying Graph.

---

## Security

Extensions may execute implementation-specific logic.

The Core Specification does not define execution semantics.

Execution environments are implementation-specific.

---

## Versioning

Extensions SHOULD follow semantic versioning.

Breaking changes SHOULD require a major version increment.

---

## Stability

The Graph MUST remain valid regardless of which Extensions are installed.

Extensions enhance the ecosystem.

They never redefine the foundation.
