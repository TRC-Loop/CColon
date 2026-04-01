package stdlib

import "github.com/TRC-Loop/ccolon/vm"

type Registry struct {
	modules map[string]*vm.ModuleValue
}

func NewRegistry() *Registry {
	r := &Registry{modules: make(map[string]*vm.ModuleValue)}
	r.Register("console", NewConsoleModule())
	r.Register("math", NewMathModule())
	r.Register("random", NewRandomModule())
	r.Register("json", NewJsonModule())
	r.Register("fs", NewFsModule())
	return r
}

func (r *Registry) Register(name string, mod *vm.ModuleValue) {
	r.modules[name] = mod
}

func (r *Registry) Get(name string) (*vm.ModuleValue, bool) {
	mod, ok := r.modules[name]
	return mod, ok
}

func (r *Registry) RegisterAll(machine *vm.VM) {
	for name, mod := range r.modules {
		machine.RegisterModule(name, mod)
	}
}
