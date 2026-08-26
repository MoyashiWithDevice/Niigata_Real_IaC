package extension

import (
	"IACForge/src/renderer"
)

// RendererRegistry provides access to registered renderers.
type RendererRegistry interface {
	Register(r renderer.Renderer)
	Get(id string) (renderer.Renderer, bool)
	All() []renderer.Renderer
}

// rendererAdapter wraps a simple renderer registry for use with extensions.
type rendererAdapter struct {
	registry map[string]renderer.Renderer
}

// NewRendererRegistry creates a new RendererRegistry.
func NewRendererRegistry() RendererRegistry {
	return &rendererAdapter{
		registry: make(map[string]renderer.Renderer),
	}
}

// Register adds a renderer to the registry.
func (ra *rendererAdapter) Register(r renderer.Renderer) {
	ra.registry[r.ID()] = r
}

// Get returns a renderer by ID.
func (ra *rendererAdapter) Get(id string) (renderer.Renderer, bool) {
	r, ok := ra.registry[id]
	return r, ok
}

// All returns all registered renderers.
func (ra *rendererAdapter) All() []renderer.Renderer {
	result := make([]renderer.Renderer, 0, len(ra.registry))
	for _, r := range ra.registry {
		result = append(result, r)
	}
	return result
}

// RendererExtensionPoint manages renderer extensions.
type RendererExtensionPoint struct {
	registry  RendererRegistry
	renderers map[string]string // renderer ID -> extension ID that registered it
}

// NewRendererExtensionPoint creates a new renderer extension point.
func NewRendererExtensionPoint(registry RendererRegistry) *RendererExtensionPoint {
	return &RendererExtensionPoint{
		registry:  registry,
		renderers: make(map[string]string),
	}
}

// Type returns the extension point type.
func (ep *RendererExtensionPoint) Type() ExtensionPointType {
	return ExtensionPointRenderers
}

// Register registers all renderer contributions from the given extension.
func (ep *RendererExtensionPoint) Register(ext *Extension) error {
	for _, contrib := range ext.Renderers {
		if contrib.Renderer == nil {
			continue
		}
		ep.registry.Register(contrib.Renderer)
		ep.renderers[contrib.Renderer.ID()] = ext.Manifest.ID
	}
	return nil
}

// GetRenderersByExtension returns all renderer IDs registered by a specific extension.
func (ep *RendererExtensionPoint) GetRenderersByExtension(extensionID string) []string {
	var result []string
	for rendererID, extID := range ep.renderers {
		if extID == extensionID {
			result = append(result, rendererID)
		}
	}
	return result
}

// GetExtensionForRenderer returns the extension ID that registered the given renderer.
func (ep *RendererExtensionPoint) GetExtensionForRenderer(rendererID string) (string, bool) {
	extID, ok := ep.renderers[rendererID]
	return extID, ok
}
