package ein

type environment struct {
	contents map[string]interface{}
	parent   *environment
}

func newEnvironment() *environment {
    return &environment{contents: make(map[string]interface{}, 8), parent: nil}
}

func (env *environment) get(name string) (interface{}, bool) {
    v, ok := env.contents[name]
    if ok {
        return v, true
    }
    if env.parent == nil {
        return nil, false
    }
    return env.parent.get(name)
}
