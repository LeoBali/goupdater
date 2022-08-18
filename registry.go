package main

import (
	"fmt"
	"strings"
	"golang.org/x/sys/windows/registry"
)

const registryPath = "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall"
const registryPath6432 = "SOFTWARE\\WOW6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall"

type Software struct {
	name, version string
}

var apps []Software
var isWin64 bool

func GetVerAndApps() (string, string) {
	var bit string
	apps = apps[:0]
	getAppsFromRegistry(registry.LOCAL_MACHINE, false)
	getAppsFromRegistry(registry.CURRENT_USER, false)
	err := getAppsFromRegistry(registry.LOCAL_MACHINE, true)
	if err != nil {
		bit = "32"
	} else {
		getAppsFromRegistry(registry.CURRENT_USER, true)
		bit = "64"
	}
	var sb strings.Builder
	for _, app := range apps {
		//sb.WriteString(fmt.Sprintf("%s;;%s\r\n", app.name, app.version))
		sb.WriteString(fmt.Sprintf("%s;;%s\r\n", app.name, "1.0.0"))
	}
	productName, _ := getProductName()
	return fmt.Sprintf("%s;%s", productName, bit), sb.String();
}

func getProductName() (string, error) {
	// final key = Registry.openPath(RegistryHive.localMachine, path: "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion");
	// productName = key.getValueAsString("ProductName");
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	val, _, err := k.GetStringValue("ProductName")
	return val, err
}

func getAppsFromRegistry(hive registry.Key, wow6432 bool) error {
	var k registry.Key
	var err error
	if wow6432 {
		k, err = registry.OpenKey(hive, registryPath6432, registry.ENUMERATE_SUB_KEYS)
	} else {
		k, err = registry.OpenKey(hive, registryPath, registry.ENUMERATE_SUB_KEYS)
	}
	if err != nil {
		return err
	}
	defer k.Close()
	keys, _ := k.ReadSubKeyNames(0)
	for _, name := range keys {
		var key registry.Key
		var softwareName, softwareVersion, systemComponent string
		var iSystemComponent uint64
		if wow6432 {
			key, err = registry.OpenKey(hive, registryPath6432+"\\"+name, registry.QUERY_VALUE)
		} else {
			key, err = registry.OpenKey(hive, registryPath+"\\"+name, registry.QUERY_VALUE)
		}
		if err != nil {
			goto close
		}
		softwareName, _, err = key.GetStringValue("DisplayName")
		if err != nil {
			goto close
		}
		softwareVersion, _, _ = key.GetStringValue("DisplayVersion")
		if strings.Contains(softwareName, " (KB") {
			goto close
		}
		iSystemComponent, _, err = key.GetIntegerValue("SystemComponent")
		if err == nil {
			if iSystemComponent != 0 {
				goto close
			}
		} else {
			if err == registry.ErrUnexpectedType {
				systemComponent, _, err = key.GetStringValue("SystemComponent")
				if err == nil {
					if systemComponent != "" {
						goto close
					}
				}
			}
		}
		apps = append(apps, Software{name: softwareName, version: softwareVersion})
	close:
		key.Close()
	}
	return nil
}