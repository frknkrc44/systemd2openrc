/*
   Systemd2OpenRC - Convert systemd calls to openrc
   Copyright (C) 2025  Furkan Karcıoğlu

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func findNoOpArg() bool {
	for i, arg := range os.Args {
		if i < 1 {
			continue
		}

		if strings.Index(arg, "-") != 0 {
			return true
		}
	}

	return false
}

func main() {
	commandType := -1
	cmds := []string{}
	userMode := false
	serviceName := ""

	if len(os.Args) < 2 || !findNoOpArg() {
		newArgs := []string{os.Args[0]}
		newArgs = append(newArgs, "list-units")
		newArgs = append(newArgs, os.Args[1:]...)
		os.Args = newArgs
	}

	for i, arg := range os.Args {
		switch arg {
		case "start", "stop", "restart", "status":
			commandType = 0
			cmds = append(cmds, arg)
		case "try-restart", "reload-or-restart", "try-reload-or-restart":
			commandType = 0
			cmds = append(cmds, "restart")
		case "--user":
			userMode = true
			if commandType < 0 {
				commandType = 0
			}
		case "list-units", "list-unit-files":
			commandType = 1
			cmds = append(cmds, "show")
		case "--type=service":
			if commandType == 1 {
				cmds = append(cmds, "-v")
			}
		case "enable", "reenable":
			commandType = 1
			cmds = append(cmds, "add")
		case "disable":
			commandType = 1
			cmds = append(cmds, "del")
		case "mask":
			commandType = 2
			cmds = append(cmds, "444")
		case "unmask":
			commandType = 2
			cmds = append(cmds, "555")
		case "daemon-reload", "kill", "list-automounts", 
			"list-paths", "list-sockets", "list-timers", 
			"is-active", "is-failed", "cat", 
			"list-dependencies", "reload", "isolate",
			"clean", "freeze", "thaw",
			"set-property", "bind", "mount-image",
			"service-log-level", "service-log-target",
			"reset-failed", "whoami", "preset",
			"preset-all", "is-enabled", "link",
			"revert", "add-wants", "edit",
			"get-default", "set-default", "list-machines",
			"list-jobs", "cancel", "show-environment",
			"set-environment", "unset-environment",
			"import-environment", "daemon-reexec",
			"log-level", "log-target", "service-watchdogs",
			"default", "kexec", "soft-reboot",
			"exit", "switch-root", "--type=socket":
			fmt.Printf("Not implemented: %s\n", arg)
			os.Exit(0)
		case "rescue", "emergency":
			commandType = 4
			cmds = append(cmds, "/sbin/openrc", "single")
		case "halt", "shutdown":
			commandType = 3
			cmds = append(cmds, "poweroff")
		case "poweroff", "reboot", "hibernate", "hybrid-sleep", "suspend-then-hibernate":
			commandType = 3
			cmds = append(cmds, arg)
		case "sleep", "suspend":
			commandType = 3
			cmds = append(cmds, "suspend")
		case "is-system-running":
			fmt.Println("running")
		case "help", "version":
			fmt.Println("SystemD2OpenRC v1.0.0")
			fmt.Println("by frknkrc44")
			fmt.Println("Use the man documentation, a search engine etc. to find help.")
		default:
			if i != 0 && serviceName == "" && strings.Index(arg, "-") != 0 {
				serviceName = arg
			}
		}
	}

	var finalCmd *[]string = nil

	switch commandType {
	case 0:
		rcServiceCmd := []string{"/bin/rc-service"}

		if userMode {
			rcServiceCmd = append(rcServiceCmd, "--user")
		}

		rcServiceCmd = append(rcServiceCmd, serviceName)
		rcServiceCmd = append(rcServiceCmd, cmds...)
		finalCmd = &rcServiceCmd
	case 1:
		rcUpdateCmd := []string{"/bin/rc-update"}

		if userMode {
			rcUpdateCmd = append(rcUpdateCmd, "--user")
		}

		rcUpdateCmd = append(rcUpdateCmd, cmds...)

		if len(serviceName) > 0 {
			rcUpdateCmd = append(rcUpdateCmd, serviceName)
		}

		finalCmd = &rcUpdateCmd
	case 2:
		chmodCmd := []string{"/bin/chmod"}
		chmodCmd = append(chmodCmd, cmds...)
		chmodCmd = append(chmodCmd, fmt.Sprintf("/etc/init.d/%s", serviceName))
		finalCmd = &chmodCmd
	case 3:
		loginctlCmd := []string{"/bin/loginctl"}
		loginctlCmd = append(loginctlCmd, cmds...)
		finalCmd = &loginctlCmd
	case 4:
		finalCmd = &cmds
	}

	if finalCmd != nil {
		proc, err := os.StartProcess((*finalCmd)[0], *finalCmd, &os.ProcAttr{
			Dir: ".",
			Env: os.Environ(),
			Files: []*os.File{
				os.Stdin,
				os.Stdout,
				os.Stderr,
			},
		})
		if err != nil {
			log.Fatalln(err)
		}

		state, err := proc.Wait()
		if err != nil {
			log.Fatalln(err)
		}

		os.Exit(state.ExitCode())
	}
}
