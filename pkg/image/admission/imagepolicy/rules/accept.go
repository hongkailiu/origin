package rules

import (
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/openshift/origin/pkg/image/admission/imagepolicy/api"
)

type Accepter interface {
	Covers(schema.GroupResource) bool

	Accepts(*ImagePolicyAttributes) bool
}

// mappedAccepter implements the Accepter interface for a map of group resources and accepters
type mappedAccepter map[schema.GroupResource]Accepter

func (a mappedAccepter) Covers(gr schema.GroupResource) bool {
	_, ok := a[gr]
	return ok
}

// Accepts returns true if no Accepter is registered for the group resource in attributes,
// or if the registered Accepter also returns true.
func (a mappedAccepter) Accepts(attr *ImagePolicyAttributes) bool {
	accepter, ok := a[attr.Resource]
	if !ok {
		return true
	}
	return accepter.Accepts(attr)
}

type executionAccepter struct {
	rules         []api.ImageExecutionPolicyRule
	covers        schema.GroupResource
	defaultReject bool

	integratedRegistryMatcher RegistryMatcher
}

// NewExecutionRuleseAccepter creates an Accepter from the provided rules.
func NewExecutionRulesAccepter(rules []api.ImageExecutionPolicyRule, integratedRegistryMatcher RegistryMatcher) (Accepter, error) {
	mapped := make(mappedAccepter)

	for _, rule := range rules {
		over, selectors, err := imageConditionInfo(&rule.ImageCondition)
		if err != nil {
			return nil, err
		}
		rule.ImageCondition.MatchImageLabelSelectors = selectors
		for gr := range over {
			a, ok := mapped[gr]
			if !ok {
				a = &executionAccepter{
					covers: gr,
					integratedRegistryMatcher: integratedRegistryMatcher,
				}
				mapped[gr] = a
			}
			byResource := a.(*executionAccepter)
			byResource.rules = append(byResource.rules, rule)
		}
	}

	for _, a := range mapped {
		byResource := a.(*executionAccepter)
		if len(byResource.rules) > 0 {
			// if all rules are reject, the default behavior is allow
			allReject := true
			for _, rule := range byResource.rules {
				if !rule.Reject {
					allReject = false
					break
				}
			}
			byResource.defaultReject = !allReject
		}
	}

	return mapped, nil
}

func (r *executionAccepter) Covers(gr schema.GroupResource) bool {
	return r.covers == gr
}

func (r *executionAccepter) Accepts(attrs *ImagePolicyAttributes) bool {
	if attrs.Resource != r.covers {
		return true
	}

	anyMatched := false
	for _, rule := range r.rules {
		glog.V(5).Infof("image policy checking rule %q", rule.Name)
		if attrs.ExcludedRules.Has(rule.Name) && !rule.IgnoreNamespaceOverride {
			glog.V(5).Infof("skipping because rule is excluded by namespace annotations\n")
			continue
		}

		// if we don't have a resolved image and we're supposed to skip the rule if that happens,
		// continue here.  Otherwise, the reject option is impossible to reason about.
		if attrs.Image == nil && rule.SkipOnResolutionFailure {
			glog.V(5).Infof("skipping because image is not resolved and skip on failure is true\n")
			continue
		}

		matches := matchImageCondition(&rule.ImageCondition, r.integratedRegistryMatcher, attrs)
		glog.V(5).Infof("Rule %q(reject=%t) applies to image %v: %t", rule.Name, rule.Reject, attrs.Name, matches)
		if matches {
			if rule.Reject {
				return false
			}
			anyMatched = true
		}
	}
	return anyMatched || !r.defaultReject
}
