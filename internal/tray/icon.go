//go:build darwin || windows

package tray

import "encoding/base64"

func templateIconBytes() []byte {
	data, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAZklEQVR4nO2UQQoAIAzD/P+n9S6ooSuC0sBOljXsYGshBJ0+TQT+EpiXuyYCMnS5tdR5jXcF6JlJ7j2B1dKTxC5XFli90ZwsQARJTip3CkgSFUm6C5XsykmuzLXPJwKVUrtMBP5nAK7ZScVmgvfoAAAAAElFTkSuQmCC")
	return data
}

func regularIconBytes() []byte {
	data, _ := base64.StdEncoding.DecodeString("AAABAAEAICAAAAEAIACfAAAAFgAAAIlQTkcNChoKAAAADUlIRFIAAAAgAAAAIAgGAAAAc3p69AAAAGZJREFUeJztlEEKACAMw/z/p/UuqKErgtLATpY17GBrIQSdPk0E/hKYl7smAjJ0ubXUeY13BeiZSe49gdXSk8QuVxZYvdGcLEAESU4qdwpIEhVJuguV7MpJrsy1zycClVK7TAT+ZwCu2UnFZoL36AAAAABJRU5ErkJggg==")
	return data
}
