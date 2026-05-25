//go:build darwin

package sound

import "time"

func nowNano() int64 {
	return time.Now().UnixNano()
}
