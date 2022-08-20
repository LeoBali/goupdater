package main

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"strings"
)

const registryPath = "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall"
const registryPath6432 = "SOFTWARE\\WOW6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall"

type Software struct {
	name, version string
}

//var apps []Software
//

func GetVerAndApps() (string, string) {
	//var bit string
	var appsOriginal, appsWow []Software
	var isWin64 bool
	apps1, _ := getAppsFromRegistry(registry.LOCAL_MACHINE, false)
	apps2, _ := getAppsFromRegistry(registry.CURRENT_USER, false)
	appsOriginal = append(apps1, apps2...)
	apps3, err := getAppsFromRegistry(registry.LOCAL_MACHINE, true)
	if err != nil {
		isWin64 = false
	} else {
		apps4, _ := getAppsFromRegistry(registry.CURRENT_USER, true)
		isWin64 = true
		appsWow = append(apps3, apps4...)
	}
	var sb strings.Builder
	productName, _ := getProductName()
	if isWin64 {
		for _, app := range appsOriginal {
			sb.WriteString(fmt.Sprintf("%s;;%s;;64\r\n", app.name, app.version))
		}
		for _, app := range appsWow {
			sb.WriteString(fmt.Sprintf("%s;;%s;;0\r\n", app.name, app.version))
		}
		return fmt.Sprintf("%s;%s", productName, "64"), sb.String()
	} else {
		for _, app := range appsOriginal {
			sb.WriteString(fmt.Sprintf("%s;;%s;;0\r\n", app.name, app.version))
		}
		return fmt.Sprintf("%s;%s", productName, "32"), sb.String()
	}
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

func getAppsFromRegistry(hive registry.Key, wow6432 bool) (apps []Software, err error) {
	var k registry.Key
	//var err error
	if wow6432 {
		k, err = registry.OpenKey(hive, registryPath6432, registry.ENUMERATE_SUB_KEYS)
	} else {
		k, err = registry.OpenKey(hive, registryPath, registry.ENUMERATE_SUB_KEYS)
	}
	if err != nil {
		return nil, err
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
		softwareVersion = "1.0.0"
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
	return apps, nil
}
