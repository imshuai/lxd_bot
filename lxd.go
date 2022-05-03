package main

import (
	lxd "github.com/lxc/lxd/client"
)

func getHostInfo(hostID int) error {
	instance, err := lxd.ConnectLXD("https://hax.v2up.tk:8443", &lxd.ConnectionArgs{
		TLSClientCert: "", //TODO: 完善客户端证书配置,
		TLSClientKey:  "",
	})
	if err != nil {
		return err
	}
	defer instance.Disconnect()
	return nil
}
