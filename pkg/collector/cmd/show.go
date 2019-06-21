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
	inputFile      string
	pathsOnly      bool
	requiredFields []string
	requiredRegexp *regexp.Regexp
	omitEmpty      bool
	truncateArrays bool
	ignoreFields   []string
	ignoreRegexp   *regexp.Regexp
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
		if len(showConfig.requiredFields) != 0 {
			pattern := strings.Join(showConfig.requiredFields, "|")
			r, err := regexp.Compile(pattern)
			if err != nil {
				return err
			}
			showConfig.requiredRegexp = r
		}

		well.Go(func(ctx context.Context) error {
			collected, err := collectOrLoad(ctx, showConfig.inputFile, rootConfig.baseRuleFile)
			if err != nil {
				return err
			}

			matchedRules := make(map[string]bool)
			result := make(map[string]interface{})
			for k, v := range collected.Data() {
				if showConfig.ignoreRegexp != nil {
					ignoreFields(v, showConfig.ignoreRegexp)
				}
				if showConfig.truncateArrays {
					leaveFirstItem(v)
				}
				if showConfig.omitEmpty {
					if omitEmpty(v) {
						delete(collected.Data(), k)
					}
				}
				if showConfig.pathsOnly {
					result[k] = struct{}{}
				} else {
					result[k] = v.Data()
				}

				if showConfig.requiredRegexp != nil {
					if !requiredFields(v, showConfig.requiredRegexp) {
						delete(result, k)
						continue
					}
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
		return well.Wait()
	},
}

func requiredFields(current *gabs.Container, required *regexp.Regexp) bool {
	if childrenMap, err := current.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			if required != nil && required.MatchString(k) {
				return true
			}
			if ok := requiredFields(v, required); ok {
				return true
			}
		}
		return false
	}

	if children, err := current.Children(); err == nil {
		for _, child := range children {
			if ok := requiredFields(child, required); ok {
				return true
			}
		}
	}
	return false
}

func ignoreFields(current *gabs.Container, ignorePattern *regexp.Regexp) {
	if childrenMap, err := current.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			if ignorePattern != nil && ignorePattern.MatchString(k) {
				err = current.Delete(k)
				if err != nil {
					log.Warn("failed to delete", map[string]interface{}{
						log.FnError: err,
						"key":       k,
					})
				}
				continue
			}
			ignoreFields(v, ignorePattern)
		}
		return
	}

	if children, err := current.Children(); err == nil {
		for _, child := range children {
			ignoreFields(child, ignorePattern)
		}
	}
}

func leaveFirstItem(parsed *gabs.Container) int {
	if childrenMap, err := parsed.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			count := leaveFirstItem(v)
			for i := count - 1; i > 0; i-- {
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
			count := leaveFirstItem(children[0])
			for i := count - 1; i > 0; i-- {
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

func omitEmpty(current *gabs.Container) bool {
	if childrenMap, err := current.ChildrenMap(); err == nil {
		for k, v := range childrenMap {
			if omitEmpty(v) {
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
			if omitEmpty(v) {
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
	showCmd.Flags().StringVar(&showConfig.inputFile, "input-file", "", "pre-collected Redfish data")
	showCmd.Flags().BoolVar(&showConfig.pathsOnly, "paths-only", false, "show paths only")
	showCmd.Flags().StringSliceVar(&showConfig.requiredFields, "required-field", nil, "required fields to show a page")
	showCmd.Flags().BoolVar(&showConfig.omitEmpty, "omit-empty", false, "omit empty arrays and objects")
	showCmd.Flags().BoolVar(&showConfig.truncateArrays, "truncate-arrays", false, "show first array element only")
	showCmd.Flags().StringSliceVar(&showConfig.ignoreFields, "ignore-field", nil, "ignore fields")
}
