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
				if showConfig.ignoreRegexp != nil {
					ignoreFields(ctx, v, showConfig.ignoreRegexp)
				}
				if showConfig.noDup {
					leaveFirstItem(ctx, v)
				}
				if showConfig.omitempty {
					if omitEmpty(ctx, v) {
						delete(collected.Data(), k)
					}
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

func ignoreFields(ctx context.Context, current *gabs.Container, ignorePattern *regexp.Regexp) {
	if childrenMap, err := current.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			if ignorePattern != nil && ignorePattern.Match([]byte(k)) {
				err = current.Delete(k)
				if err != nil {
					log.Warn("failed to delete", map[string]interface{}{
						log.FnError: err,
						"key":       k,
					})
				}
				continue
			}
			ignoreFields(ctx, v, ignorePattern)
		}
		return
	}

	if children, err := current.Children(); err == nil {
		for _, child := range children {
			ignoreFields(ctx, child, ignorePattern)
		}
	}
}

func leaveFirstItem(ctx context.Context, parsed *gabs.Container) int {
	if childrenMap, err := parsed.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			count := leaveFirstItem(ctx, v)
			for i := count; i > 0; i-- {
				err = parsed.ArrayRemove(i, k)
				if err != nil {
					log.Warn("failed to remove", map[string]interface{}{
						log.FnError: err,
						"key":       k,
						"index":     i,
					})
				}
			}
		}
		return 0
	}

	if children, err := parsed.Children(); err == nil {
		if len(children) > 0 {
			count := leaveFirstItem(ctx, children[0])
			for i := count; i > 0; i-- {
				err = parsed.ArrayRemove(i)
				if err != nil {
					log.Warn("failed to remove", map[string]interface{}{
						log.FnError: err,
						"index":     i,
					})
				}
			}
		}
		return len(children)
	}

	return 0
}

func omitEmpty(ctx context.Context, current *gabs.Container) bool {
	if childrenMap, err := current.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			if omitEmpty(ctx, v) {
				err = current.Delete(k)
				if err != nil {
					log.Warn("failed to delete", map[string]interface{}{
						log.FnError: err,
						"key":       k,
					})
				}
			}
		}
		childrenMap, err = current.ChildrenMap()
		if err != nil {
			panic(err)
		}
		return len(childrenMap) == 0
	}
	if children, err := current.Children(); err == nil {
		for i := len(children) - 1; i >= 0; i-- {
			v := children[i]
			if omitEmpty(ctx, v) {
				err = current.ArrayRemove(i)
				if err != nil {
					log.Warn("failed to remove", map[string]interface{}{
						log.FnError: err,
						"index":     i,
					})
				}
			}
		}
		children, err = current.Children()
		if err != nil {
			panic(err)
		}
		return len(children) == 0
	}
	return false
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().BoolVar(&showConfig.keysOnly, "keys-only", false, "show keys only")
	showCmd.Flags().BoolVar(&showConfig.omitempty, "omitempty", false, "omit empty fields")
	showCmd.Flags().BoolVar(&showConfig.noDup, "no-dup", false, "remove duplicate fields")
	showCmd.Flags().StringSliceVar(&showConfig.ignoreFields, "ignore-field", nil, "ignore fields")
}
