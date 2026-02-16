package common

import "github.com/ctfer-io/chall-manager/sdk"

// Variable represent a content that can be variated.
type Variable struct {
	// The content to set.
	Content string `form:"content" json:"content"`

	// Whether to variate the content according per a PRNG seeded by the instance's identity (reproducible).
	Variate bool `form:"variate" json:"variate"`

	// Variation functional options

	Lowercase *bool `form:"lowercase" json:"lowercase,omitempty"`
	Uppercase *bool `form:"uppercase" json:"uppercase,omitempty"`
	Numeric   *bool `form:"numeric"   json:"numeric,omitempty"`
	Special   *bool `form:"special"   json:"special,omitempty"`
}

// Produce the content given its configuration, and a seed (should be the instance identity
// for proper reproducibility).
func (v Variable) Produce(seed string) string {
	if !v.Variate {
		return v.Content
	}

	return sdk.Variate(seed, v.Content,
		sdk.WithLowercase(v.Lowercase == nil || *v.Lowercase),
		sdk.WithUppercase(v.Uppercase == nil || *v.Uppercase),
		sdk.WithNumeric(v.Numeric == nil || *v.Numeric),
		sdk.WithSpecial(v.Special != nil && *v.Special),
	)
}
