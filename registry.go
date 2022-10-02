package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const myRegistryPath = "Software\\Appsitory Updater"
const firststartValue = "firststart"
const uninstallRegistryPath = "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall"
const uninstallRegistryPath6432 = "SOFTWARE\\WOW6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall"

const runRegistryPath = "Software\\Microsoft\\Windows\\CurrentVersion\\Run";
const runRegistryPath6432 = "Software\\WOW6432Node\\Microsoft\\Windows\\CurrentVersion\\Run";

type Software struct {
	name, version string
}

//var apps []Software
//

func logApps(list []Software) {
	for _, software := range list {
		log.Printf("  %s %s\r\n", software.name, software.version)
	}
}

const autorunKeyName = "Appsitory Updater"

func openRunKey() (registry.Key, error) {
	x64System := GetCPUArch()
	var k registry.Key
	var err error
	if !x64System {
		k, err = registry.OpenKey(registry.CURRENT_USER, runRegistryPath, registry.QUERY_VALUE | registry.SET_VALUE)
		return k, err
	} else {
		x86 := false
		if x86 {
			k, err = registry.OpenKey(registry.CURRENT_USER, runRegistryPath6432, registry.QUERY_VALUE | registry.SET_VALUE | registry.WOW64_32KEY)
		} else {
			k, err = registry.OpenKey(registry.CURRENT_USER, runRegistryPath, registry.QUERY_VALUE | registry.SET_VALUE |registry.WOW64_64KEY)
		}
		return k, err
	}
}

func GetAutorun() (bool, error) {
	k, err := openRunKey()
	if err != nil {
		return false, err
	}
	log.Println("key opened")
	value, _, err := k.GetStringValue(autorunKeyName)
	log.Printf("got value %value %err", value, err)
	if err != nil {
		log.Println("autorun key empty")
		return false, err
	}
	log.Printf("autorun key value: %v", value)
	return true, nil
}

func SetAutorun(value bool) error {
	k, err := openRunKey()
	if err != nil {
		return err
	}
	if value {
		executable, err := os.Executable()
		if err != nil {
			executable = os.Args[0]
		}
		log.Printf("setting autorun key %v", executable)
		err = k.SetStringValue(autorunKeyName, executable)
		if err != nil {
			log.Printf("err: %v", err)
		}
	} else {
		log.Println("deleting autorun key")
		err = k.DeleteValue(autorunKeyName)
		if err != nil {
			log.Printf("err: %v", err)
		}
	}
	return nil
}

func GetVerAndApps() (string, string) {
	//var bit string
	var appsX86, appsX64 []Software
	is64 := GetCPUArch()
	log.Printf("cpu architecture 64: %v", is64)
	if is64 {
		apps1, _ := getAppsFromRegistry(true, false, true)
		log.Printf("read %v apps from HKLM 86\r\n", len(apps1))
		logApps(apps1)
		apps2, _ := getAppsFromRegistry(false, false, true)
		log.Printf("read %v apps from HKCU 86\r\n", len(apps2))
		logApps(apps2)
		appsX86 = append(apps1, apps2...)
		apps3, _ := getAppsFromRegistry(true, false, false)
		log.Printf("read %v apps from HKLM 64\r\n", len(apps3))
		logApps(apps3)
		apps4, _ := getAppsFromRegistry(false, false, false)
		log.Printf("read %v apps from HKCU 64\r\n", len(apps4))
		logApps(apps4)
		appsX64 = append(apps3, apps4...)
	} else {
		apps1, _ := getAppsFromRegistry(true, true, true)
		log.Printf("read %v apps from HKLM\r\n", len(apps1))
		logApps(apps1)
		apps2, _ := getAppsFromRegistry(false, true, true)
		log.Printf("read %v apps from HKCU\r\n", len(apps2))
		logApps(apps2)
		appsX86 = append(apps1, apps2...)
	}

	var sb strings.Builder
	productName, _ := getProductName()
	for _, app := range appsX86 {
		sb.WriteString(fmt.Sprintf("%s;;%s;;0\r\n", app.name, app.version))
	}
	for _, app := range appsX64 {
		sb.WriteString(fmt.Sprintf("%s;;%s;;64\r\n", app.name, app.version))
	}
	if is64 {
		return fmt.Sprintf("%s;%s", productName, "64"), sb.String()
	} else {
		return fmt.Sprintf("%s;%s", productName, "32"), sb.String()
	}
}

func ReadFirstStart() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, myRegistryPath, registry.ALL_ACCESS)
	if err != nil {
		log.Printf("cannot open registry key %s, creating new\r\n", myRegistryPath)
		k, openedExisting, err := registry.CreateKey(registry.CURRENT_USER, myRegistryPath, registry.ALL_ACCESS)
		if err != nil {
			log.Printf("error creating registry key %v\r\n", err)
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
		log.Printf("cannot read registry value %s %v\r\n", firststartValue, err)
		return false
	}
	log.Printf("first start value %d", value)
	if value == 2 {
		return true
	}
	if value == 1 {
		k.SetDWordValue(firststartValue, 0)
		return true
	}
	return false
}

func GetCPUArch() (is64 bool) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment", registry.QUERY_VALUE)
	if err == nil {
		val, _, err := k.GetStringValue("PROCESSOR_ARCHITECTURE")
		if err == nil {
			if strings.Contains(val, "AMD64") {
				return true
			}
			if strings.Contains(val, "x86") {
				return false
			}
		}
	}
	k, err = registry.OpenKey(registry.LOCAL_MACHINE, "HARDWARE\\DESCRIPTION\\SYSTEM\\CENTERALPROCESSOR\\0", registry.QUERY_VALUE)
	if err == nil {
		val, _, err := k.GetStringValue("Identifier")
		if err == nil {
			if strings.Contains(val, "Intel64") {
				return true
			}
			if strings.Contains(val, "x64") {
				return true
			}
			if strings.Contains(val, "Intel32") {
				return false
			}
			if strings.Contains(val, "x86") {
				return false
			}
		}
	}
	fmt.Println("cannot detect CPU architecture")
	return false
}

func GetCurrentVersion() (major uint64, minor uint64, err error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", registry.QUERY_VALUE)
	if err != nil {
		return 0, 0, err
	}
	major, _, err = k.GetIntegerValue("CurrentMajorVersionNumber")
	if err != nil {
		return 0, 0, err
	}

	minor, _, err = k.GetIntegerValue("CurrentMinorVersionNumber")
	if err != nil {
		return 0, 0, err
	}
	return major, minor, nil
}

func getProductName() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	val, _, err := k.GetStringValue("ProductName")
	return val, err
}

func getAppsFromRegistry(localMachine bool, x86System bool, x86 bool) (apps []Software, err error) {
	var hive registry.Key
	if localMachine {
		log.Printf("getAppsFromRegistry localMachine x86System: %v x86: %v", x86System, x86)
		hive = registry.LOCAL_MACHINE
	} else {
		log.Printf("getAppsFromRegistry currentUser x86System: %v x86: %v", x86System, x86)
		hive = registry.CURRENT_USER
	}

	var k registry.Key
	if x86System {
		k, err = registry.OpenKey(hive, uninstallRegistryPath, registry.ENUMERATE_SUB_KEYS)
	} else {
		if x86 {
			k, err = registry.OpenKey(hive, uninstallRegistryPath6432, registry.ENUMERATE_SUB_KEYS|registry.WOW64_32KEY)
		} else {
			k, err = registry.OpenKey(hive, uninstallRegistryPath, registry.ENUMERATE_SUB_KEYS|registry.WOW64_64KEY)
		}
	}
	if err != nil {
		log.Printf("error in getAppsFromRegistry %v", err)
		return make([]Software, 0), err
	}
	defer k.Close()
	keys, _ := k.ReadSubKeyNames(0)
	log.Printf("sub keys %v", len(keys))
	for _, name := range keys {
		var key registry.Key
		var softwareName, softwareVersion, systemComponent string
		var iSystemComponent uint64
		if x86System {
			key, err = registry.OpenKey(hive, uninstallRegistryPath+"\\"+name, registry.QUERY_VALUE)
		} else {
			if x86 {
				key, err = registry.OpenKey(hive, uninstallRegistryPath6432+"\\"+name, registry.QUERY_VALUE|registry.WOW64_32KEY)
			} else {
				key, err = registry.OpenKey(hive, uninstallRegistryPath+"\\"+name, registry.QUERY_VALUE|registry.WOW64_64KEY)
			}
		}
		if err != nil {
			goto close
		}
		softwareName, _, err = key.GetStringValue("DisplayName")
		if err != nil {
			goto close
		}
		softwareVersion, _, _ = key.GetStringValue("DisplayVersion")
		//softwareVersion = "1.0.0"
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
