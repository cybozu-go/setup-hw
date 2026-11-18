package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"sigs.k8s.io/yaml"
)

func TestGenerateRule(t *testing.T) {
	powerPath := "/redfish/v1/Chassis/System.Embedded.1/Power"
	powerPatternedPath := "/redfish/v1/Chassis/{chassis}/Power"
	powerJSON := `
{
    "@odata.context": "/redfish/v1/$metadata#Power.Power",
    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power",
    "@odata.type": "#Power.v1_5_0.Power",
    "Description": "Power",
    "Id": "Power",
    "Name": "Power",
    "PowerSupplies": [
        {
            "@odata.context": "/redfish/v1/$metadata#Power.Power",
            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.1",
            "@odata.type": "#Power.v1_5_0.PowerSupply",
            "LineInputVoltage": 104,
            "Redundancy": [
                {
                    "@odata.context": "/redfish/v1/$metadata#Redundancy.Redundancy",
                    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/Redundancy/iDRAC.Embedded.1%23SystemBoardPSRedundancy",
                    "@odata.type": "#Redundancy.v1_3_0.Redundancy",
                    "Status": {
                        "Health": "OK",
                        "State": "Enabled"
                    }
                }
            ],
            "Status": {
                "Health": "OK",
                "State": "Enabled"
            }
        },
        {
            "@odata.context": "/redfish/v1/$metadata#Power.Power",
            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.2",
            "@odata.type": "#Power.v1_5_0.PowerSupply",
            "LineInputVoltage": 104,
            "Redundancy": [
                {
                    "@odata.context": "/redfish/v1/$metadata#Redundancy.Redundancy",
                    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/Redundancy/iDRAC.Embedded.1%23SystemBoardPSRedundancy",
                    "@odata.type": "#Redundancy.v1_3_0.Redundancy",
                    "Status": {
                        "Health": "OK",
                        "State": "Enabled"
                    }
                }
            ],
            "Status": {
                "Health": "OK",
                "State": "Enabled"
            }
        }
    ]
}
`
	powerParsedJSON, err := gabs.ParseJSON([]byte(powerJSON))
	if err != nil {
		t.Fatal(err)
	}

	input := map[string]*gabs.Container{
		powerPath: powerParsedJSON,
	}

	expected := []*redfish.MetricRule{
		{
			Path: powerPatternedPath,
			PropertyRules: []*redfish.PropertyRule{
				{
					Pointer: "/PowerSupplies/{powersupply}/LineInputVoltage",
					Name:    "chassis_power_powersupplies_lineinputvoltage",
					Type:    "number",
				},
				{
					Pointer: "/PowerSupplies/{powersupply}/Redundancy/{redundancy}/Status/Health",
					Name:    "chassis_power_powersupplies_redundancy_status_health",
					Type:    "health",
				},
				{
					Pointer: "/PowerSupplies/{powersupply}/Status/Health",
					Name:    "chassis_power_powersupplies_status_health",
					Type:    "health",
				},
			},
		},
	}

	// collector generate-rule --key=Health:health --key=LineInputVoltage:number
	result := generateRule(input, []*keyType{
		{
			key: "Health",
			typ: "health",
		},
		{
			key: "LineInputVoltage",
			typ: "number",
		},
	}, &redfish.CollectRule{
		TraverseRule: redfish.TraverseRule{
			Root: defaultRootPath,
		},
		MetricRules: []*redfish.MetricRule{
			{
				Path: powerPatternedPath,
			},
		},
	})

	opts := cmpopts.IgnoreUnexported(redfish.TraverseRule{}, redfish.PropertyRule{})
	if !cmp.Equal(result, expected, opts) {
		t.Error("generateRule() returned unexpected result:", cmp.Diff(expected, result, opts))
	}
}

func TestGenerateRuleMergesPages(t *testing.T) {
	t.Parallel()

	// Pages matching the same patterned path, and elements in the same array,
	// have different properties.  The first page has no properties at all.
	pages := map[string]string{
		"/redfish/v1/Chassis/System.Embedded.1/Sensors/Sensor0": `{"Name": "no metrics"}`,
		"/redfish/v1/Chassis/System.Embedded.1/Sensors/Sensor1": `{
			"Reading": 30,
			"Thresholds": {"UpperCritical": {"Reading": 90}}
		}`,
		"/redfish/v1/Chassis/System.Embedded.1/Sensors/Sensor2": `{
			"Reading": 40,
			"Status": {"Health": "OK"},
			"Thresholds": {"LowerCritical": {"Reading": 5}}
		}`,
		"/redfish/v1/Chassis/System.Embedded.1/Power": `{
			"PowerSupplies": [
				{"Name": "no metrics"},
				{"Status": {"Health": "OK"}},
				{"Status": {"State": "Enabled"}}
			]
		}`,
	}
	input := make(map[string]*gabs.Container)
	for path, j := range pages {
		parsed, err := gabs.ParseJSON([]byte(j))
		if err != nil {
			t.Fatal(err)
		}
		input[path] = parsed
	}

	expected := []*redfish.MetricRule{
		{
			Path: "/redfish/v1/Chassis/{chassis}/Power",
			PropertyRules: []*redfish.PropertyRule{
				{Pointer: "/PowerSupplies/{powersupply}/Status/Health", Name: "chassis_power_powersupplies_status_health", Type: "health"},
				{Pointer: "/PowerSupplies/{powersupply}/Status/State", Name: "chassis_power_powersupplies_status_state", Type: "state"},
			},
		},
		{
			Path: "/redfish/v1/Chassis/{chassis}/Sensors/{sensor}",
			PropertyRules: []*redfish.PropertyRule{
				{Pointer: "/Reading", Name: "chassis_sensors_reading", Type: "number"},
				{Pointer: "/Status/Health", Name: "chassis_sensors_status_health", Type: "health"},
				{Pointer: "/Thresholds/LowerCritical/Reading", Name: "chassis_sensors_thresholds_lowercritical_reading", Type: "number"},
				{Pointer: "/Thresholds/UpperCritical/Reading", Name: "chassis_sensors_thresholds_uppercritical_reading", Type: "number"},
			},
		},
	}

	keyTypes := []*keyType{
		{key: "Health", typ: "health"},
		{key: "Reading", typ: "number"},
		{key: "State", typ: "state"},
	}
	baseRule := &redfish.CollectRule{
		TraverseRule: redfish.TraverseRule{
			Root: defaultRootPath,
		},
		MetricRules: []*redfish.MetricRule{
			{Path: "/redfish/v1/Chassis/{chassis}/Power"},
			{Path: "/redfish/v1/Chassis/{chassis}/Sensors/{sensor}"},
		},
	}

	opts := cmpopts.IgnoreUnexported(redfish.TraverseRule{}, redfish.PropertyRule{})
	// Go randomizes the iteration order of maps, so try several times.
	for i := 0; i < 10; i++ {
		result := generateRule(input, keyTypes, baseRule)
		if !cmp.Equal(result, expected, opts) {
			t.Fatal("generateRule() returned unexpected result:", cmp.Diff(expected, result, opts))
		}
	}
}

func TestMergeCollectRules(t *testing.T) {
	input1 := &redfish.CollectRule{
		TraverseRule: redfish.TraverseRule{
			Root: "/redfish/v1",
		},
		MetricRules: []*redfish.MetricRule{
			{
				Path: "/redfish/v1/CommonPage",
				PropertyRules: []*redfish.PropertyRule{
					{
						Pointer: "/CommonProperty",
						Name:    "commonpage_commonproperty",
						Type:    "number",
					},
					{
						Pointer: "/PropertyC1",
						Name:    "commonpage_propertyc1",
						Type:    "number",
					},
				},
			},
			{
				Path: "/redfish/v1/Page1",
				PropertyRules: []*redfish.PropertyRule{
					{
						Pointer: "/SimilarProperty",
						Name:    "page1_similarproperty",
						Type:    "number",
					},
				},
			},
		},
	}
	input2 := &redfish.CollectRule{
		TraverseRule: redfish.TraverseRule{
			Root: "/redfish/v1",
		},
		MetricRules: []*redfish.MetricRule{
			{
				Path: "/redfish/v1/CommonPage",
				PropertyRules: []*redfish.PropertyRule{
					{
						Pointer: "/CommonProperty",
						Name:    "commonpage_commonproperty",
						Type:    "number",
					},
					{
						Pointer: "/PropertyC2",
						Name:    "commonpage_propertyc2",
						Type:    "number",
					},
				},
			},
			{
				Path: "/redfish/v1/Page2",
				PropertyRules: []*redfish.PropertyRule{
					{
						Pointer: "/SimilarProperty",
						Name:    "page2_similarproperty",
						Type:    "number",
					},
				},
			},
		},
	}

	expected := &redfish.CollectRule{
		TraverseRule: redfish.TraverseRule{
			Root: "/redfish/v1",
		},
		MetricRules: []*redfish.MetricRule{
			{
				Path: "/redfish/v1/CommonPage",
				PropertyRules: []*redfish.PropertyRule{
					{
						Pointer: "/CommonProperty",
						Name:    "commonpage_commonproperty",
						Type:    "number",
					},
					{
						Pointer: "/PropertyC1",
						Name:    "commonpage_propertyc1",
						Type:    "number",
					},
					{
						Pointer: "/PropertyC2",
						Name:    "commonpage_propertyc2",
						Type:    "number",
					},
				},
			},
			{
				Path: "/redfish/v1/Page1",
				PropertyRules: []*redfish.PropertyRule{
					{
						Pointer: "/SimilarProperty",
						Name:    "page1_similarproperty",
						Type:    "number",
					},
				},
			},
			{
				Path: "/redfish/v1/Page2",
				PropertyRules: []*redfish.PropertyRule{
					{
						Pointer: "/SimilarProperty",
						Name:    "page2_similarproperty",
						Type:    "number",
					},
				},
			},
		},
	}

	merged := mergeCollectRules([]*redfish.CollectRule{input1, input2})
	opts := cmpopts.IgnoreUnexported(redfish.TraverseRule{}, redfish.PropertyRule{})
	if !cmp.Equal(merged, expected, opts) {
		t.Error("mergeCollectRules() returned unexpected result:", cmp.Diff(expected, merged, opts))
	}
}

func TestRenderRule(t *testing.T) {
	t.Parallel()

	rule := &redfish.CollectRule{
		TraverseRule: redfish.TraverseRule{
			Root:         "/redfish/v1",
			ExcludeRules: []string{"/JsonSchemas", "/redfish/v1/Chassis/$"},
		},
		MetricRules: []*redfish.MetricRule{
			{
				Path: "/redfish/v1/Chassis/{chassis}",
				PropertyRules: []*redfish.PropertyRule{
					{Pointer: "/Status/Health", Name: "chassis_status_health", Type: "health"},
				},
			},
			{
				Path: "/redfish/v1/Systems/{system}",
				PropertyRules: []*redfish.PropertyRule{
					{Pointer: "/Status/Health", Name: "systems_status_health", Type: "health"},
					{Pointer: "/Status/State", Name: "systems_status_state", Type: "state"},
				},
			},
		},
	}
	expected := `Traverse:
  Root: /redfish/v1
  Excludes:
  - /JsonSchemas
  - /redfish/v1/Chassis/$

Metrics:
- Path: /redfish/v1/Chassis/{chassis}
  Properties:
  - Name: chassis_status_health
    Pointer: /Status/Health
    Type: health

- Path: /redfish/v1/Systems/{system}
  Properties:
  - Name: systems_status_health
    Pointer: /Status/Health
    Type: health
  - Name: systems_status_state
    Pointer: /Status/State
    Type: state
`
	out, err := renderRule(rule)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(expected, string(out)); diff != "" {
		t.Errorf("renderRule() mismatch (-want +got):\n%s", diff)
	}

	var parsed redfish.CollectRule
	if err := yaml.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	opts := cmpopts.IgnoreUnexported(redfish.TraverseRule{}, redfish.PropertyRule{})
	if diff := cmp.Diff(rule, &parsed, opts); diff != "" {
		t.Errorf("rendered rule is not parsed back to the same rule (-want +got):\n%s", diff)
	}
}

func TestRenderRuleFiles(t *testing.T) {
	t.Parallel()

	// These rule files are in the layout that renderRule produces,
	// except for comments added by hand, which are ignored in comparison.
	files := []string{
		"dell_redfish_1.20.1.yml",
	}
	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join("..", "..", "..", "redfish", "rules", name))
			if err != nil {
				t.Fatal(err)
			}
			var rule redfish.CollectRule
			if err := yaml.Unmarshal(data, &rule); err != nil {
				t.Fatal(err)
			}
			out, err := renderRule(&rule)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(removeComments(string(data)), string(out)); diff != "" {
				t.Errorf("renderRule() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func removeComments(s string) string {
	var lines []string
	for _, line := range strings.SplitAfter(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "")
}
