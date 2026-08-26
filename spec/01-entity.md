# Entity

## Overview

An Entity represents any identifiable object within an infrastructure model.

Everything that exists in infrastructure is represented as an Entity.

Examples include physical devices, virtual resources, network components, software, locations, security constructs, and abstract concepts.

The object model intentionally does not distinguish between physical and logical objects.

Both are represented uniformly as Entities.

---

## Identity

Every Entity MUST have a stable identifier.

Identifiers are unique within the scope of their Owner.

The fully-qualified path uniquely identifies an Entity within a model.

Example:

/region01/rack01/pve01/eno1

Identifiers SHOULD remain stable throughout the lifetime of an Entity.

Changing an identifier SHOULD be treated as replacing the Entity.

---

## Ownership

Every Entity MUST have exactly one Owner.

The child Entity declares its owner by specifying the parent Entity identifier in the `owner` property.

The root Entity does not specify an owner.

Ownership defines hierarchy only.

Ownership does not imply dependency, communication, execution, or containment in the real world.

Ownership relationships form a tree.

All other relationships MUST be represented as Relations.

---

## Entity Types

Every Entity MUST define a kind.

Kinds represent the role of an Entity within the object model.

Kinds are defined by the specification.

Implementations MAY introduce additional kinds through extensions.

Vendor-specific implementations MUST NOT introduce vendor names as kinds.

Example:

Correct

kind: server

platform: proxmox

Incorrect

kind: proxmox

---

## Common Properties

Every Entity shares the following common properties.

Required

- id
- kind
- name

Optional

- owner
- description
- status
- tags
- labels
- extensions

The `owner` property specifies the parent Entity identifier for ownership hierarchy.

The root Entity omits the `owner` property.

Additional properties MAY be defined by individual Entity kinds.

---

## Description

The description field contains human-readable documentation.

Descriptions MAY contain Markdown.

Descriptions MUST NOT influence the behavior of the model.

---

## Status

Status represents the lifecycle state of an Entity.

The core specification defines the following statuses.

- planned
- active
- maintenance
- deprecated
- offline
- standby

Implementations MAY introduce additional statuses.

Validation profiles MAY restrict allowed values.

---

## Tags

Tags are unordered labels used for grouping and filtering.

Tags have no predefined meaning.

---

## Labels

Labels are key-value pairs.

Labels are intended for machine-readable categorization.

Keys SHOULD be unique within an Entity.

---

## Extensions

Extensions contain implementation-specific information organized into two sub-keys:

- `attributes` - Key-value pairs for general metadata (labels, platform info, etc.)
- `spec` - Structured data specific to providers or renderers

The core specification MUST ignore extensions.

Providers, renderers, and external integrations MAY interpret extensions.

Example:

```yaml
extensions:
  attributes:
    platform: proxmox
    environment: production
  spec:
    proxmox:
      vmid: 100
      pool: production
```

---

## Vendor Neutrality

The core object model represents infrastructure concepts.

Vendor-specific concepts belong outside the core model.

Examples include:

Vendor

Dell
Cisco
Proxmox
VMware

Implementation

platform
provider
extensions

These values never replace the Entity kind.

---

## Extensibility

Additional Entity kinds MAY be introduced by extensions.

Extensions MUST preserve compatibility with the core object model.

Extensions MUST NOT redefine the semantics of existing Entity kinds.

---

## Incomplete Entities

An Entity does not need to contain all information.

Unknown values MAY be omitted.

Incomplete Entities remain valid Entities.

Validation determines whether an Entity satisfies a specific deployment profile.

---

## Entity Equality

Two Entities are considered different if their identifiers differ.

Changing properties does not create a new Entity.

Changing an identifier creates a different Entity.
