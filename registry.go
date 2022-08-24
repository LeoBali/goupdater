package main

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const myRegistryPath = "Software\\Appsitory Updater"
const firststartValue = "firststart"
const uninstallRegistryPath = "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall"
const uninstallRegistryPath6432 = "SOFTWARE\\WOW6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall"

type Software struct {
	name, version string
}

//var apps []Software
//

func toString(list [] Software) string {
	var sb strings.Builder
	for _, software := range list {
		sb.WriteString(software.name + ", ")
	}
	return sb.String()
}

func GetVerAndApps() (string, string) {
	//var bit string
	var appsOriginal, appsWow []Software
	var isWin64 bool
	apps1, _ := getAppsFromRegistry(registry.LOCAL_MACHINE, false)
	log.Printf("read %v apps from %v: %v", len(apps1), registry.LOCAL_MACHINE, toString(apps1))
	apps2, _ := getAppsFromRegistry(registry.CURRENT_USER, false)
	log.Printf("read %v apps from %v: %v", len(apps2), registry.CURRENT_USER, toString(apps2))
	appsOriginal = append(apps1, apps2...)
	apps3, err := getAppsFromRegistry(registry.LOCAL_MACHINE, true)
	if err != nil {
		log.Printf("error reading from %v wow6432. running on x86", registry.LOCAL_MACHINE)
		isWin64 = false
	} else {
		log.Printf("read %v apps from %v wow6432: %v", len(apps3), registry.LOCAL_MACHINE, toString(apps3))
		apps4, _ := getAppsFromRegistry(registry.CURRENT_USER, true)
		log.Printf("read %v apps from %v wow6432: %v", len(apps4), registry.CURRENT_USER, toString(apps4))
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

func ReadFirstStart() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, myRegistryPath, registry.ALL_ACCESS)
	if err != nil {
		log.Printf("cannot open registry key %s, creating new", myRegistryPath)
		k, openedExisting, err := registry.CreateKey(registry.CURRENT_USER, myRegistryPath, registry.ALL_ACCESS)
		if err != nil {
			log.Printf("error creating registry key %v", err)
			return false
		}
		if openedExisting {
			log.Println("opened existing registry path ?")
		}
		k.SetDWordValue(firststartValue, 0)
		return true
	}
	value, _, err := k.GetIntegerValue(firststartValue)
	if err != nil {
		log.Printf("cannot read registry value %s %v", firststartValue, err)
		return false
	}
	if value == 2 {
		return true
	}
	if value == 1 {
		k.SetDWordValue(firststartValue, 0)
		return true
	}
	return false
}

func getProductName() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	val, _, err := k.GetStringValue("ProductName")
	return val, err
}

func getAppsFromRegistry(hive registry.Key, wow6432 bool) (apps []Software, err error) {
	var k registry.Key
	if wow6432 {
		log.Println(uninstallRegistryPath6432)
		k, err = registry.OpenKey(hive, uninstallRegistryPath6432, registry.ENUMERATE_SUB_KEYS)
	} else {
		log.Println(uninstallRegistryPath)
		k, err = registry.OpenKey(hive, uninstallRegistryPath, registry.ENUMERATE_SUB_KEYS)
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
			key, err = registry.OpenKey(hive, uninstallRegistryPath6432+"\\"+name, registry.QUERY_VALUE)
		} else {
			key, err = registry.OpenKey(hive, uninstallRegistryPath+"\\"+name, registry.QUERY_VALUE)
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
