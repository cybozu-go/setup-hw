package redfish

import (
	"context"
	"math"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	prommodel "github.com/prometheus/client_model/go"
)

func TestStubClient(t *testing.T) {
	t.Parallel()

	expectedSet := []*metrics{
		{
			name:  "hw_chassis_networkadapters_networkdevicefunctions_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis":  "System.Embedded.1",
				"nic":      "NIC.Integrated.1",
				"function": "NIC.Integrated.1-1-1",
			},
		},
		{
			name:  "hw_chassis_networkadapters_networkdevicefunctions_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis":  "System.Embedded.1",
				"nic":      "NIC.Integrated.1",
				"function": "NIC.Integrated.1-1-1",
			},
		},
		{
			name:  "hw_chassis_networkadapters_networkports_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"nic":     "NIC.Integrated.1",
				"port":    "NIC.Integrated.1-1",
			},
		},
		{
			name:  "hw_chassis_networkadapters_networkports_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"nic":     "NIC.Integrated.1",
				"port":    "NIC.Integrated.1-1",
			},
		},
		{
			name:  "hw_chassis_networkadapters_networkports_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"nic":     "NIC.Integrated.1",
				"port":    "NIC.Integrated.1-2",
			},
		},
		{
			name:  "hw_chassis_networkadapters_networkports_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"nic":     "NIC.Integrated.1",
				"port":    "NIC.Integrated.1-2",
			},
		},
		{
			name:  "hw_chassis_pciedevices_pciefunctions_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis":  "System.Embedded.1",
				"device":   "0-1",
				"function": "0-1-0",
			},
		},
		{
			name:  "hw_chassis_pciedevices_pciefunctions_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis":  "System.Embedded.1",
				"device":   "0-1",
				"function": "0-1-0",
			},
		},
		{
			name:  "hw_chassis_pciedevices_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"device":  "0-1",
			},
		},
		{
			name:  "hw_chassis_pciedevices_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"device":  "0-1",
			},
		},
		{
			name:  "hw_chassis_pcieslots_slots_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"slot":    "0",
			},
		},
		{
			name:  "hw_chassis_power_powersupplies_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis":     "System.Embedded.1",
				"powersupply": "0",
			},
		},
		{
			name:  "hw_chassis_power_powersupplies_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis":     "System.Embedded.1",
				"powersupply": "0",
			},
		},
		{
			name:  "hw_chassis_power_voltages_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"voltage": "0",
			},
		},
		{
			name:  "hw_chassis_power_voltages_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"voltage": "0",
			},
		},
		{
			name:  "hw_chassis_powersubsystem_powersupplies_metrics_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"unit":    "PSU.Slot.1",
			},
		},
		{
			name:  "hw_chassis_powersubsystem_powersupplies_metrics_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"unit":    "PSU.Slot.1",
			},
		},
		{
			name:  "hw_chassis_powersubsystem_powersupplies_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"unit":    "PSU.Slot.1",
				"chassis": "System.Embedded.1",
			},
		},
		{
			name:  "hw_chassis_powersubsystem_powersupplies_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"unit":    "PSU.Slot.1",
			},
		},
		{
			name:  "hw_chassis_powersubsystem_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
			},
		},
		{
			name:  "hw_chassis_powersubsystem_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
			},
		},
		{
			name:  "hw_chassis_sensors_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"sensor":  "CPU1MEMABCDVR",
			},
		},
		{
			name:  "hw_chassis_sensors_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"sensor":  "CPU1MEMABCDVR",
			},
		},
		{
			name:  "hw_chassis_sensors_thresholds_lowercaution_reading",
			typ:   prommodel.MetricType_GAUGE,
			value: math.NaN(),
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"sensor":  "PS1Current1",
			},
		},
		{
			name:  "hw_chassis_sensors_thresholds_lowercaution_reading",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"sensor":  "Fan.Embedded.4D",
			},
		},
		{
			name:  "hw_chassis_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "Enclosure.Internal.0-1",
			},
		},
		{
			name:  "hw_chassis_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "Enclosure.Internal.0-1",
			},
		},
		{
			name:  "hw_chassis_thermal_fans_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"fan":     "0",
				"chassis": "System.Embedded.1",
			},
		},
		{
			name:  "hw_chassis_thermal_fans_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"fan":     "0",
			},
		},
		{
			name:  "hw_chassis_thermal_redundancy_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"redundancy": "0",
				"chassis":    "System.Embedded.1",
			},
		},
		{
			name:  "hw_chassis_thermal_redundancy_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis":    "System.Embedded.1",
				"redundancy": "0",
			},
		},
		{
			name:  "hw_chassis_thermal_temperatures_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"temperature": "0",
				"chassis":     "System.Embedded.1",
			},
		},
		{
			name:  "hw_chassis_thermal_temperatures_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis":     "System.Embedded.1",
				"temperature": "0",
			},
		},
		{
			name:  "hw_chassis_thermalsubsystem_fanredundancy_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis":       "System.Embedded.1",
				"fanredundancy": "0",
			},
		},
		{
			name:  "hw_chassis_thermalsubsystem_fanredundancy_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis":       "System.Embedded.1",
				"fanredundancy": "0",
			},
		},
		{
			name:  "hw_chassis_thermalsubsystem_fans_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"unit":    "Fan.Embedded.1A",
			},
		},
		{
			name:  "hw_chassis_thermalsubsystem_fans_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
				"unit":    "Fan.Embedded.1A",
			},
		},
		{
			name:  "hw_chassis_thermalsubsystem_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
			},
		},
		{
			name:  "hw_chassis_thermalsubsystem_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"chassis": "System.Embedded.1",
			},
		},
		{
			name:  "hw_fabrics_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"fabric": "PCIe",
			},
		},
		{
			name:  "hw_fabrics_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"fabric": "PCIe",
			},
		},
		{
			name:  "hw_managers_ethernetinterfaces_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"manager":   "iDRAC.Embedded.1",
				"interface": "NIC.1",
			},
		},
		{
			name:  "hw_managers_ethernetinterfaces_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"interface": "NIC.1",
				"manager":   "iDRAC.Embedded.1",
			},
		},
		{
			name:  "hw_managers_networkprotocol_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"manager": "iDRAC.Embedded.1",
			},
		},
		{
			name:  "hw_managers_networkprotocol_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"manager": "iDRAC.Embedded.1",
			},
		},
		{
			name:  "hw_managers_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"manager": "iDRAC.Embedded.1",
			},
		},
		{
			name:  "hw_managers_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"manager": "iDRAC.Embedded.1",
			},
		},
		{
			name:  "hw_systems_ethernetinterfaces_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"interface": "NIC.Embedded.1-1-1",
				"system":    "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_ethernetinterfaces_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"interface": "NIC.Embedded.1-1-1",
				"system":    "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_hostwatchdogtimer_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 1.000000,
			labels: map[string]string{
				"system": "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_memory_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"memory": "DIMM.Socket.A1",
				"system": "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_memory_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"memory": "DIMM.Socket.A1",
				"system": "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_memorysummary_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system": "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_memorysummary_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system": "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_networkinterfaces_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"nic":    "NIC.Embedded.1",
				"system": "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_networkinterfaces_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"nic":    "NIC.Embedded.1",
				"system": "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_processors_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"processor": "CPU.Socket.1",
				"system":    "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_processors_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"processor": "CPU.Socket.1",
				"system":    "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_processorsummary_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system": "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_processorsummary_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system": "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_simplestorage_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "CPU.1",
			},
		},
		{
			name:  "hw_systems_simplestorage_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "CPU.1",
			},
		},
		{
			name:  "hw_systems_simplestorage_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: -1.000000,
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"controller": "AHCI.Embedded.1-1",
			},
		},
		{
			name:  "hw_systems_simplestorage_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"controller": "AHCI.Embedded.1-1",
				"system":     "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system": "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system": "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_storage_controllers_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: -1.000000,
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"storage":    "AHCI.Embedded.1-1",
				"controller": "AHCI.Embedded.1-1",
			},
		},
		{
			name:  "hw_systems_storage_controllers_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system":     "System.Embedded.1",
				"storage":    "AHCI.Embedded.1-1",
				"controller": "AHCI.Embedded.1-1",
			},
		},
		{
			name:  "hw_systems_storage_drives_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"device":  "Disk.Bay.0:Enclosure.Internal.0-1",
				"storage": "CPU.1",
				"system":  "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_storage_drives_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system":  "System.Embedded.1",
				"device":  "Disk.Bay.0:Enclosure.Internal.0-1",
				"storage": "CPU.1",
			},
		},
		{
			name:  "hw_systems_storage_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: -1.000000,
			labels: map[string]string{
				"storage": "AHCI.Embedded.1-1",
				"system":  "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_storage_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"storage": "AHCI.Embedded.1-1",
				"system":  "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_storage_storagecontrollers_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: -1.000000,
			labels: map[string]string{
				"controller": "0",
				"storage":    "AHCI.Embedded.1-1#",
				"system":     "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_storage_storagecontrollers_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"storage":    "AHCI.Embedded.1-1#",
				"system":     "System.Embedded.1",
				"controller": "0",
			},
		},
		{
			name:  "hw_systems_storage_storagecontrollers_storagecontrollers_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: -1.000000,
			labels: map[string]string{
				"controller":        "0",
				"storage":           "AHCI.Embedded.1-1#",
				"storagecontroller": "0",
				"system":            "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_storage_storagecontrollers_storagecontrollers_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"controller":        "0",
				"storage":           "AHCI.Embedded.1-1#",
				"storagecontroller": "0",
				"system":            "System.Embedded.1",
			},
		},
		{
			name:  "hw_systems_storage_volumes_status_health",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system":  "System.Embedded.1",
				"storage": "CPU.1",
				"volume":  "Disk.Bay.0:Enclosure.Internal.0-1",
			},
		},
		{
			name:  "hw_systems_storage_volumes_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system":  "System.Embedded.1",
				"storage": "CPU.1",
				"volume":  "Disk.Bay.0:Enclosure.Internal.0-1",
			},
		},
		{
			name:  "hw_systems_trustedmodules_status_state",
			typ:   prommodel.MetricType_GAUGE,
			value: 0.000000,
			labels: map[string]string{
				"system":        "System.Embedded.1",
				"trustedmodule": "0",
			},
		},
	}

	rule, err := collectRule("../redfish/rules/dell_redfish_1.20.1.yml")
	if err != nil {
		t.Fatal(err)
	}

	client := NewStubClient("../testdata/redfish-1.20-from-idrac9-v11.json")

	checkActualResult(t, rule, client, expectedSet)
}

func checkActualResult(t *testing.T, rule *CollectRule, client Client, expectedSet []*metrics) {

	collector, err := NewCollector(func(context.Context) (*CollectRule, error) {
		return rule, nil
	}, client)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	collector.Update(ctx)

	registry := prometheus.NewRegistry()
	err = registry.Register(collector)
	if err != nil {
		t.Fatal(err)
	}

	// This calls collector.Collect() internally.
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	actualTestDataSet := make([]*metrics, 0)

	for _, metricFamily := range metricFamilies {
		for _, m := range metricFamily.GetMetric() {
			actualLabels := make(map[string]string)
			for _, label := range m.GetLabel() {
				actualLabels[label.GetName()] = label.GetValue()
			}
			var val float64
			switch *metricFamily.Type {
			case prommodel.MetricType_GAUGE:
				val = m.GetGauge().GetValue()
			case prommodel.MetricType_COUNTER:
				val = m.GetCounter().GetValue()
			default:
				t.Fatalf("unknown type: ")
			}
			actualTestData := &metrics{
				name:   metricFamily.GetName(),
				labels: actualLabels,
				typ:    *metricFamily.Type,
				value:  val,
			}
			actualTestDataSet = append(actualTestDataSet, actualTestData)
		}
	}

	for _, expected := range expectedSet {
		found := false
		for _, testCase := range actualTestDataSet {
			if expected.key() == testCase.key() {
				found = true
			}
		}
		if !found {
			t.Errorf("expected metric not found: %s", expected.key())
		}
	}

}
