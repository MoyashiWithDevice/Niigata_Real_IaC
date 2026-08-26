# Validation

## Overview

Validation evaluates a Graph against one or more rules.

Validation never modifies the Graph.

Validation reports findings.

---

## Purpose

Validation determines whether a Graph satisfies a particular Schema or Profile.

Validation is independent from serialization.

Validation is independent from rendering.

---

## Validation Rules

Validation consists of Rules.

Rules evaluate Objects and Relations.

Rules never modify the Graph.

---

## Profiles

Validation Rules are grouped into Profiles.

Examples include:

- Core
- Homelab
- Enterprise
- Proxmox
- Kubernetes

A Graph may satisfy multiple Profiles simultaneously.

---

## Levels

Every validation finding has a severity.

The core specification defines:

- Info
- Warning
- Error

Implementations MAY define additional levels.

---

## Results

Validation produces Findings.

Every Finding includes:

- Rule
- Severity
- Message
- Related Objects

Implementations MAY include additional information.

---

## Determinism

Validation is deterministic.

Running the same validation against the same Graph MUST always produce identical findings.

---

## Extensibility

Implementations MAY introduce custom Validation Rules.

Custom Rules MUST NOT redefine the semantics of the Core Schema.

---

## Relationship to Query

Validation is a specialized Query.

Validation Rules operate by querying the Graph and evaluating the result.
