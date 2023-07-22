package netgen

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

func generateMetricMethod(handler_name string, service_name string, func_names []string) (parser.FuncInfo, string) {
	method := parser.FuncInfo{}
	method.Name = "startMetrics"
	method.Public = false
	var body string
	body += "ticker := time.NewTicker(1 * time.Second)\n"
	body += "for {\n"
	body += "\tselect {\n"
	body += "\tcase <-ticker.C:\n"
	prefix := "\t\t"
	for _, fn_name := range func_names {
		body += prefix + fmt.Sprintf("avg_latency_%s := float64(%s) / float64(%s)\n", fn_name, handler_name+".sumLatency_"+fn_name, handler_name+".numReqs_"+fn_name)
		body += prefix + fmt.Sprintf("debug.ReportMetric(\"%s\", %s)\n", "avg_latency_"+fn_name, "avg_latency_"+fn_name)
		body += prefix + fmt.Sprintf("debug.ReportMetric(\"%s\", %s)\n", "num_reqs_"+fn_name, handler_name+".numReqs_"+fn_name)
		// Reset the variables
		body += prefix + handler_name + ".numReqs_" + fn_name + " = 0\n"
		body += prefix + handler_name + ".sumLatency_" + fn_name + " = 0\n"
	}
	body += "\t}\n"
	body += "}\n"
	return method, body
}

func generateFunctionWrapperBody(handler_name string, func_name string) string {
	var body string
	body += handler_name + ".numReqs_" + func_name + " += 1\n"
	body += "start_duration := time.Now()\n"
	body += "defer func() {\n"
	body += "\tend_duration := time.Now().Sub(start_duration)\n"
	body += "\t" + handler_name + ".sumLatency_" + func_name + " += int64(end_duration / time.Nanosecond)\n"
	body += "}()\n"
	return body
}

func generateMetricFields(func_names []string) []parser.ArgInfo {
	var args []parser.ArgInfo
	for _, fn_name := range func_names {
		args = append(args, parser.GetBasicArg("numReqs_"+fn_name, "int64"))
		args = append(args, parser.GetBasicArg("sumLatency_"+fn_name, "int64"))
	}
	return args
}

func generateMetricImports() []parser.ImportInfo {
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/stdlib/debug"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "time"})
	return imports
}

func generateMetricConstructorBody(handler_name string) string {
	return "go " + handler_name + ".startMetrics()\n"
}
