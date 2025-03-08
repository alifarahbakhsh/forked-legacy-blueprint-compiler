package netgen

import (
	"errors"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type RemoteTypeInfo struct {
	Name    string
	Val     string
	Fields  []FieldInfo
	IsEnum  bool
	PkgPath string
}

type MethodInfo struct {
	Name string
	Val  string
}

type ResponseInfo struct {
	Name   string
	Val    string
	Fields []FieldInfo
}

type RequestInfo struct {
	Name   string
	Val    string
	Fields []FieldInfo
}

type FieldInfo struct {
	Name string
	Type parser.TypeInfo
}

type ServiceInfo struct {
	Name    string
	Methods []MethodInfo
}

type NetworkGeneratorFactory struct {
	Generators     map[string]NetworkGenerator
	GeneratorFuncs map[string]func() NetworkGenerator
}

var factory *NetworkGeneratorFactory

func GetNetGenFactory() *NetworkGeneratorFactory {
	if factory == nil {
		gen_funcs := make(map[string]func() NetworkGenerator)
		gen_funcs["aiothrift"] = NewThriftGenerator
		gen_funcs["default"] = NewDefaultWebGenerator
		gen_funcs["grpc"] = NewGRPCGenerator
		factory = &NetworkGeneratorFactory{Generators: make(map[string]NetworkGenerator), GeneratorFuncs: gen_funcs}
	}
	return factory
}

func (nf *NetworkGeneratorFactory) GetGenerator(framework string) (NetworkGenerator, error) {
	if gen, ok := nf.Generators[framework]; !ok {
		if gen_func, ok2 := nf.GeneratorFuncs[framework]; ok2 {
			generator := gen_func()
			nf.Generators[framework] = generator
			return generator, nil
		} else {
			return nil, errors.New("Framework " + framework + " was not registered as a valid NetworkGenerator")
		}
	} else {
		return gen, nil
	}
	return nil, nil
}
