package netgen

import "gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"

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
	SetCustomParameters(params map[string]string)
}
