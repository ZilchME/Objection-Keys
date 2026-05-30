//go:build windows

// Package keyboard provides global keyboard event listening on Windows using
// a low-level WH_KEYBOARD_LL hook.
package keyboard

import (
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	whKeyboardLL = 13
	hcAction     = 0

	wmKeyDown    = 0x0100
	wmSysKeyDown = 0x0104
	wmQuit       = 0x0012
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procSetWindowsHookEx   = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx     = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHook  = user32.NewProc("UnhookWindowsHookEx")
	procGetMessage         = user32.NewProc("GetMessageW")
	procPeekMessage        = user32.NewProc("PeekMessageW")
	procTranslateMessage   = user32.NewProc("TranslateMessage")
	procDispatchMessage    = user32.NewProc("DispatchMessageW")
	procPostThreadMessage  = user32.NewProc("PostThreadMessageW")
	procGetModuleHandle    = kernel32.NewProc("GetModuleHandleW")
	procGetCurrentThreadID = kernel32.NewProc("GetCurrentThreadId")

	keyboardProc = windows.NewCallback(lowLevelKeyboardProc)

	globalMu sync.RWMutex
)

// Listener provides global keyboard event listening on Windows.
type Listener struct {
	ch       chan string
	once     sync.Once
	wg       sync.WaitGroup
	hook     uintptr
	threadID uint32
}

type kbdLLHookStruct struct {
	vkCode      uint32
	scanCode    uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

type point struct {
	x int32
	y int32
}

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

var keyCodeToName = map[uint32]string{
	0x08: "backspace",
	0x09: "tab",
	0x0d: "enter",
	0x1b: "esc",
	0x20: "space",
	0x2d: "insert",
	0x2e: "backspace",
	0x24: "home",
	0x21: "pageup",
	0x23: "end",
	0x22: "pagedown",

	0x30: "0",
	0x31: "1",
	0x32: "2",
	0x33: "3",
	0x34: "4",
	0x35: "5",
	0x36: "6",
	0x37: "7",
	0x38: "8",
	0x39: "9",

	0x41: "a",
	0x42: "b",
	0x43: "c",
	0x44: "d",
	0x45: "e",
	0x46: "f",
	0x47: "g",
	0x48: "h",
	0x49: "i",
	0x4a: "j",
	0x4b: "k",
	0x4c: "l",
	0x4d: "m",
	0x4e: "n",
	0x4f: "o",
	0x50: "p",
	0x51: "q",
	0x52: "r",
	0x53: "s",
	0x54: "t",
	0x55: "u",
	0x56: "v",
	0x57: "w",
	0x58: "x",
	0x59: "y",
	0x5a: "z",

	0xba: ";",
	0xbb: "=",
	0xbc: ",",
	0xbd: "-",
	0xbe: ".",
	0xbf: "/",
	0xc0: "`",
	0xdb: "[",
	0xdc: "\\",
	0xdd: "]",
	0xde: "'",

	0x60: "0",
	0x61: "1",
	0x62: "2",
	0x63: "3",
	0x64: "4",
	0x65: "5",
	0x66: "6",
	0x67: "7",
	0x68: "8",
	0x69: "9",
	0x6a: "*",
	0x6b: "+",
	0x6d: "-",
	0x6e: ".",
	0x6f: "/",
}

var modifierKeys = map[uint32]struct{}{
	0x10: {}, // Shift
	0x11: {}, // Control
	0x12: {}, // Alt
	0x14: {}, // Caps Lock
	0x5b: {}, // Left Windows
	0x5c: {}, // Right Windows
	0xa0: {}, // Left Shift
	0xa1: {}, // Right Shift
	0xa2: {}, // Left Control
	0xa3: {}, // Right Control
	0xa4: {}, // Left Alt
	0xa5: {}, // Right Alt
}

func lowLevelKeyboardProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if nCode == hcAction && (wParam == wmKeyDown || wParam == wmSysKeyDown) {
		event := (*kbdLLHookStruct)(unsafe.Pointer(lParam))
		if _, skip := modifierKeys[event.vkCode]; !skip {
			if name, ok := keyCodeToName[event.vkCode]; ok {
				globalMu.RLock()
				listener := globalListener
				globalMu.RUnlock()
				if listener != nil {
					select {
					case listener.ch <- name:
					default:
					}
				}
			}
		}
	}

	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return ret
}

// New creates a new keyboard listener.
func New() (*Listener, error) {
	l := &Listener{
		ch: make(chan string, 512),
	}

	globalMu.Lock()
	globalListener = l
	globalMu.Unlock()

	ready := make(chan error, 1)
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		threadID, _, _ := procGetCurrentThreadID.Call()
		l.threadID = uint32(threadID)
		var m msg
		procPeekMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, 0)

		module, _, _ := procGetModuleHandle.Call(0)
		hook, _, _ := procSetWindowsHookEx.Call(
			uintptr(whKeyboardLL),
			keyboardProc,
			module,
			0,
		)
		if hook == 0 {
			globalMu.Lock()
			globalListener = nil
			globalMu.Unlock()
			ready <- errNoHook
			return
		}
		l.hook = hook
		ready <- nil

		for {
			ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
			if int32(ret) <= 0 {
				break
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
			procDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
		}

		procUnhookWindowsHook.Call(l.hook)
	}()

	if err := <-ready; err != nil {
		l.wg.Wait()
		return nil, err
	}

	return l, nil
}

// Keys returns the channel that receives key event names.
func (l *Listener) Keys() <-chan string {
	return l.ch
}

// Close shuts down the keyboard listener.
func (l *Listener) Close() {
	l.once.Do(func() {
		if l.threadID != 0 {
			procPostThreadMessage.Call(uintptr(l.threadID), uintptr(wmQuit), 0, 0)
		}
		l.wg.Wait()

		globalMu.Lock()
		if globalListener == l {
			globalListener = nil
		}
		globalMu.Unlock()

		close(l.ch)
	})
}

func HasAccessibilityPermission() bool {
	return true
}

func PromptAccessibility() bool {
	return true
}
