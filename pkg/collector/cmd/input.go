package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/cybozu-go/setup-hw/redfish"
)

func collectOrLoad(ctx context.Context, inputFile string, rootPath string, excludes []string) (map[string]*gabs.Container, error) {
	rule := &redfish.CollectRule{
		TraverseRule: redfish.TraverseRule{
			Root:         rootPath,
			ExcludeRules: excludes,
		},
	}
	err := rule.Compile()
	if err != nil {
		return nil, err
	}

	if len(inputFile) == 0 {
		ac, uc, err := config.LoadConfig()
		if err != nil {
			return nil, err
		}

		cc := &redfish.ClientConfig{
			AddressConfig: ac,
			UserConfig:    uc,
			NoEscape:      true,
		}
		client, err := redfish.NewRedfishClient(cc)
		if err != nil {
			return nil, err
		}

		collected := client.Traverse(ctx, rule)
		return collected.Data(), nil
	}

	f, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}

	var input map[string]interface{}
	err = json.NewDecoder(f).Decode(&input)
	if err != nil {
		return nil, err
	}

	data := make(map[string]*gabs.Container)
	for k, v := range input {
		if !rule.TraverseRule.NeedTraverse(k) {
			continue
		}
		c, err := gabs.Consume(v)
		if err != nil {
			return nil, err
		}
		data[k] = c
	}
	return data, nil
}
