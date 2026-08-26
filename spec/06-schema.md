# Schema

## Overview

A Schema defines the structure of a valid infrastructure model.

The Schema is independent of any serialization format.

The Schema defines what may exist.

Validation determines whether a model satisfies the Schema.

---

## Purpose

A Schema describes:

- Object kinds
- Relation types
- Property definitions
- Cardinality
- Constraints

Schemas are declarative.

Schemas never contain executable logic.

---

## Core Schema

The core specification defines the Core Schema.

Every compliant implementation MUST support the Core Schema.

Extensions MAY extend the Core Schema.

Extensions MUST NOT redefine existing semantics.

---

## Entity Definitions

Every Entity kind is defined by a Schema.

A definition includes:

- required properties
- optional properties
- property types
- default values
- documentation

---

## Relation Definitions

Every Relation type is defined by a Schema.

Definitions include:

- directionality
- participant constraints
- documentation

Future specifications MAY define additional semantics.

---

## Property Definitions

Properties have:

- name
- type
- optionality
- default value
- description

Property definitions are independent of serialization.

---

## Defaults

Schemas MAY define default values.

Missing properties are interpreted as unknown unless a default value exists.

Default values never imply that a property was explicitly specified.

---

## Profiles

A Schema MAY define profiles.

Examples:

- Homelab
- Enterprise
- Proxmox
- Kubernetes

Profiles specialize the Core Schema.

---

## Compatibility

Schema evolution SHOULD preserve compatibility whenever possible.

Breaking changes require a new major specification version.
