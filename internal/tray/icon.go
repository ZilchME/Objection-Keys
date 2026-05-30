//go:build darwin

package tray

import "encoding/base64"

func iconBytes() []byte {
	data, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAZklEQVR4nO2UQQoAIAzD/P+n9S6ooSuC0sBOljXsYGshBJ0+TQT+EpiXuyYCMnS5tdR5jXcF6JlJ7j2B1dKTxC5XFli90ZwsQARJTip3CkgSFUm6C5XsykmuzLXPJwKVUrtMBP5nAK7ZScVmgvfoAAAAAElFTkSuQmCC")
	return data
}
