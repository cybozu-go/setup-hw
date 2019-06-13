Writing Rules
=============

One rule file can be applied to one combination of vendor, model and firmware version.
You should update [collection rules](rule.md) when you introduce new model machine, update firmware and so on.

How to update rule
------------------

1. Run `collector show`.
   ```console
   ./collector show --no-dup=true --ignore-field=Attributes --ignore-field="@odata.*" --omitempty --base-rule=base-rule.yaml
   ```
1. List up keys that you want to collect metrics. Such as `Status`, `ReadingCelsius`.
1. Run `collector generate-rule` with the keys on nodes by hardware type.
1. Edit the generated rule file.
    1. Add `Traverse` and its descendants fields.

    1. Combine rules with similar path.
    
        You will get a generated rules like the following.
    
        ```yaml
          - Path: /redfish/v1/Chassis/Chassis1
            Properties:
              - Pointer: /Status/Health
                Name: chassis_chassis1_status_health
                Type: health
          - Path: /redfish/v1/Chassis/Chassis2
            Properties:
              - Pointer: /Status/Health
                Name: chassis_chassis2_status_health
                Type: health
        ```
        
        You can rewrite the rule by using pattern.
        
        ```yaml
          - Path: /redfish/v1/Chassis/{chassis}
            Properties:
              - Pointer: /Status/Health
                Name: chassis_status_health
                Type: health
        ```
        
    1. Rewrite pattern name of a property.
    
        You may have a `Pointer` fields filled by `{TBD}`.
        
        ```yaml
          - Path: /redfish/v1/Chassis/Chassis/Power
            Properties:
              - Pointer: /PowerSupplies/{TBD}/Status/Health
                Name: chassis_psu_status_health
                Type: health
        ```
        
        You can rewrite it to a meaningful name.
        
        ```yaml
          - Path: /redfish/v1/Chassis/Chassis/Power
            Properties:
              - Pointer: /PowerSupplies/{psu}/Status/Health
                Name: chassis_psu_status_health
                Type: health
        ```

1. Name the rule file `<vendor>_<role>_<firmware version>.yml`. Such as `dell_boot_1.4.0.yml`, `dell_cs_1.4.0.yml`, `supermicro_ss_1.2.0.yml`.
1. Put the rule file under `redfish/rules/` directory.
1. Specify `role` argument of `monitor-hw` on each node. (`vendor` and `firmware version` will be detected automatically)
