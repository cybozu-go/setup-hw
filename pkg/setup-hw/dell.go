package main

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"gopkg.in/ini.v1"
)

const (
	racadmPath = "/opt/dell/srvadmin/bin/idracadm7"

	retryCount = 5
)

func racadm(ctx context.Context, args ...string) (string, error) {
	cmd := well.CommandContext(ctx, racadmPath, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func racadmRetry(ctx context.Context, args ...string) error {
	retries := 0
RETRY:
	cmd := well.CommandContext(ctx, racadmPath, args...)
	err := cmd.Run()
	if err == nil {
		time.Sleep(1 * time.Second)
		return nil
	}

	retries++
	if retries == retryCount {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Second):
	}
	goto RETRY
}

// racadmGetConfig returns a string corresponding to the given key.
// In many cases 'idracadm7 get KEY' returns INI format output,
// but in some cases it returns the value not conforming to INI format.
//
// case1: conform to INI format
//     $ sudo idracadm7 get iDRAC.SNMP.AgentEnable
//     [Key=iDRAC.Embedded.1#SNMP.1]
//     AgentEnable=Enabled
//
// case2: not conform to INI format. return only value.
//     $ sudo idracadm7 get System.ServerPwr.PSRapidOn
//     Enabled
//
func racadmGetConfig(ctx context.Context, key string) (string, error) {
	cmd := well.CommandContext(ctx, racadmPath, "get", key)
	cmd.Severity = log.LvDebug
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return parseRacadmGetOutput(string(out), key)
}

func parseRacadmGetOutput(out, key string) (string, error) {
	if !strings.HasPrefix(out, "[") {
		return strings.TrimSpace(out), nil
	}

	cfg, err := ini.Load([]byte(out))
	if err != nil {
		return "", err
	}

	var sectionName string
	for _, name := range cfg.SectionStrings() {
		if name != ini.DEFAULT_SECTION {
			sectionName = name
			break
		}
	}

	section := cfg.Section(sectionName)
	keys := section.Keys()
	if len(keys) == 0 {
		return "", errors.New("unexpected output for " + key)
	}

	return keys[0].String(), nil
}

// racadmSetConfig check the current value of key and compares it to value.
// If the current value is the same as value, this returns (false, nil).
// Otherwise, this sets key to value and returns (true, nil).
func racadmSetConfig(ctx context.Context, key, value string) (bool, error) {
	cur, err := racadmGetConfig(ctx, key)
	if err != nil {
		return false, err
	}
	if cur == value {
		return false, nil
	}

	err = racadmRetry(ctx, "set", key, value)
	if err != nil {
		return false, err
	}
	return true, nil
}

func iDRACWait(ctx context.Context) error {
	for i := 0; i < 60; i++ {
		out, _ := racadm(ctx, "get", "iDRAC.Info.Name")
		if strings.Contains(out, "Name=iDRAC") {
			return nil
		}

		log.Info("waiting iDRAC...", nil)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}

	return errors.New("giving up waiting for iDRAC")
}

type dellConfigurator struct {
	addressConfig *AddressConfig
	userConfig    *UserConfig
	queued        bool
}

func (dc *dellConfigurator) Run(ctx context.Context) error {
	if err := iDRACWait(ctx); err != nil {
		return err
	}

	out, err := racadm(ctx, "jobqueue", "view")
	if err != nil {
		return err
	}
	if strings.Contains(out, "Status=Scheduled") {
		dc.queued = true
		log.Warn("scheduled jobs exist", map[string]interface{}{
			"output": out,
		})
		return errors.New("scheduled jobs are queued")
	}

	if err := dc.configBIOS(ctx); err != nil {
		return err
	}
	if err := dc.configSystem(ctx); err != nil {
		return err
	}
	if err := dc.configiDRAC(ctx); err != nil {
		return err
	}

	if dc.queued {
		if _, err = racadm(ctx, "jobqueue", "create", "BIOS.Setup.1-1"); err != nil {
			return err
		}
	}

	return nil
}

func (dc *dellConfigurator) enqueueConfig(ctx context.Context, key, value string) error {
	updated, err := racadmSetConfig(ctx, key, value)
	if err != nil {
		return err
	}
	if updated {
		dc.queued = true
	}
	return nil
}

func (dc *dellConfigurator) configBIOS(ctx context.Context) error {
	if err := dc.configPerformance(ctx); err != nil {
		return err
	}
	if err := dc.configProcessor(ctx); err != nil {
		return err
	}
	if err := dc.configTpmSecurity(ctx); err != nil {
		return err
	}

	key := "BIOS.SysSecurity.TpmStatus"
	val, err := racadmGetConfig(ctx, key)
	if err != nil {
		return err
	}
	if val != "Enabled, Activated" {
		if err := dc.configTpmCommand(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (dc *dellConfigurator) configPerformance(ctx context.Context) error {
	return dc.enqueueConfig(ctx, "BIOS.SysProfileSettings.SysProfile", "PerfPerWattOptimizedOs")
}

func (dc *dellConfigurator) configProcessor(ctx context.Context) error {
	return dc.enqueueConfig(ctx, "BIOS.ProcSettings.LogicalProc", "Disabled")
}

func (dc *dellConfigurator) configTpmSecurity(ctx context.Context) error {
	return dc.enqueueConfig(ctx, "BIOS.SysSecurity.TpmSecurity", "OnPbm")
}

func (dc *dellConfigurator) configTpmCommand(ctx context.Context) error {
	return dc.enqueueConfig(ctx, "BIOS.SysSecurity.TpmCommand", "Activate")
}

func (dc *dellConfigurator) configSystem(ctx context.Context) error {
	if err := dc.configPowerSupply(ctx); err != nil {
		return err
	}
	if err := dc.configFanSpeed(ctx); err != nil {
		return err
	}
	return nil
}

func (dc *dellConfigurator) configPowerSupply(ctx context.Context) error {
	_, err := racadmSetConfig(ctx, "System.ServerPwr.PSRapidOn", "Disabled")
	return err
}

// configFanSpeed adjusts fan speed calculation algorithm.
// ref: https://www.dell.com/support/article/us/en/04/sln283419/adjusting-fan-speed-offset-in-dell-poweredge-12th-generation-servers?lang=en
func (dc *dellConfigurator) configFanSpeed(ctx context.Context) error {
	key := "System.ThermalSettings.FanSpeedOffset"
	val, err := racadmGetConfig(ctx, key)
	if err != nil {
		return err
	}

	if val == "Low" {
		return nil
	}

	return racadmRetry(ctx, "set", key, "0")
}

func (dc *dellConfigurator) configiDRAC(ctx context.Context) error {
	if err := dc.configSNMP(ctx); err != nil {
		return err
	}
	if err := dc.configNIC(ctx); err != nil {
		return err
	}
	if err := dc.configIPMI(ctx); err != nil {
		return err
	}
	if err := dc.configUsers(ctx); err != nil {
		return err
	}
	if err := dc.configVirtualConsole(ctx); err != nil {
		return err
	}
	return nil
}

func (dc *dellConfigurator) configSNMP(ctx context.Context) error {
	_, err := racadmSetConfig(ctx, "iDRAC.SNMP.AgentEnable", "Enabled")
	return err
}

func (dc *dellConfigurator) configNIC(ctx context.Context) error {
	if _, err := racadmSetConfig(ctx, "iDRAC.NIC.Selection", "Dedicated"); err != nil {
		return err
	}
	if _, err := racadmSetConfig(ctx, "iDRAC.IPv4.DHCPEnable", "Disabled"); err != nil {
		return err
	}
	cfg := dc.addressConfig.IPv4
	if _, err := racadmSetConfig(ctx, "iDRAC.IPv4.Address", cfg.Address); err != nil {
		return err
	}
	if _, err := racadmSetConfig(ctx, "iDRAC.IPv4.Netmask", cfg.Netmask); err != nil {
		return err
	}
	if _, err := racadmSetConfig(ctx, "iDRAC.IPv4.Gateway", cfg.Gateway); err != nil {
		return err
	}
	hname, err := os.Hostname()
	if err != nil {
		return err
	}
	if _, err := racadmSetConfig(ctx, "iDRAC.NIC.DNSRacName", hname+"-idrac"); err != nil {
		return err
	}
	return nil
}

func (dc *dellConfigurator) configIPMI(ctx context.Context) error {
	if _, err := racadmSetConfig(ctx, "iDRAC.IPMILan.PrivLimit", "3"); err != nil {
		return err
	}
	key := "iDRAC.IPMILan.Enable"
	value, err := racadmGetConfig(ctx, key)
	if err != nil {
		return err
	}
	if value == "Enabled" {
		return nil
	}
	return racadmRetry(ctx, "set", key, "1")
}

func (dc *dellConfigurator) configUser(ctx context.Context, idx, name, priv, ipmiPriv string, cred Credentials) error {
	// ipmipriv:
	// - 1 Callback level
	// - 2 User level
	// - 3 Operator level
	// - 4 Administrator level
	// - 5 OEM Proprietary level
	// - 15 No access

	prefix := "iDRAC.Users." + idx + "."
	if _, err := racadmSetConfig(ctx, prefix+"Username", name); err != nil {
		return err
	}
	if cred.Password.Raw != "" {
		if err := racadmRetry(ctx, "set", prefix+"Password", cred.Password.Raw); err != nil {
			return err
		}
	} else {
		if _, err := racadmSetConfig(ctx, prefix+"SHA256Password", cred.Password.Hash); err != nil {
			return err
		}
		if _, err := racadmSetConfig(ctx, prefix+"SHA256PasswordSalt", cred.Password.Salt); err != nil {
			return err
		}
	}

	if _, err := racadm(ctx, "sshpkauth", "-i", idx, "-d", "-k", "all"); err != nil {
		return err
	}
	for i, k := range cred.AuthorizedKeys {
		kidx := strconv.Itoa(i + 1)
		if _, err := racadm(ctx, "sshpkauth", "-i", idx, "-k", kidx, "-t", k); err != nil {
			return err
		}
	}

	if _, err := racadmSetConfig(ctx, prefix+"Privilege", priv); err != nil {
		return err
	}
	if _, err := racadmSetConfig(ctx, prefix+"IpmiLanPrivilege", ipmiPriv); err != nil {
		return err
	}
	if _, err := racadmSetConfig(ctx, prefix+"IpmiSerialPrivilege", ipmiPriv); err != nil {
		return err
	}
	if _, err := racadmSetConfig(ctx, prefix+"Enable", "Enabled"); err != nil {
		return err
	}

	return nil
}

func (dc *dellConfigurator) configUsers(ctx context.Context) error {
	if err := dc.configUser(ctx, "2", "root", "0x1ff", "4", dc.userConfig.Root); err != nil {
		return err
	}
	if err := dc.configUser(ctx, "3", "support", "0x11", "15", dc.userConfig.Support); err != nil {
		return err
	}
	if err := dc.configUser(ctx, "4", "power", "0x11", "3", dc.userConfig.Power); err != nil {
		return err
	}
	return nil
}

func (dc *dellConfigurator) configVirtualConsole(ctx context.Context) error {
	_, err := racadmSetConfig(ctx, "iDRAC.VirtualConsole.PluginType", "2")
	return err
}

// setupDell configures BIOS and iDRAC for Dell servers.
func setupDell(ac *AddressConfig, uc *UserConfig) (bool, error) {
	_, err := os.Stat(racadmPath)
	if err != nil {
		return false, err
	}

	configurator := &dellConfigurator{
		addressConfig: ac,
		userConfig:    uc,
	}
	well.Go(configurator.Run)
	well.Stop()
	err = well.Wait()
	if err != nil {
		return false, err
	}

	return configurator.queued, nil
}
