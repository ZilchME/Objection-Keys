package keyboard

import "errors"

var (
	errNoEventTap = errors.New("failed to create CGEventTap, check Accessibility permission")
	errNoRunLoop  = errors.New("failed to create run loop source")
	errNoHook     = errors.New("failed to install keyboard hook")
)
