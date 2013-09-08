package ein

import (
    "fmt"
)

const DEBUG = false

func debugf(format string, args ...interface{}) {
    if !DEBUG {
        return
    }
    fmt.Printf(format, args...)
}
