package main

import (
	"encoding/json"

	"github.com/ctfer-io/chall-manager/sdk"
	"github.com/ctfer-io/recipes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// This scenario is made to debug automation.
//
// It returns the input configuration in the connection info such that the concerns
// boundaries could be crossed, thus ease debug.

type Config map[string]string

func main() {
	recipes.Run(func(req *recipes.Request[Config], resp *sdk.Response, opts ...pulumi.ResourceOption) error {
		conf, err := json.Marshal(req.Config)
		if err != nil {
			return err
		}
		resp.ConnectionInfo = pulumi.Sprintf("Configuration: %s", conf)
		return nil
	})
}
