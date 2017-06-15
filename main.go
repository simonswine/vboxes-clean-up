package main

// This sample demonstrates the creation of a VM. After running it, you should
// see a "vbox-sample-vm" machine in the VirtualBox UI.

import (
	"bufio"
	"os/exec"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type vboxVM struct {
	Name            string
	UUID            string
	State           string
	StateTime       string
	LastStateChange time.Time
}

func main() {
	log.Info("Listing VMs")
	cmd := exec.Command("VBoxManage", "list", "vms", "-l")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	reKeyValue := regexp.MustCompile(`^([A-Za-z]+): +(.*)$`)

	scanner := bufio.NewScanner(stdout)
	vm := &vboxVM{}
	var vms []*vboxVM
	for scanner.Scan() {
		text := scanner.Text()
		parts := reKeyValue.FindStringSubmatch(text)
		if len(parts) > 2 {
			if parts[1] == "Name" {
				if strings.HasPrefix(parts[2], "'hosthome'") {
					continue
				}
				vm = &vboxVM{}
				vms = append(vms, vm)
				vm.Name = parts[2]
			}
			if parts[1] == "UUID" {
				vm.UUID = parts[2]
			}
			if parts[1] == "State" {
				subParts := strings.Split(parts[2], "(")
				vm.State = strings.Trim(subParts[0], " ")
				if len(subParts) > 1 {
					lastStateChangeString := subParts[1][6 : len(subParts[1])-1]
					lastStatechange, err := time.Parse("2006-01-02T15:04:05.999999999", lastStateChangeString)
					if err != nil {
						log.Warn(err)
						continue
					}
					vm.LastStateChange = lastStatechange
				}
			}
		}
	}

	for _, vm := range vms {
		if vm.State != "running" {
			continue
		}
		lastChange := time.Now().Sub(vm.LastStateChange)
		if lastChange < 2*time.Hour {
			log.Infof("not removing vm '%s', running for %s (less than 2 hours)", vm.Name, lastChange)
			continue
		}

		log.Infof("powering off VM %s", vm.Name)
		cmd := exec.Command("VBoxManage", "controlvm", vm.UUID, "poweroff")
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		log.Infof("deleting VM %s", vm.Name)
		cmd = exec.Command("VBoxManage", "unregistervm", vm.UUID, "--delete")
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("reading standard input:", err)
	}

}
