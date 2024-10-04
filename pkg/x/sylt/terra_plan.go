package sylt

import (
	tfjson "github.com/hashicorp/terraform-json"
)

type plan struct {
	out *tfjson.Plan
	// isApplied is true if terra apply has been run after the plan was created.
	// It should be reset if plan is called again.
	isApplied bool
}

// diff returns true if the plan has proposed changes.
// If a plan has been applied, we assume that the plan is in sync with the
// state (i.e. whatever diff existed has been applied, hence no diff).
func (p *plan) diff() bool {
	if p.isApplied {
		return false
	}
	for _, res := range p.out.ResourceChanges {
		for _, action := range res.Change.Actions {
			switch action {
			case tfjson.ActionCreate, tfjson.ActionDelete, tfjson.ActionUpdate:
				return true
			default:
				continue
			}
		}
	}
	return false
}
