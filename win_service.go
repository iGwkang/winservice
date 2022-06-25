//go:build windows
// +build windows

package winservice

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type WinService struct {
	SvcName     string
	ExecuteFunc func()

	m *mgr.Mgr
}

func (ws *WinService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	changes <- svc.Status{State: svc.StartPending}

	if ws.ExecuteFunc != nil {
		go ws.ExecuteFunc()
	}

	changes <- svc.Status{
		State:   svc.Running,
		Accepts: svc.AcceptStop | svc.AcceptShutdown,
	}

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				return
			case svc.Interrogate:
				changes <- c.CurrentStatus
			}
		}
	}
}

func (ws *WinService) Run() error {
	return svc.Run(ws.SvcName, ws)
}

func (ws *WinService) IsWindowsService() bool {
	b, err := svc.IsWindowsService()
	return b && err == nil
}

func (ws *WinService) InstallService(args ...string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	path, err := os.Executable()
	if err != nil {
		return err
	}

	s, err := m.OpenService(ws.SvcName)
	if err == nil {
		status, err := s.Query()
		if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
			return err
		}
		if status.State != svc.Stopped && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
			s.Close()
			return errors.New("WinService already installed and running")
		}
		err = s.Delete()
		if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
			return err
		}
		s.Close()

		for {
			s, err = m.OpenService(ws.SvcName)
			if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
				break
			}
			if s != nil {
				s.Close()
			}
			time.Sleep(time.Second / 3)
		}
	}
	conf := mgr.Config{
		ServiceType:  windows.SERVICE_WIN32_OWN_PROCESS,
		StartType:    mgr.StartAutomatic,
		ErrorControl: mgr.ErrorNormal,
		DisplayName:  ws.SvcName,
		Description:  ws.SvcName + " Service",
		SidType:      windows.SERVICE_SID_TYPE_UNRESTRICTED,
	}

	s, err = m.CreateService(ws.SvcName, path, conf, args...)
	if err != nil {
		return err
	}
	err = s.Start()
	s.Close()
	return nil
}

func (ws *WinService) UninstallService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(ws.SvcName)
	if err != nil {
		return err
	}
	defer s.Close()
	s.Control(svc.Stop)

	for try := 0; try < 3; try++ {
		time.Sleep(time.Second / 3)

		status, err := s.Query()
		if err != nil {
			return err
		}
		if status.ProcessId == 0 {
			break
		}
	}

	err = s.Delete()
	if err != nil && err != windows.ERROR_SERVICE_MARKED_FOR_DELETE {
		return err
	}
	return nil
}

func (ws *WinService) StartService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(ws.SvcName)
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Start()
}

func (ws *WinService) StopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(ws.SvcName)
	if err != nil {
		return err
	}
	defer s.Close()

	s.Control(svc.Stop)

	for try := 0; try < 3; try++ {
		time.Sleep(time.Second / 3)

		status, err := s.Query()
		if err != nil {
			return err
		}
		if status.ProcessId == 0 {
			break
		}
	}

	return nil
}

func (ws *WinService) Status() (svc.State, error) {
	m, err := mgr.Connect()
	if err != nil {
		return svc.Stopped, err
	}
	defer m.Disconnect()
	s, err := m.OpenService(ws.SvcName)
	if err != nil {
		return svc.Stopped, err
	}
	defer s.Close()
	stat, err := s.Query()
	return stat.State, err
}
