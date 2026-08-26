package renderer

import (
	"testing"

	"IACForge/src/core"
)

func entitiesForTree() []*core.Entity {
	region := core.NewEntity("region-1", "region", "Region 1")
	rack := core.NewEntity("rack-1", "rack", "Rack 1")
	rack.SetOwner("region-1")
	server := core.NewEntity("srv-1", "server", "Server 1")
	server.SetOwner("rack-1")
	return []*core.Entity{region, rack, server}
}

func TestBuildOwnershipTreeNested(t *testing.T) {
	roots := buildOwnershipTree(entitiesForTree())
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}

	region := roots[0]
	if region.Entity.ID != "region-1" {
		t.Errorf("expected root region-1, got %s", region.Entity.ID)
	}
	if len(region.Children) != 1 {
		t.Fatalf("expected 1 child under region, got %d", len(region.Children))
	}

	rack := region.Children[0]
	if rack.Entity.ID != "rack-1" {
		t.Errorf("expected child rack-1, got %s", rack.Entity.ID)
	}
	if len(rack.Children) != 1 || rack.Children[0].Entity.ID != "srv-1" {
		t.Error("expected server nested under rack")
	}
}

func TestBuildOwnershipTreeOrphanOwner(t *testing.T) {
	region := core.NewEntity("region-1", "region", "Region 1")
	orphan := core.NewEntity("srv-1", "server", "Server 1")
	orphan.SetOwner("hidden-owner")

	roots := buildOwnershipTree([]*core.Entity{region, orphan})
	if len(roots) != 2 {
		t.Fatalf("expected orphan to become a root, got %d roots", len(roots))
	}
	for _, root := range roots {
		if root.Entity.ID == "srv-1" && len(root.Children) != 0 {
			t.Error("orphan should have no children")
		}
	}
}

func TestBuildOwnershipTreeCycle(t *testing.T) {
	a := core.NewEntity("a", "server", "A")
	b := core.NewEntity("b", "server", "B")
	a.SetOwner("b")
	b.SetOwner("a")

	roots := buildOwnershipTree([]*core.Entity{a, b})
	if len(roots) != 2 {
		t.Fatalf("expected cycle members to be treated as roots, got %d roots", len(roots))
	}
}

func TestBuildOwnershipTreeSelfOwner(t *testing.T) {
	self := core.NewEntity("a", "server", "A")
	self.SetOwner("a")

	roots := buildOwnershipTree([]*core.Entity{self})
	if len(roots) != 1 || roots[0].Entity.ID != "a" {
		t.Fatal("expected self-owned entity to be a single root")
	}
}

func TestBuildOwnershipTreePreservesOrder(t *testing.T) {
	first := core.NewEntity("r-1", "region", "R1")
	second := core.NewEntity("r-2", "region", "R2")

	roots := buildOwnershipTree([]*core.Entity{first, second})
	if len(roots) != 2 || roots[0].Entity.ID != "r-1" || roots[1].Entity.ID != "r-2" {
		t.Fatal("expected input order preserved among roots")
	}
}
