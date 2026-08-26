# Rendering

## Overview

Rendering transforms a View into a presentation artifact.

Rendering is the final stage of the infrastructure modeling pipeline.

A Renderer consumes a View.

A Renderer produces an output artifact.

Rendering never modifies the underlying Graph or View.

Every compliant implementation MUST support at least one Renderer.

---

## Renderer Structure

Every Renderer is defined with:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | yes | Unique renderer identifier |
| name | string | yes | Human-readable name |
| description | string | no | Renderer description |
| format | enum | yes | Output format |
| capabilities | list[string] | no | Renderer capabilities |
| options | map | no | Format-specific options |

---

## Output Formats

The core specification defines the following output formats.

> **Implementation status:** The current IACForge implementation ships the
> `svg`, `mermaid`, `markdown`, and `json` renderers. The `png`, `pdf`,
> `graphviz`, `d2`, `html`, and `csv` renderers described below are part of the
> specification but are **not yet implemented**; the CLI and MCP accept only
> the implemented formats. Renderers contributed through extensions can add
> additional formats at runtime.

### Diagram Formats

| Format | Description | Extension |
|--------|-------------|-----------|
| svg | Scalable Vector Graphics | .svg |
| png | Portable Network Graphics | .png |
| pdf | Portable Document Format | .pdf |
| mermaid | Mermaid diagram syntax | .mmd |
| graphviz | Graphviz DOT format | .dot |
| d2 | D2 diagram syntax | .d2 |

### Document Formats

| Format | Description | Extension |
|--------|-------------|-----------|
| html | HyperText Markup Language | .html |
| markdown | Markdown format | .md |
| json | JavaScript Object Notation | .json |
| csv | Comma-Separated Values | .csv |

### Data Formats

| Format | Description | Extension |
|--------|-------------|-----------|
| yaml | YAML format | .yaml |
| toml | TOML format | .toml |

---

## Artifact Structure

Every Artifact is defined with:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | yes | Unique artifact identifier |
| renderer_id | string | yes | Renderer identifier |
| view_id | string | yes | Source View identifier |
| format | enum | yes | Output format |
| content | any | yes | Artifact content |
| metadata | map | no | Artifact metadata |
| timestamp | string | yes | Creation timestamp |

### Artifact Metadata

| Field | Type | Description |
|-------|------|-------------|
| title | string | Artifact title |
| description | string | Artifact description |
| author | string | Artifact author |
| version | string | Artifact version |
| source_hash | string | Hash of source View |

---

## Renderer Types

### SVG Renderer

Renders View as SVG diagram.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| width | integer | 800 | Image width in pixels |
| height | integer | 600 | Image height in pixels |
| standalone | boolean | true | Include SVG headers |

### PNG Renderer

Renders View as PNG image.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| width | integer | 800 | Image width in pixels |
| height | integer | 600 | Image height in pixels |
| dpi | integer | 96 | Dots per inch |
| background | string | #ffffff | Background color |

### Mermaid Renderer

Renders View as Mermaid diagram.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| direction | enum | TB | Diagram direction (TB, LR, BT, RL) |
| theme | enum | default | Mermaid theme |
| title | string | - | Diagram title |

Lifted content (see the View Model spec) is rendered as follows:

- Lifted structural groups are emitted as `subgraph` blocks, declared before
  any edge that references them.
- Lifted edges are drawn with dashed arrows (`-.->`) to distinguish them from
  explicitly modeled Relations; symmetric lifted edges use dashed open links
  (`-.-`).
- An endpoint referencing a lifted group links against the subgraph ID.
- When a lifted edge collapsed multiple anchor candidates onto a
  representative, its label carries an aggregation suffix (e.g. `connects ×3`).

### Graphviz Renderer

Renders View as Graphviz DOT.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| layout | enum | dot | Layout engine (dot, neato, fdp, circo) |
| rankdir | enum | TB | Rank direction (TB, LR, BT, RL) |
| dpi | integer | 96 | Dots per inch |

### D2 Renderer

Renders View as D2 diagram.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| layout | enum | dagre | Layout engine (dagre, elk) |
| theme | integer | 0 | D2 theme ID |
| sketch | boolean | false | Hand-drawn style |

### HTML Renderer

Renders View as HTML page.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| template | string | - | Custom template path |
| css | string | - | Custom CSS path |
| interactive | boolean | true | Enable interactivity |

### Markdown Renderer

Renders View as Markdown document.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| toc | boolean | true | Include table of contents |
| diagrams | boolean | true | Include embedded diagrams |
| images | boolean | false | Include image links |

### JSON Renderer

Renders View as JSON document.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| indent | integer | 2 | Indentation spaces |
| sort_keys | boolean | false | Sort object keys |
| include_metadata | boolean | true | Include metadata |

### CSV Renderer

Renders View as CSV document.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| delimiter | string | , | Column delimiter |
| quote | string | " | Quote character |
| header | boolean | true | Include header row |

---

## Layout

Layout determines spatial arrangement of rendered elements.

### Layout Types

| Type | Description | Use Case |
|------|-------------|----------|
| hierarchical | Tree-like structure | Ownership, dependencies |
| force-directed | Physics-based simulation | Network topology |
| orthogonal | Right-angle connections | Infrastructure diagrams |
| rack | Physical rack view | Data center layout |
| timeline | Time-based arrangement | Change history |
| grid | Grid arrangement | Inventory views |

### Layout Configuration

| Field | Type | Description |
|-------|------|-------------|
| type | enum | Layout type |
| direction | enum | Layout direction |
| spacing | number | Element spacing |
| padding | number | View padding |
| alignment | enum | Element alignment |

### Layout Example

```yaml
layout:
  type: hierarchical
  direction: top-down
  spacing: 50
  padding: 20
  alignment: center
```

---

## Theme

Themes define presentation characteristics.

### Theme Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | yes | Unique theme identifier |
| name | string | yes | Human-readable name |
| colors | ColorPalette | no | Color definitions |
| typography | Typography | no | Font definitions |
| lines | LineStyles | no | Line style definitions |

### Color Palette

| Field | Type | Description |
|-------|------|-------------|
| primary | string | Primary color |
| secondary | string | Secondary color |
| background | string | Background color |
| surface | string | Surface color |
| text | string | Text color |
| border | string | Border color |
| success | string | Success state color |
| warning | string | Warning state color |
| error | string | Error state color |
| info | string | Info state color |

### Typography

| Field | Type | Description |
|-------|------|-------------|
| font_family | string | Font family |
| font_size | integer | Base font size |
| heading_size | integer | Heading font size |
| code_font | string | Monospace font |

### Line Styles

| Field | Type | Description |
|-------|------|-------------|
| default | LineStyle | Default line style |
| connection | LineStyle | Connection line style |
| ownership | LineStyle | Ownership line style |
| dependency | LineStyle | Dependency line style |

### Line Style

| Field | Type | Description |
|-------|------|-------------|
| color | string | Line color |
| width | number | Line width |
| style | enum | Line style (solid, dashed, dotted) |

### Theme Example

```yaml
theme:
  id: dark-theme
  name: Dark Theme
  colors:
    primary: "#3b82f6"
    secondary: "#6b7280"
    background: "#111827"
    surface: "#1f2937"
    text: "#f9fafb"
    border: "#374151"
  typography:
    font_family: "Inter, sans-serif"
    font_size: 14
    heading_size: 18
    code_font: "Fira Code, monospace"
```

---

## Interactivity

Renderers MAY generate interactive artifacts.

### Interactive Features

| Feature | Description |
|---------|-------------|
| expandable_groups | Click to expand/collapse groups |
| hyperlinks | Clickable links to related Objects |
| tooltips | Hover to show Object details |
| filtering | Filter displayed Objects |
| search | Search within artifact |
| animation | Animated transitions |
| zoom | Zoom in/out |
| pan | Pan across artifact |

### Interactivity Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| enabled | boolean | true | Enable interactivity |
| tooltips | boolean | true | Enable tooltips |
| hyperlinks | boolean | true | Enable hyperlinks |
| filtering | boolean | true | Enable filtering |
| search | boolean | true | Enable search |
| zoom | boolean | true | Enable zoom |

---

## Accessibility

Renderers SHOULD support accessible output.

### Accessibility Features

| Feature | Description |
|---------|-------------|
| alt_text | Alternative text for images |
| keyboard_nav | Keyboard navigation support |
| high_contrast | High contrast themes |
| scalable_text | Scalable text size |
| aria_labels | ARIA labels for elements |

### Accessibility Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| alt_text | boolean | true | Include alt text |
| keyboard_nav | boolean | true | Enable keyboard navigation |
| high_contrast | boolean | false | Use high contrast |
| aria_labels | boolean | true | Include ARIA labels |

---

## Rendering Pipeline

### Pipeline Stages

| Stage | Input | Output | Description |
|-------|-------|--------|-------------|
| 1 | Graph | Graph | Load Graph |
| 2 | Graph | Projection | Apply Projection |
| 3 | Projection | View | Apply View |
| 4 | View | Artifact | Render Artifact |

### Pipeline Configuration

```yaml
pipeline:
  graph: infrastructure.yaml
  projection: physical-topology
  view: rack-layout
  renderer: svg
  options:
    width: 1200
    height: 800
  theme: datacenter-theme
  output: rack-layout.svg
```

---

## Determinism

Rendering SHOULD be deterministic.

Rendering the same View with the same Renderer configuration SHOULD produce equivalent output.

Differences caused by timestamps, metadata, or implementation details MAY exist but SHOULD be minimized.

---

## Extensibility

Implementations MAY define additional Renderer capabilities.

Additional capabilities MUST preserve the semantics of the View.
