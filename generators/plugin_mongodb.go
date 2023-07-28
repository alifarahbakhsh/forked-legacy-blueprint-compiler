package generators

import (
	"os"
	"path"
	"strconv"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/generators/deploy"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type ReplicationHandler struct {
	Names        map[string][]string
	PrimaryNames map[string]string
	Hosts        map[string]string
}

var replHandler *ReplicationHandler
var generated bool

func getReplicationHandler() *ReplicationHandler {
	if replHandler == nil {
		replHandler = &ReplicationHandler{Names: make(map[string][]string), PrimaryNames: make(map[string]string), Hosts: make(map[string]string)}
	}
	return replHandler
}

type MongoScriptGenerator struct{}

func (s *MongoScriptGenerator) Generate(out_dir string) error {
	if generated {
		return nil
	}
	for name, members := range replHandler.Names {
		filename := path.Join(out_dir, "rs-init-"+name+".sh")
		outf, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer outf.Close()
		primary_name := replHandler.PrimaryNames[name]
		script_body := ""
		script_body += "#!/bin/bash\n\nDELAY=25\n\n"
		script_body += "mongosh <<EOF\n"
		script_body += "var config={\n"
		script_body += "\t\"_id\": \"" + name + "\",\n"
		script_body += "\t\"version\": 1,\n"
		script_body += "\t\"members\": [\n"
		for i, member := range members {
			host := replHandler.Hosts[member]
			prefix := "\t\t"
			script_body += prefix + "{\n"
			script_body += prefix + "\t\"_id\":" + strconv.Itoa(i+1) + ",\n"
			script_body += prefix + "\t\"host\": \"" + host + "\",\n"
			script_body += prefix + "\t\"priority\": "
			if member == primary_name {
				script_body += "2,\n"
			} else {
				script_body += "1,\n"
			}
			if i != len(members)-1 {
				script_body += prefix + "},\n"
			} else {
				script_body += prefix + "}\n"
			}
		}
		script_body += "\t]\n"
		script_body += "};\n"
		script_body += "rs.initiate(config, {force: true});\nEOF\n\n"
		script_body += "echo \"****** Waiting for ${DELAY} seconds for replicaset configuration to be applied ****** \"\n\n"
		script_body += "sleep $DELAY"
		_, err = outf.WriteString(script_body)
		if err != nil {
			return err
		}
	}
	generated = true
	return nil
}

type MongoDBNode struct {
	Name            string
	TypeName        string
	ReplHandler     *ReplicationHandler
	IsReplicated    bool
	IsPrimary       bool
	ReplicaSetName  string
	ScriptGenerator *MongoScriptGenerator
	Params          []Parameter
	ClientModifiers []Modifier
	ServerModifiers []Modifier
	ASTNodes        []*ServiceImplInfo
	DepInfo         *deploy.DeployInfo
}

func (n *MongoDBNode) Accept(v Visitor) {
	v.VisitMongoDBNode(v, n)
}

func (n *MongoDBNode) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	for _, child := range n.ClientModifiers {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	for _, child := range n.ServerModifiers {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func GenerateMongoDBNode(node parser.DetailNode) Node {
	var params []Parameter
	var cmodifiers []Modifier
	var smodifiers []Modifier
	is_primary := false
	replica_set_value := ""
	is_replicated := false
	for _, arg := range node.Arguments {
		param := convert_argument_node(arg)
		switch ptype := param.(type) {
		case *ValueParameter:
			if ptype.KeywordName == "replica_set" {
				replica_set_value = ptype.Value
				is_replicated = true
			} else if ptype.KeywordName == "is_primary" {
				is_primary, _ = strconv.ParseBool(ptype.Value)
			} else {
				params = append(params, param)
			}
		case *InstanceParameter:
			params = append(params, param)
		}
	}
	for _, modifier := range node.ClientModifiers {
		cmodifiers = append(cmodifiers, convert_modifier_node(modifier))
	}
	for _, modifier := range node.ServerModifiers {
		smodifiers = append(smodifiers, convert_modifier_node(modifier))
	}

	handler := getReplicationHandler()
	if is_replicated {
		if _, ok := replHandler.Names[replica_set_value]; !ok {
			replHandler.Names[replica_set_value] = []string{}
		}
		replHandler.Names[replica_set_value] = append(replHandler.Names[replica_set_value], node.Name)
		if is_primary {
			replHandler.PrimaryNames[replica_set_value] = node.Name
		}
	}

	return &MongoDBNode{Name: node.Name, TypeName: "MongoDB", Params: params, ClientModifiers: cmodifiers, ServerModifiers: smodifiers, DepInfo: deploy.NewDeployInfo(), IsReplicated: is_replicated, IsPrimary: is_primary, ReplicaSetName: replica_set_value, ReplHandler: handler, ScriptGenerator: &MongoScriptGenerator{}}
}

func (n *MongoDBNode) getConstructorBody(info *parser.ImplInfo) string {
	body := ""
	body += "addr := os.Getenv(\"" + n.Name + "_ADDRESS\")\n"
	body += "port := os.Getenv(\"" + n.Name + "_PORT\")\n"
	body += "int_db := nosqldb.GetMongo(addr, port)\n"
	body += "return &" + n.Name + "{internal: int_db}\n"
	return body
}

func (n *MongoDBNode) GenerateClientNode(info *parser.ImplInfo) {
	methods := copyMap(info.Methods)
	con_name := "New" + n.Name
	con_args := []parser.ArgInfo{}
	con_rets := []parser.ArgInfo{parser.GetPointerArg("", n.Name)}
	constructor := parser.FuncInfo{Name: con_name, Args: con_args, Return: con_rets}
	imports := []parser.ImportInfo{parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib/choices/nosqldb"}, parser.ImportInfo{ImportName: "", FullName: "os"}, parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib/components"}}
	fields := []parser.ArgInfo{parser.GetPointerArg("internal", "nosqldb."+n.TypeName)}
	bodies := make(map[string]string)
	bodies[con_name] = n.getConstructorBody(info)
	for name, method := range methods {
		var arg_names []string
		for _, arg := range method.Args {
			arg_names = append(arg_names, arg.Name)
		}
		bodies[name] = "return c.internal." + name + "(" + strings.Join(arg_names, ", ") + ")\n"
	}
	client_node := &ServiceImplInfo{Name: n.Name, ReceiverName: "c", Methods: methods, Constructors: []parser.FuncInfo{constructor}, Imports: imports, Fields: fields, MethodBodies: bodies, PluginName: "MongoDB"}
	n.ASTNodes = append(n.ASTNodes, client_node)
}
