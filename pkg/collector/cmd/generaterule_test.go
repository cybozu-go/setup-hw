package cmd

import (
	"reflect"
	"testing"

	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/ghodss/yaml"
)

func TestGenerateRule(t *testing.T) {
	powerPath := "/redfish/v1/Chassis/System.Embedded.1/Power"
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
			Path: powerPath,
			PropertyRules: []*redfish.PropertyRule{
				{
					Pointer: "/PowerSupplies/{TBD}/LineInputVoltage",
					Name:    "chassis_systemembedded1_power_powersupplies_lineinputvoltage",
					Type:    "number",
				},
				{
					Pointer: "/PowerSupplies/{TBD}/Redundancy/{TBD}/Status/Health",
					Name:    "chassis_systemembedded1_power_powersupplies_redundancy_status_health",
					Type:    "health",
				},
				{
					Pointer: "/PowerSupplies/{TBD}/Status/Health",
					Name:    "chassis_systemembedded1_power_powersupplies_status_health",
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
	})

	if !reflect.DeepEqual(result, expected) {
		expectedOut, err := yaml.Marshal(expected)
		if err != nil {
			t.Fatal(err)
		}
		resultOut, err := yaml.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("generateRule() returned unexpected result;\nexpected:\n%s\nresult:\n%s", string(expectedOut), string(resultOut))
	}
}
