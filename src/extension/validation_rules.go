package extension

import (
	"fmt"

	"IACForge/src/validation"
)

// ValidationRulesExtensionPoint manages validation rule extensions.
type ValidationRulesExtensionPoint struct {
	engine *validation.Engine
	rules  map[string]string // rule ID -> extension ID that registered it
}

// NewValidationRulesExtensionPoint creates a new validation rules extension point.
func NewValidationRulesExtensionPoint(engine *validation.Engine) *ValidationRulesExtensionPoint {
	return &ValidationRulesExtensionPoint{
		engine: engine,
		rules:  make(map[string]string),
	}
}

// Type returns the extension point type.
func (ep *ValidationRulesExtensionPoint) Type() ExtensionPointType {
	return ExtensionPointValidationRules
}

// Register registers all validation rule contributions from the given extension.
func (ep *ValidationRulesExtensionPoint) Register(ext *Extension) error {
	for _, contrib := range ext.ValidationRules {
		if contrib.Rule == nil || contrib.Fn == nil {
			return fmt.Errorf("%w: validation rule contribution has nil rule or function", ErrInvalidExtension)
		}
		ep.engine.RegisterRule(contrib.Rule, contrib.Fn)
		ep.rules[contrib.Rule.ID] = ext.Manifest.ID
	}
	return nil
}

// GetRulesByExtension returns all rule IDs registered by a specific extension.
func (ep *ValidationRulesExtensionPoint) GetRulesByExtension(extensionID string) []string {
	var result []string
	for ruleID, extID := range ep.rules {
		if extID == extensionID {
			result = append(result, ruleID)
		}
	}
	return result
}

// GetExtensionForRule returns the extension ID that registered the given rule.
func (ep *ValidationRulesExtensionPoint) GetExtensionForRule(ruleID string) (string, bool) {
	extID, ok := ep.rules[ruleID]
	return extID, ok
}
