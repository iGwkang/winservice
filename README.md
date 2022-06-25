# winservice
Go implements simple and easy-to-use windows services

# How to use

`go get "github.com/iGwkang/winservice"`

```go
import "github.com/iGwkang/winservice"

var (
	command = flag.String("o", "", "[install/uninstall/start/stop/status]")
)

func main() {
	flag.Parse()

	ws := &winservice.WinService{
		SvcName: "Your Service Name",
		ExecuteFunc: func() {
			// service run function
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
```
