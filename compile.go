package ein

import (
    "io"
)

func Compile(source io.Reader) (node, error) {
    debugf("compiling...\n")
    root, err := parse(source)
    if err != nil {
        return nil, err
    }
    debugf("done compiling\n")
    return root, nil
}
