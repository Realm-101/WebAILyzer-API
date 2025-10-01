package insights

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
)

// RuleRegistry manages and executes insight rules
type RuleRegistry struct {
	rules map[string]InsightRule
	mutex sync.RWMutex
}

// NewRuleRegistry creates a new rule registry
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		rules: make(map[string]InsightRule),
	}
}

// RegisterRule adds a new rule to the registry
func (r *RuleRegistry) RegisterRule(rule InsightRule) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	name := rule.Name()
	if name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}
	
	if _, exists := r.rules[name]; exists {
		return fmt.Errorf("rule with name '%s' already exists", name)
	}
	
	r.rules[name] = rule
	return nil
}

// UnregisterRule removes a rule from the registry
func (r *RuleRegistry) UnregisterRule(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.rules[name]; !exists {
		return fmt.Errorf("rule with name '%s' not found", name)
	}
	
	delete(r.rules, name)
	return nil
}

// GetRule retrieves a specific rule by name
func (r *RuleRegistry) GetRule(name string) (InsightRule, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	rule, exists := r.rules[name]
	if !exists {
		return nil, fmt.Errorf("rule with name '%s' not found", name)
	}
	
	return rule, nil
}

// ListRules returns all registered rules
func (r *RuleRegistry) ListRules() []InsightRule {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	rules := make([]InsightRule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}
	
	// Sort rules by priority (critical first, then high, medium, low)
	sort.Slice(rules, func(i, j int) bool {
		return getPriorityWeight(rules[i].Priority()) > getPriorityWeight(rules[j].Priority())
	})
	
	return rules
}

// ExecuteRules runs all registered rules against the provided data
func (r *RuleRegistry) ExecuteRules(ctx context.Context, data *AnalysisData) ([]*models.Insight, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var allInsights []*models.Insight
	var errors []error
	
	// Execute rules in priority order
	rules := r.ListRules()
	
	for _, rule := range rules {
		select {
		case <-ctx.Done():
			return allInsights, ctx.Err()
		default:
			insights, err := rule.Evaluate(ctx, data)
			if err != nil {
				errors = append(errors, fmt.Errorf("rule '%s' failed: %w", rule.Name(), err))
				continue
			}
			
			allInsights = append(allInsights, insights...)
		}
	}
	
	// Return insights even if some rules failed
	if len(errors) > 0 && len(allInsights) == 0 {
		return nil, fmt.Errorf("all rules failed: %v", errors)
	}
	
	// Deduplicate and prioritize insights
	return r.deduplicateInsights(allInsights), nil
}

// ExecuteRulesByType runs only rules that generate insights of the specified type
func (r *RuleRegistry) ExecuteRulesByType(ctx context.Context, data *AnalysisData, insightType models.InsightType) ([]*models.Insight, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var allInsights []*models.Insight
	var errors []error
	
	for _, rule := range r.rules {
		if rule.Type() != insightType {
			continue
		}
		
		select {
		case <-ctx.Done():
			return allInsights, ctx.Err()
		default:
			insights, err := rule.Evaluate(ctx, data)
			if err != nil {
				errors = append(errors, fmt.Errorf("rule '%s' failed: %w", rule.Name(), err))
				continue
			}
			
			allInsights = append(allInsights, insights...)
		}
	}
	
	if len(errors) > 0 && len(allInsights) == 0 {
		return nil, fmt.Errorf("all rules of type '%s' failed: %v", insightType, errors)
	}
	
	return r.deduplicateInsights(allInsights), nil
}

// GetRuleCount returns the number of registered rules
func (r *RuleRegistry) GetRuleCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return len(r.rules)
}

// GetRulesByType returns all rules that generate insights of the specified type
func (r *RuleRegistry) GetRulesByType(insightType models.InsightType) []InsightRule {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var rules []InsightRule
	for _, rule := range r.rules {
		if rule.Type() == insightType {
			rules = append(rules, rule)
		}
	}
	
	return rules
}

// deduplicateInsights removes duplicate insights based on title and type
func (r *RuleRegistry) deduplicateInsights(insights []*models.Insight) []*models.Insight {
	seen := make(map[string]bool)
	var unique []*models.Insight
	
	for _, insight := range insights {
		key := fmt.Sprintf("%s:%s", insight.InsightType, insight.Title)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, insight)
		}
	}
	
	// Sort by priority and impact score
	sort.Slice(unique, func(i, j int) bool {
		iPriority := getPriorityWeight(unique[i].Priority)
		jPriority := getPriorityWeight(unique[j].Priority)
		
		if iPriority != jPriority {
			return iPriority > jPriority
		}
		
		// If priorities are equal, sort by impact score
		iImpact := 0
		jImpact := 0
		
		if unique[i].ImpactScore != nil {
			iImpact = *unique[i].ImpactScore
		}
		if unique[j].ImpactScore != nil {
			jImpact = *unique[j].ImpactScore
		}
		
		return iImpact > jImpact
	})
	
	return unique
}

// getPriorityWeight returns a numeric weight for priority comparison
func getPriorityWeight(priority models.Priority) int {
	switch priority {
	case models.PriorityCritical:
		return 4
	case models.PriorityHigh:
		return 3
	case models.PriorityMedium:
		return 2
	case models.PriorityLow:
		return 1
	default:
		return 0
	}
}