package renderer

import (
	"IACForge/src/core"
)

// OwnershipNode represents a node in the ownership hierarchy derived from
// the Owner field of visible entities.
type OwnershipNode struct {
	Entity   *core.Entity
	Children []*OwnershipNode
}

// buildOwnershipTree builds an ownership forest from visible entities.
//
// An entity whose Owner references another visible entity becomes a child of
// that entity. Entities with no owner, or with an owner that is not part of
// the visible set, become roots. Entities caught in an ownership cycle among
// visible entities are defensively treated as roots so rendering always
// terminates. Child order follows the input order of entities.
func buildOwnershipTree(entities []*core.Entity) []*OwnershipNode {
	nodes := make(map[string]*OwnershipNode, len(entities))
	for _, entity := range entities {
		nodes[entity.ID] = &OwnershipNode{Entity: entity}
	}

	var roots []*OwnershipNode
	for _, entity := range entities {
		node := nodes[entity.ID]
		if ownerNode, ok := nodes[entity.Owner]; ok && entity.Owner != entity.ID {
			ownerNode.Children = append(ownerNode.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	reachable := make(map[*OwnershipNode]struct{})
	var mark func(node *OwnershipNode)
	mark = func(node *OwnershipNode) {
		if _, seen := reachable[node]; seen {
			return
		}
		reachable[node] = struct{}{}
		for _, child := range node.Children {
			mark(child)
		}
	}
	for _, root := range roots {
		mark(root)
	}

	var cycleRoots []*OwnershipNode
	for _, entity := range entities {
		node := nodes[entity.ID]
		if _, seen := reachable[node]; !seen {
			cycleRoots = append(cycleRoots, node)
		}
	}
	for _, node := range cycleRoots {
		mark(node)
	}

	return append(roots, cycleRoots...)
}
