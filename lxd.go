package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"

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

func restartInstance(name string) error {
	_, err := lxcDo(fmt.Sprintf("lxc restart %s", name))
	return err
}

func startInstance(name string) error {
	_, err := lxcDo(fmt.Sprintf("lxc start %s", name))
	return err
}

func stopInstance(name string) error {
	_, err := lxcDo(fmt.Sprintf("lxc stop %s", name))
	return err
}

func deleteInstance(name string) error {
	_, err := lxcDo(fmt.Sprintf("lxc delete %s", name))
	return err
}

func lxcDo(cmdStr string) (out string, err error) {
	cmds := strings.Split(cmdStr, " ")
	cmd := exec.Command(cmds[0], cmds[1:]...)
	var stdout io.ReadCloser
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return "", err
	}
	defer stdout.Close()
	cmd.Start()
	err = cmd.Wait()
	if err != nil {
		return "", err
	}
	byts, _ := ioutil.ReadAll(stdout)
	return string(byts), nil
}
