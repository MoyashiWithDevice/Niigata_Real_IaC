package renderer

import (
	"math"

	"IACForge/src/view"
)

// LayoutEngine computes spatial arrangement of elements.
type LayoutEngine struct {
	config *LayoutConfig
}

// NewLayoutEngine creates a new LayoutEngine.
func NewLayoutEngine(config *LayoutConfig) *LayoutEngine {
	if config == nil {
		config = &LayoutConfig{
			Type:      "hierarchical",
			Direction: "top-down",
			Spacing:   50,
			Padding:   20,
		}
	}
	return &LayoutEngine{config: config}
}

// ComputeLayout computes positions for all nodes and edges.
func (le *LayoutEngine) ComputeLayout(result *view.ViewResult) *LayoutResult {
	switch le.config.Type {
	case "force-directed":
		return le.computeForceDirectedLayout(result)
	default:
		return le.computeHierarchicalLayout(result)
	}
}

// nodeBox holds the measured size of a subtree in the containment layout.
type nodeBox struct {
	width  float64
	height float64
}

// computeHierarchicalLayout arranges nodes as nested containers following
// the ownership hierarchy. Child nodes are placed inside the bounding box of
// their parents; roots are laid out side by side.
func (le *LayoutEngine) computeHierarchicalLayout(result *view.ViewResult) *LayoutResult {
	spacing := le.config.Spacing
	padding := le.config.Padding

	if spacing == 0 {
		spacing = 50
	}
	if padding == 0 {
		padding = 20
	}

	const (
		nodeWidth    = 120.0
		nodeHeight   = 40.0
		headerHeight = 24.0
		innerPad     = 8.0
	)

	roots := buildOwnershipTree(result.VisibleEntities)

	// measure returns the bounding box needed to draw the node and its
	// descendants. Leaves are drawn at their natural size; containers wrap
	// their children laid out side by side below a label header.
	var measure func(node *OwnershipNode) *nodeBox
	measure = func(node *OwnershipNode) *nodeBox {
		if len(node.Children) == 0 {
			return &nodeBox{width: nodeWidth, height: nodeHeight}
		}
		width := 2*innerPad - spacing
		maxChildHeight := 0.0
		for _, child := range node.Children {
			childBox := measure(child)
			width += childBox.width + spacing
			if childBox.height > maxChildHeight {
				maxChildHeight = childBox.height
			}
		}
		if width < nodeWidth {
			width = nodeWidth
		}
		return &nodeBox{
			width:  width,
			height: headerHeight + maxChildHeight + innerPad,
		}
	}

	boxes := make(map[string]*nodeBox, len(result.VisibleEntities))
	for _, root := range roots {
		collectBoxes(root, measure, boxes)
	}

	var layoutResult LayoutResult
	layoutResult.Nodes = make([]NodePosition, 0)
	layoutResult.Edges = make([]EdgePosition, 0)

	var place func(node *OwnershipNode, x, y float64)
	place = func(node *OwnershipNode, x, y float64) {
		box := boxes[node.Entity.ID]
		children := make([]string, 0, len(node.Children))
		for _, child := range node.Children {
			children = append(children, child.Entity.ID)
		}
		layoutResult.Nodes = append(layoutResult.Nodes, NodePosition{
			ID:       node.Entity.ID,
			Position: Position{X: x, Y: y},
			Width:    box.width,
			Height:   box.height,
			Children: children,
		})

		childX := x + innerPad
		childY := y + headerHeight
		for _, child := range node.Children {
			place(child, childX, childY)
			childX += boxes[child.Entity.ID].width + spacing
		}
	}

	x := padding
	for _, root := range roots {
		place(root, x, padding)
		x += boxes[root.Entity.ID].width + spacing
	}

	for i := range layoutResult.Nodes {
		node := &layoutResult.Nodes[i]
		if right := node.Position.X + node.Width; right > layoutResult.Width {
			layoutResult.Width = right
		}
		if bottom := node.Position.Y + node.Height; bottom > layoutResult.Height {
			layoutResult.Height = bottom
		}
	}
	layoutResult.Width += padding
	layoutResult.Height += padding

	for _, rel := range result.VisibleRelations {
		sourcePos := le.findNodePosition(layoutResult.Nodes, rel.Source())
		targetPos := le.findNodePosition(layoutResult.Nodes, rel.Target())
		if sourcePos != nil && targetPos != nil {
			edge := EdgePosition{
				ID:     rel.ID,
				Source: rel.Source(),
				Target: rel.Target(),
				Points: []Position{
					{X: sourcePos.Position.X + sourcePos.Width/2, Y: sourcePos.Position.Y + sourcePos.Height},
					{X: targetPos.Position.X + targetPos.Width/2, Y: targetPos.Position.Y},
				},
			}
			layoutResult.Edges = append(layoutResult.Edges, edge)
		}
	}

	return &layoutResult
}

// collectBoxes records the measured size of every node in the subtree.
func collectBoxes(node *OwnershipNode, measure func(*OwnershipNode) *nodeBox, boxes map[string]*nodeBox) {
	boxes[node.Entity.ID] = measure(node)
	for _, child := range node.Children {
		collectBoxes(child, measure, boxes)
	}
}

// computeForceDirectedLayout arranges nodes using a physics simulation.
func (le *LayoutEngine) computeForceDirectedLayout(result *view.ViewResult) *LayoutResult {
	nodeCount := len(result.VisibleEntities)
	if nodeCount == 0 {
		return &LayoutResult{
			Nodes:  []NodePosition{},
			Edges:  []EdgePosition{},
			Width:  0,
			Height: 0,
		}
	}

	nodeWidth := 120.0
	nodeHeight := 40.0
	spacing := le.config.Spacing
	if spacing == 0 {
		spacing = 100
	}

	nodes := make([]NodePosition, 0, nodeCount)
	for i, entity := range result.VisibleEntities {
		angle := 2 * math.Pi * float64(i) / float64(nodeCount)
		radius := spacing * math.Sqrt(float64(nodeCount))
		nodes = append(nodes, NodePosition{
			ID: entity.ID,
			Position: Position{
				X: radius * math.Cos(angle),
				Y: radius * math.Sin(angle),
			},
			Width:  nodeWidth,
			Height: nodeHeight,
		})
	}

	for iter := 0; iter < 50; iter++ {
		le.applyForces(nodes, result)
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for _, node := range nodes {
		if node.Position.X < minX {
			minX = node.Position.X
		}
		if node.Position.Y < minY {
			minY = node.Position.Y
		}
		if node.Position.X+node.Width > maxX {
			maxX = node.Position.X + node.Width
		}
		if node.Position.Y+node.Height > maxY {
			maxY = node.Position.Y + node.Height
		}
	}

	padding := le.config.Padding
	if padding == 0 {
		padding = 20
	}

	for i := range nodes {
		nodes[i].Position.X -= minX - padding
		nodes[i].Position.Y -= minY - padding
	}

	var layoutResult LayoutResult
	layoutResult.Nodes = nodes
	layoutResult.Edges = make([]EdgePosition, 0)
	layoutResult.Width = maxX - minX + 2*padding
	layoutResult.Height = maxY - minY + 2*padding

	for _, rel := range result.VisibleRelations {
		edge := EdgePosition{
			ID:     rel.ID,
			Source: rel.Source(),
			Target: rel.Target(),
		}
		layoutResult.Edges = append(layoutResult.Edges, edge)
	}

	return &layoutResult
}

// applyForces applies repulsive and attractive forces.
func (le *LayoutEngine) applyForces(nodes []NodePosition, result *view.ViewResult) {
	repulsion := 1000.0
	attraction := 0.01
	damping := 0.9

	velocities := make([]Position, len(nodes))

	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			dx := nodes[j].Position.X - nodes[i].Position.X
			dy := nodes[j].Position.Y - nodes[i].Position.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 1 {
				dist = 1
			}

			force := repulsion / (dist * dist)
			fx := force * dx / dist
			fy := force * dy / dist

			velocities[i].X -= fx
			velocities[i].Y -= fy
			velocities[j].X += fx
			velocities[j].Y += fy
		}
	}

	for _, rel := range result.VisibleRelations {
		sourceIdx := le.findNodeIndex(nodes, rel.Source())
		targetIdx := le.findNodeIndex(nodes, rel.Target())
		if sourceIdx >= 0 && targetIdx >= 0 {
			dx := nodes[targetIdx].Position.X - nodes[sourceIdx].Position.X
			dy := nodes[targetIdx].Position.Y - nodes[sourceIdx].Position.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 1 {
				dist = 1
			}

			force := attraction * dist
			fx := force * dx / dist
			fy := force * dy / dist

			velocities[sourceIdx].X += fx
			velocities[sourceIdx].Y += fy
			velocities[targetIdx].X -= fx
			velocities[targetIdx].Y -= fy
		}
	}

	for i := range nodes {
		nodes[i].Position.X += velocities[i].X * damping
		nodes[i].Position.Y += velocities[i].Y * damping
	}
}

// findNodePosition finds a node position by ID.
func (le *LayoutEngine) findNodePosition(nodes []NodePosition, id string) *NodePosition {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
	}
	return nil
}

// findNodeIndex finds a node index by ID.
func (le *LayoutEngine) findNodeIndex(nodes []NodePosition, id string) int {
	for i := range nodes {
		if nodes[i].ID == id {
			return i
		}
	}
	return -1
}
