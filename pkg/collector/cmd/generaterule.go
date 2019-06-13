package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/cybozu-go/well"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var generateRuleConfig struct {
	keys []string
}

type keyType struct {
	key string
	typ string
}

// generateRuleCmd represents the generateRule command
var generateRuleCmd = &cobra.Command{
	Use:   "generate-rule",
	Short: "output a collection rule to collect specified keys as metrics",
	Long:  `output a collection rule to collect specified keys as metrics.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		keyTypes := make([]*keyType, len(generateRuleConfig.keys))
		for i, k := range generateRuleConfig.keys {
			ks := strings.Split(k, ":")
			if len(ks) != 2 {
				return fmt.Errorf("key must be given as 'key:type': %s", k)
			}
			keyTypes[i] = &keyType{
				key: ks[0],
				typ: ks[1],
			}
		}

		well.Go(func(ctx context.Context) error {
			data, err := collectOrLoad(ctx, rootConfig.inputFile, rootConfig.rootPath, rootConfig.excludes)
			if err != nil {
				return err
			}

			rules := generateRule(data, keyTypes, rootConfig.rootPath)
			collectRule := &redfish.CollectRule{
				TraverseRule: redfish.TraverseRule{
					Root:         rootConfig.rootPath,
					ExcludeRules: rootConfig.excludes,
				},
				MetricRules: rules,
			}

			out, err := yaml.Marshal(collectRule)
			if err != nil {
				return err
			}
			_, err = os.Stdout.Write(out)
			if err != nil {
				return err
			}
			return nil
		})

		well.Stop()
		err := well.Wait()
		if err != nil {
			return err
		}
		return nil
	},
}

func generateRule(data map[string]*gabs.Container, keyTypes []*keyType, rootPath string) []*redfish.MetricRule {
	var rules []*redfish.MetricRule

	for path, parsedJSON := range data {
		prefix := normalize(strings.ReplaceAll(path[len(rootPath):], "/", "_"))
		propertyRules := generateRuleAux(parsedJSON, keyTypes, "", prefix, []*redfish.PropertyRule{})

		if len(propertyRules) > 0 {
			sort.Slice(propertyRules, func(i, j int) bool { return propertyRules[i].Pointer < propertyRules[j].Pointer })
			rules = append(rules, &redfish.MetricRule{
				Path:          path,
				PropertyRules: propertyRules,
			})
		}
	}

	sort.Slice(rules, func(i, j int) bool { return rules[i].Path < rules[j].Path })
	return rules
}

func generateRuleAux(data *gabs.Container, keyTypes []*keyType, pointer, name string, rules []*redfish.PropertyRule) []*redfish.PropertyRule {
	if childrenMap, err := data.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			kt := findKeyInKeyTypes(k, keyTypes)
			if kt != nil {
				rules = append(rules, &redfish.PropertyRule{
					Pointer: pointer + "/" + k,
					Name:    (name + "_" + normalize(k))[1:], // trim the first "_"
					Type:    kt.typ,
				})
			} else {
				rules = generateRuleAux(v, keyTypes, pointer+"/"+k, name+"_"+normalize(k), rules)
			}
		}
		return rules
	}

	if childrenSlice, err := data.Children(); err == nil {
		for _, v := range childrenSlice {
			rules = generateRuleAux(v, keyTypes, pointer+"/{TBD}", name, rules)
			break // inspect the first element only; the rest should have the same structure
		}
		return rules
	}

	return rules
}

func findKeyInKeyTypes(key string, keyTypes []*keyType) *keyType {
	for _, kt := range keyTypes {
		if kt.key == key {
			return kt
		}
	}
	return nil
}

// Convert a key into a mostly-suitable name for the Prometheus metric name '[a-zA-Z_:][a-zA-Z0-9_:]*'.
// This may return a non-suitable name like "" or "1ab".  Use the returned name just for hint.
func normalize(key string) string {
	// drop runes not in '[a-zA-Z0-9_]', with lowering
	// ':' is also dropped
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return 'a' + (r - 'A')
		}
		return -1
	}, key)
}

func init() {
	rootCmd.AddCommand(generateRuleCmd)

	generateRuleCmd.Flags().StringSliceVar(&generateRuleConfig.keys, "key", nil, "Redfish data key to find")
}
