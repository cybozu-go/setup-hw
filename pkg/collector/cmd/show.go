package cmd

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var showConfig struct {
	keysOnly     bool
	omitempty    bool
	noDup        bool
	ignoreFields []string
	ignoreRegexp *regexp.Regexp
}

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show collected Redfish data",
	Long:  `show collected Redfish data`,

	RunE: func(cmd *cobra.Command, args []string) error {

		if len(showConfig.ignoreFields) != 0 {
			pattern := strings.Join(showConfig.ignoreFields, "|")
			r, err := regexp.Compile(pattern)
			if err != nil {
				return err
			}
			showConfig.ignoreRegexp = r
		}

		well.Go(func(ctx context.Context) error {
			collected, err := collectOrLoad(ctx, rootConfig.inputFile, rootConfig.baseRules)
			if err != nil {
				return err
			}

			matchedRules := make(map[string]bool)
			result := make(map[string]interface{})
			for k, v := range collected.Data() {
				if showConfig.ignoreRegexp != nil || showConfig.noDup {
					pickup(ctx, v, showConfig.ignoreRegexp, showConfig.noDup)
				}
				if showConfig.omitempty {
					omitEmpty(ctx, v)
				}
				if showConfig.keysOnly {
					result[k] = struct{}{}
				} else {
					result[k] = v.Data()
				}

				if collected.Rule() != nil {
					for _, r := range collected.Rule().MetricRules {
						matched, _ := r.MatchPath(k)
						if ok := matchedRules[r.Path]; matched && ok {
							delete(result, k)
						}
						if matched {
							matchedRules[r.Path] = true
						}
					}
				}
			}

			out, err := json.MarshalIndent(result, "", "    ")
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

func pickup(ctx context.Context, parsed *gabs.Container, ignorePattern *regexp.Regexp, noDup bool) {
	if childrenMap, err := parsed.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			if ignorePattern != nil && ignorePattern.Match([]byte(k)) {
				err = parsed.Delete(k)
				if err != nil {
					log.Warn("failed to delete", map[string]interface{}{
						log.FnError: err,
						"key":       k,
					})
				}
				continue
			}
			if noDup {
				if childrenSlice, err := v.Children(); err == nil {
					if len(childrenSlice) > 0 {
						pickup(ctx, childrenSlice[0], ignorePattern, noDup)
					}
					for i := len(childrenSlice) - 1; i > 0; i-- {
						err := parsed.ArrayRemove(i, k)
						if err != nil {
							log.Warn("failed to remove", map[string]interface{}{
								log.FnError: err,
								"key":       k,
								"index":     i,
							})
						}
					}
					continue
				}
			}
			pickup(ctx, v, ignorePattern, noDup)
		}
		return
	}
}

func omitEmpty(ctx context.Context, parsed *gabs.Container) {
	if childrenMap, err := parsed.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			omitEmpty(ctx, v)
			if chMap, err := v.ChildrenMap(); err == nil {
				if len(chMap) == 0 {
					err = parsed.Delete(k)
					if err != nil {
						log.Warn("failed to delete", map[string]interface{}{
							log.FnError: err,
							"key":       k,
						})
					}
					continue
				}
			}
			if chArray, err := v.Children(); err == nil {
				if len(chArray) == 0 {
					err = parsed.Delete(k)
					if err != nil {
						log.Warn("failed to delete", map[string]interface{}{
							log.FnError: err,
							"key":       k,
						})
					}
					continue
				}
			}
		}
	}
	if child, err := parsed.Children(); err == nil {
		for i := len(child) - 1; i >= 0; i-- {
			v := child[i]
			omitEmpty(ctx, v)
			if chMap, err := v.ChildrenMap(); err == nil {
				if len(chMap) == 0 {
					err = parsed.ArrayRemove(i)
					if err != nil {
						log.Warn("failed to remove", map[string]interface{}{
							log.FnError: err,
							"index":     i,
						})
					}
					continue
				}
			}
			if chArray, err := v.Children(); err == nil {
				if len(chArray) == 0 {
					err = parsed.ArrayRemove(i)
					if err != nil {
						log.Warn("failed to remove", map[string]interface{}{
							log.FnError: err,
							"index":     i,
						})
					}
					continue
				}
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().BoolVar(&showConfig.keysOnly, "keys-only", false, "show keys only")
	showCmd.Flags().BoolVar(&showConfig.omitempty, "omitempty", false, "omit empty fields")
	showCmd.Flags().BoolVar(&showConfig.noDup, "no-dup", false, "remove duplicate fields")
	showCmd.Flags().StringSliceVar(&showConfig.ignoreFields, "ignore-field", nil, "ignore fields")
}
