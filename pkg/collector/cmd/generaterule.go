package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
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
	Use:   "generate-rule FILE...",
	Short: "output a collection rule to collect specified keys as metrics",
	Long: `Output a collection rule to collect specified keys as metrics.

It takes one or more JSON file names that were dumped by "collector show" command.`,
	Args: cobra.MinimumNArgs(1),

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
			rules := make([]*redfish.CollectRule, len(args))
			for i, fname := range args {
				collected, err := collectOrLoad(ctx, fname, rootConfig.baseRuleFile)
				if err != nil {
					return err
				}

				metricRules := generateRule(collected.Data(), keyTypes, collected.Rule())
				collectRule := &redfish.CollectRule{
					TraverseRule: collected.Rule().TraverseRule,
					MetricRules:  metricRules,
				}
				rules[i] = collectRule
			}

			mergedRule := mergeCollectRules(rules)

			out, err := yaml.Marshal(mergedRule)
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
		return well.Wait()
	},
}

func generateRule(data map[string]*gabs.Container, keyTypes []*keyType, rule *redfish.CollectRule) []*redfish.MetricRule {
	var rules []*redfish.MetricRule

	matchedRules := make(map[string]bool)
OUTER:
	for path, parsedJSON := range data {
		if rule != nil {
			for _, r := range rule.MetricRules {
				matched, _ := r.MatchPath(path)
				if matched {
					if matchedRules[r.Path] {
						continue OUTER
					}
					matchedRules[r.Path] = true
					path = r.Path
					break
				}
			}
		}
		prefix := normalize(strings.ReplaceAll(path[len(rule.TraverseRule.Root):], "/", "_"))
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
			ps := strings.Split(pointer, "/")
			parent := ps[len(ps)-1]
			if strings.HasSuffix(parent, "ies") {
				parent = regexp.MustCompile("ies$").ReplaceAllString(parent, "y")
			} else if strings.HasSuffix(parent, "s") {
				parent = regexp.MustCompile("s$").ReplaceAllString(parent, "")
			}
			rules = generateRuleAux(v, keyTypes, pointer+"/{"+strings.ToLower(parent)+"}", name, rules)
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

func mergeCollectRules(collectRules []*redfish.CollectRule) *redfish.CollectRule {
	if len(collectRules) == 0 {
		return &redfish.CollectRule{}
	}

	// This function merges CollectRules assuming that:
	//   1. all rules are generated based on the same rule file, and
	//   2. all rules are generated for the same set of keys and types.
	// TraverseRules are exactly the same among all CollectRules from assumption 1.
	// So we concentrate on merging MetricRules.
	metricRules := []*redfish.MetricRule{}
	for _, cr := range collectRules {
		for _, mr := range cr.MetricRules {
			metricRules = appendMetricRule(metricRules, mr)
		}
	}

	return &redfish.CollectRule{
		TraverseRule: collectRules[0].TraverseRule,
		MetricRules:  metricRules,
	}
}

func appendMetricRule(rules []*redfish.MetricRule, newRule *redfish.MetricRule) []*redfish.MetricRule {
	for _, rule := range rules {
		// We can simply compare paths without considering patterns from assumption 1.
		if newRule.Path == rule.Path {
			for _, pr := range newRule.PropertyRules {
				rule.PropertyRules = appendPropertyRule(rule.PropertyRules, pr)
			}
			return rules
		}
	}

	return append(rules, newRule)
}

func appendPropertyRule(rules []*redfish.PropertyRule, newRule *redfish.PropertyRule) []*redfish.PropertyRule {
	for _, rule := range rules {
		if newRule.Pointer == rule.Pointer {
			// We don't need to compare types from assumption 2.
			return rules
		}
	}

	return append(rules, newRule)
}

func init() {
	rootCmd.AddCommand(generateRuleCmd)

	generateRuleCmd.Flags().StringSliceVar(&generateRuleConfig.keys, "key", nil, "Redfish data key to find")
}
