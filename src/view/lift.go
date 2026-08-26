package view

import (
	"fmt"
	"sort"

	"IACForge/src/core"
)

// Lifted group marker values.
const (
	liftGroupKind       = "lift"
	liftGroupDerivedKey = "derived"
)

// LiftRelations derives edges between visible entities from graph relations
// whose participants are not (all) visible. It lets diagrams such as
// "all applications" express connectivity that the model defines on hidden
// objects: host nodes, clusters, sites, or other ancestors.
//
// Each hidden relation participant is mapped to visible anchor entities:
//
//  1. If the participant itself is visible, it is its own anchor.
//  2. Otherwise every visible entity inside its ownership subtree becomes an
//     anchor (e.g. a VM maps to the applications it hosts; a cluster maps to
//     all applications beneath it).
//  3. If neither applies, the nearest visible ancestor is used (e.g. an open
//     port maps to the application owning it).
//
// The two sides of a relation are then collapsed so that one relation yields
// at most one edge:
//
//   - A single anchor is used directly.
//   - Multiple anchors sharing a common hidden ancestor whose visible subtree
//     is exactly the anchor set collapse onto a structural group (returned as
//     a LiftedGroups entry keyed by the ancestor entity ID).
//   - Otherwise the first anchor (sorted by ID) acts as a deterministic
//     representative and AggregatedCount records how many candidates were
//     folded in.
//
// Pairs collapsing onto themselves (same host, same subtree) produce no edge,
// and edges duplicating an existing direct relation between visible endpoints
// are suppressed via the existing parameter.
//
// Relations whose participants are all directly visible are ignored: they are
// either already part of the visible subgraph or intentionally hidden by view
// rules.
//
// The source Graph is never modified. Output ordering is deterministic.
func LiftRelations(g *core.Graph, visibleEntities []*core.Entity, existing []*core.Relation, sourceRelations []*core.Relation) ([]*Group, []*LiftedRelation) {
	if g == nil || len(visibleEntities) == 0 || len(sourceRelations) == 0 {
		return nil, nil
	}

	l := &relationLifter{
		graph:        g,
		visible:      make(map[string]bool, len(visibleEntities)),
		subtreeCache: make(map[string][]string),
	}
	for _, e := range visibleEntities {
		if e != nil {
			l.visible[e.ID] = true
		}
	}

	groups := make(map[string]*Group)
	type edgeKey struct {
		relType core.RelationType
		src     string
		dst     string
	}
	edges := make(map[edgeKey]*LiftedRelation)
	addVia := func(e *LiftedRelation, relID string) {
		for _, v := range e.Via {
			if v == relID {
				return
			}
		}
		e.Via = append(e.Via, relID)
	}

	// Suppress lifted edges that duplicate already-visible direct relations.
	directKeys := make(map[edgeKey]bool)
	for _, rel := range existing {
		parts := l.resolveParticipants(rel)
		if len(parts) < 2 {
			continue
		}
		directKeys[edgeKey{rel.Type, parts[0], parts[1]}] = true
	}

	for _, rel := range sourceRelations {
		parts := l.resolveParticipants(rel)
		if len(parts) < 2 {
			continue
		}

		allVisible := true
		for _, p := range parts {
			if !l.visible[p] {
				allVisible = false
				break
			}
		}
		if allVisible {
			continue
		}

		srcOrigin, srcAnchors := l.anchors(parts[0])
		dstOrigin, dstAnchors := l.anchors(parts[1])
		if len(srcAnchors) == 0 || len(dstAnchors) == 0 {
			continue
		}

		src := l.collapseSide(srcOrigin, srcAnchors, groups)
		dst := l.collapseSide(dstOrigin, dstAnchors, groups)
		if src.ref == dst.ref {
			continue
		}
		if sideContains(groups, src.ref, dst.ref) || sideContains(groups, dst.ref, src.ref) {
			continue
		}

		key := edgeKey{rel.Type, src.ref, dst.ref}
		if directKeys[key] {
			continue
		}
		if e, ok := edges[key]; ok {
			addVia(e, rel.ID)
			continue
		}
		aggregated := 0
		if src.fallback {
			aggregated = src.count
		}
		if dst.fallback && dst.count > aggregated {
			aggregated = dst.count
		}
		direction := rel.Direction
		if direction == "" {
			direction = core.DirectionDirected
		}
		edges[key] = &LiftedRelation{
			ID:              fmt.Sprintf("lifted-%s-%s-to-%s", rel.Type, src.ref, dst.ref),
			Type:            rel.Type,
			Direction:       direction,
			SourceRef:       src.ref,
			TargetRef:       dst.ref,
			AggregatedCount: aggregated,
			Via:             []string{rel.ID},
		}
	}

	if len(edges) == 0 {
		return nil, nil
	}

	// Keep only groups actually referenced by a lifted edge.
	referenced := make(map[string]bool, 2)
	for _, e := range edges {
		referenced[e.SourceRef] = true
		referenced[e.TargetRef] = true
	}

	resultGroups := make([]*Group, 0, len(groups))
	for id, grp := range groups {
		if referenced[id] {
			resultGroups = append(resultGroups, grp)
		}
	}
	sort.Slice(resultGroups, func(i, j int) bool {
		return resultGroups[i].ID < resultGroups[j].ID
	})

	resultEdges := make([]*LiftedRelation, 0, len(edges))
	for _, e := range edges {
		resultEdges = append(resultEdges, e)
	}
	sort.Slice(resultEdges, func(i, j int) bool {
		if resultEdges[i].SourceRef != resultEdges[j].SourceRef {
			return resultEdges[i].SourceRef < resultEdges[j].SourceRef
		}
		if resultEdges[i].TargetRef != resultEdges[j].TargetRef {
			return resultEdges[i].TargetRef < resultEdges[j].TargetRef
		}
		if resultEdges[i].Type != resultEdges[j].Type {
			return resultEdges[i].Type < resultEdges[j].Type
		}
		return resultEdges[i].ID < resultEdges[j].ID
	})

	return resultGroups, resultEdges
}

type sideEndpoint struct {
	ref      string // entity ID or structural group ID
	count    int    // number of anchors collapsed into ref
	fallback bool   // representative used instead of a structural group
}

type relationLifter struct {
	graph        *core.Graph
	visible      map[string]bool
	subtreeCache map[string][]string
}

// resolveParticipants resolves a relation's participant references to entity
// IDs, supporting direct IDs and path/interface notation.
func (l *relationLifter) resolveParticipants(rel *core.Relation) []string {
	resolve := func(ref string) (string, bool) {
		if e, ok := l.graph.GetEntity(ref); ok {
			return e.ID, true
		}
		if e, ok := l.graph.ResolvePathEntity(ref); ok {
			return e.ID, true
		}
		return "", false
	}

	switch {
	case len(rel.Participants.List) > 0:
		ids := make([]string, 0, len(rel.Participants.List))
		for _, pid := range rel.Participants.List {
			id, ok := resolve(pid)
			if !ok {
				return nil
			}
			ids = append(ids, id)
		}
		return ids
	case rel.Participants.Source != "" && rel.Participants.Target != "":
		src, ok1 := resolve(rel.Participants.Source)
		dst, ok2 := resolve(rel.Participants.Target)
		if !ok1 || !ok2 {
			return nil
		}
		return []string{src, dst}
	default:
		return nil
	}
}

// anchors maps a participant reference to its origin entity ID (the object
// the participant resolved to) plus the visible anchor entity IDs.
func (l *relationLifter) anchors(ref string) (string, []string) {
	e, ok := l.graph.GetEntity(ref)
	if !ok {
		if e, ok = l.graph.ResolvePathEntity(ref); !ok {
			return "", nil
		}
	}
	if l.visible[e.ID] {
		return e.ID, []string{e.ID}
	}
	if ids := l.subtreeVisible(e.ID); len(ids) > 0 {
		return e.ID, ids
	}
	for _, anc := range l.graph.Ancestors(e.ID) {
		if l.visible[anc.ID] {
			return anc.ID, []string{anc.ID}
		}
	}
	return "", nil
}

// subtreeVisible returns the sorted visible entities within the ownership
// subtree rooted at id (id itself is never included; callers check visibility
// separately).
func (l *relationLifter) subtreeVisible(id string) []string {
	if cached, ok := l.subtreeCache[id]; ok {
		return cached
	}

	seen := make(map[string]bool)
	seen[id] = true
	var result []string

	var collect func(current string)
	collect = func(current string) {
		for _, child := range l.graph.Children(current) {
			if seen[child.ID] {
				continue
			}
			seen[child.ID] = true
			if l.visible[child.ID] {
				result = append(result, child.ID)
			}
			collect(child.ID)
		}
	}
	collect(id)

	sort.Strings(result)
	l.subtreeCache[id] = result
	return result
}

// commonStructuralAncestor returns the nearest hidden ancestor shared by all
// anchors whose visible subtree equals exactly the anchor set, or "" when no
// such grouping structure exists.
func (l *relationLifter) commonStructuralAncestor(anchors []string) string {
	common := make(map[string]int)
	for i, aid := range anchors {
		chain := l.graph.Ancestors(aid)
		depths := make(map[string]int, len(chain))
		for depth, anc := range chain {
			depths[anc.ID] = depth + 1
		}
		if i == 0 {
			common = depths
			continue
		}
		for id := range common {
			if _, ok := depths[id]; !ok {
				delete(common, id)
			}
		}
		if len(common) == 0 {
			return ""
		}
	}

	best := ""
	bestDepth := -1
	for id, depth := range common {
		if bestDepth != -1 && depth >= bestDepth {
			continue
		}
		if l.visible[id] {
			continue
		}
		if e, ok := l.graph.GetEntity(id); !ok || e.IsRoot() {
			continue
		}
		subtree := l.subtreeVisible(id)
		if equalStrings(subtree, anchors) {
			best = id
			bestDepth = depth
		}
	}
	return best
}

// collapseSide reduces an anchor set to a single edge endpoint.
//
// When several anchors were derived from a hidden origin object whose visible
// subtree matches the anchors exactly (the usual case), the origin itself is
// preferred as the structural group so relations stay anchored on the object
// they were defined on. Otherwise the nearest qualifying common ancestor is
// used; as a last resort the first anchor acts as a deterministic
// representative.
func (l *relationLifter) collapseSide(origin string, anchors []string, groups map[string]*Group) sideEndpoint {
	if len(anchors) == 1 {
		return sideEndpoint{ref: anchors[0], count: 1}
	}

	candidate := ""
	if origin != "" && !l.visible[origin] && equalStrings(l.subtreeVisible(origin), anchors) {
		candidate = origin
	} else {
		candidate = l.commonStructuralAncestor(anchors)
	}

	if candidate != "" {
		gid := candidate
		name := gid
		if e, ok := l.graph.GetEntity(gid); ok {
			name = e.Name
		}
		if grp, ok := groups[gid]; ok {
			grp.Members = mergeMembers(grp.Members, anchors)
		} else {
			groups[gid] = &Group{
				ID:      gid,
				Kind:    liftGroupKind,
				Name:    name,
				Members: append([]string{}, anchors...),
				Properties: map[string]interface{}{
					liftGroupDerivedKey: true,
				},
			}
		}
		return sideEndpoint{ref: gid, count: len(anchors)}
	}

	return sideEndpoint{ref: anchors[0], count: len(anchors), fallback: true}
}

// sideContains reports whether ref points at a structural group whose members
// include target.
func sideContains(groups map[string]*Group, ref, target string) bool {
	grp, ok := groups[ref]
	if !ok {
		return false
	}
	for _, m := range grp.Members {
		if m == target {
			return true
		}
	}
	return false
}

func mergeMembers(existing, anchors []string) []string {
	seen := make(map[string]bool, len(existing)+len(anchors))
	for _, m := range existing {
		seen[m] = true
	}
	for _, a := range anchors {
		if !seen[a] {
			existing = append(existing, a)
			seen[a] = true
		}
	}
	sort.Strings(existing)
	return existing
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
