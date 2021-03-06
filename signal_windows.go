package process

import (
	"fmt"
	"os"
	"syscall"
)

// Generates a CTRL+C signal: the calling process detatches
// itself from its own terminal, attaches to the process target
// one, then fires the CTRL+C event
func StopProcess(p *os.Process) error {
	err := FreeConsole()
	if err != nil {
		return err
	}

	err = AttachConsole(p.Pid)
	if err != nil {
		return err
	}

	err = GenerateConsoleCtrlEvent()
	if err != nil {
		return err
	}

	return nil
}

// Loads the kernel32 dll that can be used to find all sort of windows APIs
func LoadKernel32() (*syscall.DLL, error) {
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		return nil, fmt.Errorf("loadDLL: %w", e)
	}

	return d, nil
}

// Generates a CTRL+C event that spreads to all the processes attached to
// the current underlying console
func GenerateConsoleCtrlEvent() error {
	d, e := LoadKernel32()
	if e != nil {
		return e
	}

	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		return fmt.Errorf("findProc: %w", e)
	}

	r, _, e := p.Call(syscall.CTRL_C_EVENT, 0)
	if r == 0 {
		return fmt.Errorf("generateConsoleCtrlEvent: %w", e)
	}

	return nil
}

func SetConsoleCtrlHandler(flag bool) error {
	d, e := LoadKernel32()
	if e != nil {
		return e
	}

	p, e := d.FindProc("SetConsoleCtrlHandler")
	if e != nil {
		return fmt.Errorf("findProc: %w", e)
	}

	a := 0
	if flag {
		a = 1
	}

	r, _, e := p.Call(0, uintptr(a))
	if r == 0 {
		return fmt.Errorf("setConsoleCtrlHandler: %w", e)
	}

	return nil
}

// Makes the calling process not a target of the CTRL+C event
func RemoveConsoleCtrlHandler() error {
	return SetConsoleCtrlHandler(true)
}

// Reverts the effect of RemoveConsoleCtrlHandler function and makes
// the calling process a target of the CTRL+C event
func RestoreConsoleCtrlHandler() error {
	return SetConsoleCtrlHandler(false)
}

// Detatches the calling process from the underlying console
func FreeConsole() error {
	d, e := LoadKernel32()
	if e != nil {
		return e
	}

	p, e := d.FindProc("FreeConsole")
	if e != nil {
		return fmt.Errorf("findProc: %v", e)
	}

	r, _, e := p.Call(0)
	if r == 0 {
		return fmt.Errorf("freeConsole: %w", e)
	}

	return nil
}

// Creates a new console for the calling process (it must not already have one)
func AllocConsole() error {
	d, e := LoadKernel32()
	if e != nil {
		return e
	}

	p, e := d.FindProc("AllocConsole")
	if e != nil {
		return fmt.Errorf("findProc: %v", e)
	}

	r, _, e := p.Call(0)
	if r == 0 {
		return fmt.Errorf("allocConsole: %w", e)
	}

	return nil
}

// Attaches to the same console as the process with the given PID (it must not already have one)
func AttachConsole(pid int) error {
	d, e := LoadKernel32()
	if e != nil {
		return e
	}

	p, e := d.FindProc("AttachConsole")
	if e != nil {
		return fmt.Errorf("findProc: %v", e)
	}

	r, _, e := p.Call(uintptr(pid))
	if r == 0 {
		return fmt.Errorf("attachConsole: %w", e)
	}

	return nil
}

// Makes che calling process immune to CTRL+C events and generates one of them.
// You may call RestoreConsoleCtrlHandler to revert the change after all child
// processes have exited
func SendCtrlC() error {
	if e := RemoveConsoleCtrlHandler(); e != nil {
		return e
	}

	if e := GenerateConsoleCtrlEvent(); e != nil {
		return e
	}

	return nil
}
