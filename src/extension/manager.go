package extension

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sort"
	"strings"

	"IACForge/src/core"
)

var (
	ErrExtensionAlreadyLoaded = errors.New("extension already loaded")
	ErrDependencyMissing      = errors.New("missing dependency")
	ErrCircularDependency     = errors.New("circular dependency detected")
	ErrCoreConflict           = errors.New("extension conflicts with core definition")
	ErrNamespaceConflict      = errors.New("namespace conflict with existing extension")
	ErrInvalidExtension       = errors.New("invalid extension")
)

// Manager manages extensions, their loading, and dependency resolution.
type Manager struct {
	extensions      map[string]*Extension
	extensionPoints map[ExtensionPointType]ExtensionPoint
	loadOrder       []string
	loaded          bool
	applied         map[string]bool
}

// NewManager creates a new Extension Manager.
func NewManager() *Manager {
	return &Manager{
		extensions:      make(map[string]*Extension),
		extensionPoints: make(map[ExtensionPointType]ExtensionPoint),
		applied:         make(map[string]bool),
	}
}

// RegisterExtensionPoint registers an extension point with the manager.
func (m *Manager) RegisterExtensionPoint(ep ExtensionPoint) {
	m.extensionPoints[ep.Type()] = ep
}

// GetExtensionPoint returns the extension point of the given type.
func (m *Manager) GetExtensionPoint(t ExtensionPointType) (ExtensionPoint, bool) {
	ep, ok := m.extensionPoints[t]
	return ep, ok
}

// Register adds an extension to the manager without loading it into extension points.
// It validates the manifest and checks for conflicts.
func (m *Manager) Register(ext *Extension) error {
	if ext == nil || ext.Manifest == nil {
		return ErrInvalidExtension
	}

	if ext.Manifest.ID == "" {
		return fmt.Errorf("%w: extension ID is required", ErrInvalidExtension)
	}

	if _, exists := m.extensions[ext.Manifest.ID]; exists {
		return fmt.Errorf("%w: %s", ErrExtensionAlreadyLoaded, ext.Manifest.ID)
	}

	if err := m.validateNoCoreConflict(ext); err != nil {
		return err
	}

	if err := m.validateNamespace(ext); err != nil {
		return err
	}

	m.extensions[ext.Manifest.ID] = ext
	return nil
}

// LoadAll resolves dependencies, orders extensions, and applies them to extension points.
// LoadAll is idempotent: extensions that have already been applied are skipped,
// so it may be called again after registering additional extensions (e.g. via
// LoadFromDir at runtime) without re-applying previously loaded extensions.
func (m *Manager) LoadAll() error {
	order, err := m.resolveLoadOrder()
	if err != nil {
		return err
	}

	for _, extID := range order {
		if m.applied[extID] {
			continue
		}
		ext := m.extensions[extID]
		if err := m.applyExtension(ext); err != nil {
			return fmt.Errorf("loading extension %s: %w", extID, err)
		}
		m.applied[extID] = true
	}

	m.loadOrder = order
	m.loaded = true
	return nil
}

// Load is a convenience method: Register + LoadAll.
func (m *Manager) Load(ext *Extension) error {
	if err := m.Register(ext); err != nil {
		return err
	}
	return m.LoadAll()
}

// IsLoaded returns whether LoadAll has been called successfully.
func (m *Manager) IsLoaded() bool {
	return m.loaded
}

// GetExtension returns the extension with the given ID.
func (m *Manager) GetExtension(id string) (*Extension, bool) {
	ext, ok := m.extensions[id]
	return ext, ok
}

// Extensions returns all registered extensions.
func (m *Manager) Extensions() []*Extension {
	result := make([]*Extension, 0, len(m.extensions))
	for _, ext := range m.extensions {
		result = append(result, ext)
	}
	return result
}

// LoadOrder returns the resolved load order of extensions.
func (m *Manager) LoadOrder() []string {
	result := make([]string, len(m.loadOrder))
	copy(result, m.loadOrder)
	return result
}

// LoadFromDir loads all plugins from the specified directory.
// Each plugin must be a Go plugin (.so file) that exports an Extension() function.
// The directory is scanned for .so files, and each is loaded as a plugin.
// After loading all plugins, LoadAll is called to resolve dependencies and apply them.
func (m *Manager) LoadFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read extension directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".so") {
			continue
		}

		pluginPath := filepath.Join(dir, entry.Name())
		ext, err := loadPlugin(pluginPath)
		if err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", pluginPath, err)
		}

		if err := m.Register(ext); err != nil {
			return fmt.Errorf("failed to register plugin %s: %w", pluginPath, err)
		}
	}

	return m.LoadAll()
}

// loadPlugin loads a single Go plugin from the given path and extracts the Extension.
func loadPlugin(path string) (*Extension, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("plugin.Open failed: %w", err)
	}

	extFunc, err := p.Lookup("Extension")
	if err != nil {
		return nil, fmt.Errorf("plugin does not export Extension function: %w", err)
	}

	// Assert that Extension is a function
	fn, ok := extFunc.(func() *Extension)
	if !ok {
		return nil, fmt.Errorf("Extension is not of type func() *Extension")
	}

	return fn(), nil
}

// Namespaces returns all registered namespaces.
func (m *Manager) Namespaces() []string {
	seen := make(map[string]bool)
	var namespaces []string
	for _, ext := range m.extensions {
		ns := ext.Manifest.Namespace
		if ns != "" && !seen[ns] {
			seen[ns] = true
			namespaces = append(namespaces, ns)
		}
	}
	sort.Strings(namespaces)
	return namespaces
}

// validateNoCoreConflict checks that the extension does not redefine core entity kinds or relation types.
func (m *Manager) validateNoCoreConflict(ext *Extension) error {
	for _, ek := range ext.EntityKinds {
		if isCoreEntityKind(ek.Kind) {
			return fmt.Errorf("%w: entity kind %q is a core kind", ErrCoreConflict, ek.Kind)
		}
	}
	for _, rt := range ext.RelationTypes {
		if rt.Augment {
			// Augmenting an existing type (e.g. adding kinds to a core type) is allowed.
			continue
		}
		if isCoreRelationType(rt.Type) {
			return fmt.Errorf("%w: relation type %q is a core type", ErrCoreConflict, rt.Type)
		}
	}
	return nil
}

// validateNamespace checks that the extension's namespace does not conflict with existing extensions.
func (m *Manager) validateNamespace(ext *Extension) error {
	ns := ext.Manifest.Namespace
	if ns == "" {
		return nil
	}
	for _, existing := range m.extensions {
		if existing.Manifest.Namespace == ns && existing.Manifest.ID != ext.Manifest.ID {
			return fmt.Errorf("%w: namespace %q already used by extension %s", ErrNamespaceConflict, ns, existing.Manifest.ID)
		}
	}
	return nil
}

// resolveLoadOrder performs topological sort of extensions based on dependencies.
func (m *Manager) resolveLoadOrder() ([]string, error) {
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for id := range m.extensions {
		if _, ok := graph[id]; !ok {
			graph[id] = []string{}
		}
		if _, ok := inDegree[id]; !ok {
			inDegree[id] = 0
		}
	}

	for id, ext := range m.extensions {
		for _, depID := range ext.Manifest.Dependencies {
			if _, exists := m.extensions[depID]; !exists {
				return nil, fmt.Errorf("%w: extension %s requires %s which is not registered", ErrDependencyMissing, id, depID)
			}
			graph[depID] = append(graph[depID], id)
			inDegree[id]++
		}
	}

	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	var order []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		for _, neighbor := range graph[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
				sort.Strings(queue)
			}
		}
	}

	if len(order) != len(m.extensions) {
		return nil, ErrCircularDependency
	}

	return order, nil
}

// applyExtension applies an extension to all registered extension points.
func (m *Manager) applyExtension(ext *Extension) error {
	for pointType, ep := range m.extensionPoints {
		if err := ep.Register(ext); err != nil {
			return fmt.Errorf("extension point %s: %w", pointType, err)
		}
	}
	return nil
}

// coreEntityKinds lists all entity kinds defined by the core specification.
var coreEntityKinds = map[string]bool{
	"region": true, "rack": true, "server": true, "interface": true,
	"cable": true, "power_distribution": true, "network": true, "vlan": true,
	"switch": true, "router": true, "firewall": true, "acl": true,
	"acl_rule": true, "vm": true, "container": true, "application": true,
	"open_port": true, "storage": true, "volume": true, "cluster": true,
	"availability_zone": true,
}

// coreRelationTypes lists all relation types defined by the core specification.
var coreRelationTypes = map[string]bool{
	"connects": true, "hosts": true, "depends_on": true, "belongs_to": true,
	"replicates_to": true, "backs_up": true, "monitors": true, "managed_by": true,
	"mounted_on": true, "applies_to": true, "listens_on": true,
}

func isCoreEntityKind(k core.EntityKind) bool {
	return coreEntityKinds[string(k)]
}

func isCoreRelationType(t core.RelationType) bool {
	return coreRelationTypes[string(t)]
}
