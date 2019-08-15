package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/cybozu-go/setup-hw/redfish"
	"sigs.k8s.io/yaml"
)

func collectOrLoad(ctx context.Context, inputFile string, baseRule string) (*redfish.Collected, error) {
	var rule *redfish.CollectRule
	if len(baseRule) != 0 {
		data, err := ioutil.ReadFile(baseRule)
		if err != nil {
			return nil, err
		}
		rule = new(redfish.CollectRule)
		err = yaml.Unmarshal(data, rule)
		if err != nil {
			return nil, err
		}
	} else {
		rule = &redfish.CollectRule{
			TraverseRule: redfish.TraverseRule{
				Root: defaultRootPath,
			},
		}
	}
	if err := rule.Validate(); err != nil {
		return nil, err
	}
	if err := rule.Compile(); err != nil {
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
		return &collected, nil
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
	return redfish.NewCollected(data, rule), nil
}
