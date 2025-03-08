package generators

import (
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

// Source: https://stackoverflow.com/questions/23057785/how-to-copy-a-map
func copyMap(m map[string]parser.FuncInfo) map[string]parser.FuncInfo {
	cp := make(map[string]parser.FuncInfo)
	for k, v := range m {
		if !v.Public {
			// No need to copy internal methods
			continue
		}
		args := make([]parser.ArgInfo, len(v.Args))
		copy(args, v.Args)
		rets := make([]parser.ArgInfo, len(v.Return))
		copy(rets, v.Return)
		cp[k] = parser.FuncInfo{Name: v.Name, Args: args, Return: rets, Public: true}
	}

	return cp
}

func copyServiceImplInfo(siInfo *ServiceImplInfo) *ServiceImplInfo {
	new_info := ServiceImplInfo{}
	new_info.Name = siInfo.Name
	new_info.ReceiverName = siInfo.ReceiverName
	new_info.InstanceName = siInfo.InstanceName
	new_info.Methods = copyMap(siInfo.Methods)
	new_info.BaseName = siInfo.BaseName
	new_info.HasUserDefinedObjs = siInfo.HasUserDefinedObjs
	new_info.HasReturnDefinedObjs = siInfo.HasReturnDefinedObjs
	new_info.MethodBodies = make(map[string]string)
	for k, v := range siInfo.MethodBodies {
		new_info.MethodBodies[k] = v
	}
	new_info.Constructors = siInfo.Constructors
	copy(new_info.Constructors, siInfo.Constructors)
	new_info.Imports = make([]parser.ImportInfo, len(siInfo.Imports))
	copy(new_info.Imports, siInfo.Imports)
	new_info.Fields = make([]parser.ArgInfo, len(siInfo.Fields))
	copy(new_info.Fields, siInfo.Fields)
	new_info.ModifierParams = siInfo.ModifierParams
	new_info.Values = make([]string, len(siInfo.Values))
	copy(new_info.Values, siInfo.Values)
	new_info.BaseImports = make([]parser.ImportInfo, len(siInfo.BaseImports))
	copy(new_info.BaseImports, siInfo.BaseImports)
	new_info.PluginName = siInfo.PluginName
	return &new_info
}

func combineMethodInfo(funcInfo *parser.FuncInfo, prev_node *ServiceImplInfo) {
	funcInfo.Args = append(funcInfo.Args, prev_node.NextNodeMethodArgs...)
	last_return := funcInfo.Return[len(funcInfo.Return)-1]
	funcInfo.Return = append(funcInfo.Return[:len(funcInfo.Return)-1], prev_node.NextNodeMethodReturn...)
	funcInfo.Return = append(funcInfo.Return, last_return)
}
