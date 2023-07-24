package netgen

import (
	"errors"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
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

type NetworkGenerator interface {
	GenerateServerMethods(handler_name string, service_name string, methods map[string]parser.FuncInfo, is_metrics_on bool, instance_name string) (map[string]string, error)
	GenerateClientMethods(handler_name string, service_name string, methods map[string]parser.FuncInfo, nextNodeMethodArgs []parser.ArgInfo, nextNodeMethodReturn []parser.ArgInfo, is_metrics_on bool, has_timeout bool) (map[string]string, error)
	SetAppName(appName string)
	GenerateFiles(outdir string) error
	ConvertRemoteTypes(remoteTypes map[string]*parser.ImplInfo) error
	ConvertEnumTypes(enums map[string]*parser.EnumInfo) error
	GetImports(hasUserDefinedObjs bool) []parser.ImportInfo
	GenerateServerConstructor(prev_handler string, service_name string, handler_name string, base_name string, is_metrics_on bool) (parser.FuncInfo, string, []parser.ImportInfo, []parser.ArgInfo, []parser.StructInfo)
	GenerateClientConstructor(service_name string, handler_name string, base_name string, is_metrics_on bool, timeout string) (parser.FuncInfo, string, []parser.ImportInfo, []parser.ArgInfo, []parser.StructInfo)
	GetRequirements() []parser.RequireInfo
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
