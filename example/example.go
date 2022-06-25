package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/iGwkang/winservice"
)

const serviceName = "test_service"

var (
	command = flag.String("o", "", "[install/uninstall/start/stop/status]")
)

func main() {
	flag.Parse()

	ws := &WinService{
		SvcName: serviceName,
		ExecuteFunc: func() {

			f, _ := os.Create("E:/test.txt")
			for {
				time.Sleep(time.Second)
				f.WriteString(*command)
				f.WriteString("\n")
			}
		},
	}

	if ws.IsWindowsService() {
		ws.Run()
	} else {
		switch *command {
		case "install":
			fmt.Println(ws.InstallService("-o", "test"))
		case "uninstall":
			fmt.Println(ws.UninstallService())
		case "start":
			fmt.Println(ws.StartService())
		case "stop":
			fmt.Println(ws.StopService())
		case "status":
			fmt.Println(ws.Status())
		default:
			ws.ExecuteFunc()
		}
	}

}
