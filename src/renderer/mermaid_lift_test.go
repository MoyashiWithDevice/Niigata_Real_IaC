package renderer

import (
	"strings"
	"testing"

	"IACForge/src/core"
	"IACForge/src/view"
)

// buildLiftedViewResult returns a ViewResult simulating an application-only
// diagram enriched by the relation lift step: a structural group "vm-web"
// containing two applications, and lifted edges to a remote application.
func buildLiftedViewResult() *view.ViewResult {
	appWeb := core.NewEntity("app-web", "application", "App Web")
	appCache := core.NewEntity("app-cache", "application", "App Cache")
	appDb := core.NewEntity("app-db", "application", "App DB")

	vr := &view.ViewResult{
		ViewID:          "apps-only",
		Title:           "Applications",
		VisibleEntities: []*core.Entity{appCache, appDb, appWeb},
		LiftedGroups: []*view.Group{
			{
				ID:      "vm-web",
				Kind:    "lift",
				Name:    "VM Web",
				Members: []string{"app-cache", "app-web"},
				Properties: map[string]interface{}{
					"derived": true,
				},
			},
		},
		LiftedRelations: []*view.LiftedRelation{
			{
				ID:        "lifted-depends_on-vm-web-to-app-db",
				Type:      "depends_on",
				Direction: core.DirectionDirected,
				SourceRef: "vm-web",
				TargetRef: "app-db",
				Via:       []string{"rel-vm-dep"},
			},
			{
				ID:              "lifted-connects-app-cache-to-app-db",
				Type:            "connects",
				Direction:       core.DirectionSymmetric,
				SourceRef:       "app-cache",
				TargetRef:       "app-db",
				AggregatedCount: 3,
				Via:             []string{"rel-x"},
			},
		},
	}
	return vr
}

func TestMermaidRendererLiftedGroupsAndEdges(t *testing.T) {
	artifact, err := NewMermaidRenderer().Render(buildLiftedViewResult(), nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	content := artifact.Content

	if !strings.Contains(content, `subgraph vm_web["VM Web"]`) {
		t.Errorf("expected lifted group subgraph, got:\n%s", content)
	}

	subgraphIdx := strings.Index(content, "subgraph vm_web")
	dashedIdx := strings.Index(content, "-.->")
	if subgraphIdx == -1 || dashedIdx == -1 || dashedIdx < subgraphIdx {
		t.Errorf("expected subgraph declared before lifted edges, got:\n%s", content)
	}

	if !strings.Contains(content, "vm_web -.->|depends_on| app_db") {
		t.Errorf("expected dashed lifted edge, got:\n%s", content)
	}

	if !strings.Contains(content, "app_cache -.-|connects ×3| app_db") {
		t.Errorf("expected symmetric dashed edge with aggregation label, got:\n%s", content)
	}
}

func TestMermaidRendererSkipsUnresolvableLiftedRefs(t *testing.T) {
	vr := buildLiftedViewResult()
	vr.LiftedRelations = append(vr.LiftedRelations, &view.LiftedRelation{
		ID:        "lifted-ghost",
		Type:      "depends_on",
		Direction: core.DirectionDirected,
		SourceRef: "does-not-exist",
		TargetRef: "app-db",
	})

	artifact, err := NewMermaidRenderer().Render(vr, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if strings.Contains(artifact.Content, "does_not_exist") {
		t.Errorf("expected unresolvable lifted edge to be skipped, got:\n%s", artifact.Content)
	}
}
