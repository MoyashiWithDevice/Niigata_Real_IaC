# Rendering

## Overview

Rendering transforms a View into a presentation.

Rendering is the final stage of the infrastructure modeling pipeline.

A Renderer consumes a View.

A Renderer produces an output artifact.

Rendering never modifies the underlying Graph.

Rendering never changes the semantics of the View.

---

## Purpose

Rendering communicates infrastructure knowledge to humans or external systems.

Examples include:

- SVG
- PNG
- PDF
- HTML
- Markdown
- Mermaid
- Graphviz DOT
- D2
- JSON
- CSV

Future renderers may target additional formats.

---

## Rendering Pipeline

Rendering operates on a View.

The conceptual pipeline is:

Graph

↓

Projection

↓

View

↓

Renderer

↓

Artifact

The Artifact is the only output of a Renderer.

---

## Rendering Independence

The same View MUST be renderable by multiple Renderers.

A Renderer MUST NOT require changes to the View definition.

Presentation is independent of infrastructure knowledge.

---

## Artifact

Artifacts are generated outputs.

Artifacts are not canonical.

Artifacts MAY be discarded and regenerated at any time.

Examples include:

- Diagram
- Documentation
- Report
- Inventory
- Configuration Template

---

## Layout

Layout determines the spatial arrangement of rendered elements.

Layout is renderer-specific.

The core specification does not define layout algorithms.

Implementations MAY provide:

- Hierarchical layout
- Force-directed layout
- Orthogonal layout
- Rack layout
- Timeline layout

---

## Themes

Renderers MAY support Themes.

Themes define presentation characteristics such as:

- colors
- typography
- spacing
- icons
- line styles

Themes MUST NOT change semantics.

---

## Interactivity

Renderers MAY generate interactive artifacts.

Examples include:

- expandable groups
- hyperlinks
- tooltips
- filtering
- search
- animation

Interactivity is presentation only.

---

## Renderer Metadata

Renderers MAY define implementation-specific metadata.

Renderer metadata MUST NOT become part of the Object Model.

---

## Determinism

Rendering SHOULD be deterministic.

Rendering the same View with the same Renderer configuration SHOULD produce equivalent output.

Differences caused by timestamps, metadata, or implementation details MAY exist but SHOULD be minimized.

---

## Accessibility

Renderers SHOULD support accessible output whenever possible.

Examples include:

- alternative text
- keyboard navigation
- high contrast themes
- scalable text

Accessibility features do not change the View.

---

## Extensibility

Implementations MAY define additional Renderer capabilities.

Additional capabilities MUST preserve the semantics of the View.
