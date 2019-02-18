// +build windows

package main

import (
	"golang.org/x/sys/windows/registry"
)

func setManifestRegistry(targetEnv, manifestPath string) error {
	var key string
	if targetEnv == "firefox" {
		key = `SOFTWARE\Mozilla\NativeMessagingHosts\`
	} else {
		key = `SOFTWARE\Google\Chrome\NativeMessagingHosts\`
	}
	key += "org.buster.client"

	k, _, err := registry.CreateKey(registry.CURRENT_USER, key, registry.WOW64_64KEY|registry.WRITE)
	if err != nil {
		return err
	}
	defer k.Close()

	if err := k.SetStringValue("", manifestPath); err != nil {
		return err
	}

	return nil
}
