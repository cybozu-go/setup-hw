package cmd

import (
	"testing"

	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
