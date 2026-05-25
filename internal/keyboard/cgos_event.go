//go:build darwin

// Package keyboard provides global keyboard event listening on macOS using CoreGraphics CGEventTap.
//
// This package requires Accessibility permission. The app must be added to:
// System Settings → Privacy & Security → Accessibility
package keyboard

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework CoreGraphics -framework Foundation -framework Carbon
// #import <CoreGraphics/CoreGraphics.h>
// #import <ApplicationServices/ApplicationServices.h>
//
// static CFMachPortRef g_tap = NULL;
// static CFRunLoopRef g_run_loop = NULL;
//
// extern void onKeyEvent(uint16_t keyCode, uint8_t isKeyDown);
//
// static CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type,
//                              CGEventRef event, void *refcon) {
//     if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
//         if (g_tap) CGEventTapEnable(g_tap, true);
//         return event;
//     }
//     if (type != kCGEventKeyDown && type != kCGEventKeyUp) return event;
//
//     CGKeyCode keyCode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
//     uint8_t isKeyDown = (type == kCGEventKeyDown) ? 1 : 0;
//     onKeyEvent(keyCode, isKeyDown);
//     return event;
// }
//
// static int setupEventTap() {
//     CGEventMask eventMask = (1 << kCGEventKeyDown) | (1 << kCGEventKeyUp);
//     CFMachPortRef tap = CGEventTapCreate(
//         kCGSessionEventTap,
//         kCGHeadInsertEventTap,
//         kCGEventTapOptionListenOnly,
//         eventMask,
//         eventTapCallback,
//         NULL
//     );
//     if (!tap) return -1;
//     g_tap = tap;
//     return 0;
// }
//
// static int startEventTap() {
//     if (!g_tap) return -1;
//     CFRunLoopSourceRef source = CFMachPortCreateRunLoopSource(NULL, g_tap, 0);
//     if (!source) return -1;
//     g_run_loop = CFRunLoopGetCurrent();
//     CFRetain(g_run_loop);
//     CFRunLoopAddSource(g_run_loop, source, kCFRunLoopCommonModes);
//     CGEventTapEnable(g_tap, true);
//     CFRelease(source);
//     return 0;
// }
//
// static void stopEventTap() {
//     if (g_tap) {
//         CFMachPortInvalidate(g_tap);
//         g_tap = NULL;
//     }
//     if (g_run_loop) {
//         CFRelease(g_run_loop);
//         g_run_loop = NULL;
//     }
// }
//
// static int checkAccessibility() {
//     return AXIsProcessTrusted() ? 0 : -1;
// }
//
// static int promptAccessibility() {
//     const void *keys[] = { kAXTrustedCheckOptionPrompt };
//     const void *values[] = { kCFBooleanTrue };
//     CFDictionaryRef options = CFDictionaryCreate(
//         kCFAllocatorDefault,
//         keys,
//         values,
//         1,
//         NULL,
//         NULL
//     );
//     int result = AXIsProcessTrustedWithOptions(options) ? 0 : -1;
//     CFRelease(options);
//     return result;
// }
//
// static void runLoopStart() {
//     CFRunLoopRun();
// }
//
// static void runLoopStop() {
//     if (g_run_loop) {
//         CFRunLoopStop(g_run_loop);
//     }
// }
import "C"

import (
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// CGKeyCode to key name mapping.
// Based on Carbon HIToolbox KeyboardEventTypes.h (USB HID Usage Table).
var keyCodeToName = map[uint16]string{
	0x00: "a",
	0x01: "s",
	0x02: "d",
	0x03: "f",
	0x04: "h",
	0x05: "g",
	0x06: "z",
	0x07: "x",
	0x08: "c",
	0x09: "v",
	0x0b: "b",
	0x0c: "q",
	0x0d: "w",
	0x0e: "e",
	0x0f: "r",
	0x10: "y",
	0x11: "t",
	0x12: "1",
	0x13: "2",
	0x14: "3",
	0x15: "4",
	0x16: "6",
	0x17: "5",
	0x18: "=",
	0x19: "9",
	0x1a: "7",
	0x1b: "-",
	0x1c: "8",
	0x1d: "0",
	0x1e: "]",
	0x1f: "o",
	0x20: "u",
	0x21: "[",
	0x22: "i",
	0x23: "p",
	0x24: "\n", // Return
	0x25: "l",
	0x26: "j",
	0x27: "'",
	0x28: "k",
	0x29: ";",
	0x2a: "\\",
	0x2b: ",",
	0x2c: "/",
	0x2d: "n",
	0x2e: "m",
	0x2f: ".",
	0x30: "tab",
	0x31: " ",
	0x32: "`",
	0x33: "\x7f", // Backspace
	0x35: "\x1b", // Escape
	0x4c: "\n",   // Keypad Enter
	0x75: "delete",

	0x72: "insert",
	0x73: "home",
	0x74: "pageup",
	0x77: "end",
	0x79: "pagedown",

	// Modifier keys
	0x39: "capslock",
	0x38: "leftshift",
	0x3b: "leftcontrol",
	0x3a: "leftalt",
	0x37: "leftcommand",
	0x3c: "rightshift",
	0x3d: "rightalt",
	0x3e: "rightcontrol",
	0x36: "rightcommand",
	0x7a: "f1",
	0x78: "f2",
	0x63: "f3",
	0x76: "f4",
	0x60: "f5",
	0x61: "f6",
	0x62: "f7",
	0x64: "f8",
	0x65: "f9",
	0x6d: "f10",
	0x67: "f11",
	0x6f: "f12",
	0x69: "f13",
	0x6b: "f14",
	0x71: "f15",
	0x6a: "f16",
	0x40: "f17",
	0x4f: "f18",
	0x50: "f19",
	0x5a: "f20",
}

const (
	_ = "skip"
)

// globalListener is the global Listener instance for C callbacks.
var globalListener *Listener

// Listener provides global keyboard event listening on macOS.
type Listener struct {
	ch   chan string
	once sync.Once
}

//export onKeyEvent
func onKeyEvent(cgKeyCode C.uint16_t, isKeyDown C.uint8_t) {
	keyCode := uint16(cgKeyCode)
	raw, ok := keyCodeToName[keyCode]
	if !ok {
		return // Ignore unmapped keys
	}

	// Skip modifier keys
	if skipModifiers(raw) {
		return
	}

	// Only process key-down events
	if isKeyDown != 1 {
		return
	}

	name := normalizeName(raw)

	// Get the global listener and send the key event
	if globalListener != nil {
		select {
		case globalListener.ch <- name:
		default:
			// Channel full, drop the event to avoid blocking
		}
	}
}

// New creates a new keyboard listener.
// Requires Accessibility permission (System Settings → Privacy & Security → Accessibility).
func New() (*Listener, error) {
	if C.setupEventTap() != 0 {
		return nil, errNoEventTap
	}

	l := &Listener{
		ch: make(chan string, 512),
	}

	// Store global reference for the C callback
	globalListener = l

	ready := make(chan int, 1)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if C.startEventTap() != 0 {
			ready <- -1
			return
		}

		ready <- 0
		C.runLoopStart()
	}()

	if <-ready != 0 {
		C.stopEventTap()
		globalListener = nil
		return nil, errNoRunLoop
	}

	return l, nil
}

// stopRunLoop stops the background run loop.
func stopRunLoop() {
	C.runLoopStop()
}

// Keys returns the channel that receives key event names.
// Only key-down events are sent.
func (l *Listener) Keys() <-chan string {
	return l.ch
}

// Close shuts down the keyboard listener.
func (l *Listener) Close() {
	l.once.Do(func() {
		stopRunLoop()
		C.stopEventTap()
		globalListener = nil
		close(l.ch)
	})
}

// skipModifiers returns true if the key name is a modifier key
// that should be ignored.
func skipModifiers(name string) bool {
	return name == "leftcontrol" || name == "rightcontrol" ||
		name == "leftshift" || name == "rightshift" ||
		name == "leftalt" || name == "rightalt" ||
		name == "leftcommand" || name == "rightcommand" ||
		name == "capslock"
}

// normalizeName converts macOS CGKeyCode names to our internal key names.
func normalizeName(raw string) string {
	switch raw {
	case "\n":
		return "enter"
	case "\x1b":
		return "esc"
	case "\x7f":
		return "backspace"
	case " ":
		return "space"
	case "\x09":
		return "tab"
	case "delete":
		return "backspace"
	default:
		return raw
	}
}

// HasAccessibilityPermission checks if the app has macOS Accessibility permission.
func HasAccessibilityPermission() bool {
	return C.checkAccessibility() == 0
}

// PromptAccessibility opens a system dialog requesting Accessibility permission and
// waits for the user to complete the authorization process.
// Returns true if permission was granted within the wait period.
func PromptAccessibility() bool {
	// Pop up the system authorization dialog
	C.promptAccessibility()

	// Loop until timeout, letting the user respond to the system prompt.
	// macOS does not block on AXIsProcessTrustedWithOptions, so we poll.
	for i := 0; i < 15; i++ {
		time.Sleep(time.Second)
		if HasAccessibilityPermission() {
			return true
		}
	}
	return false
}

// export onKeyEvent to satisfy cgo
var _ = unsafe.Sizeof(C.int(0))
