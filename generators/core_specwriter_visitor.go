package generators

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/otiai10/copy"
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type SpecSourceWriterVisitor struct {
	DefaultVisitor
	logger      *log.Logger
	curDir      string
	appName     string
	pkgName     string
	specDir     string
	remoteTypes map[string]*parser.ImplInfo
	services    map[string]*parser.ServiceInfo
	pathpkgs    map[string]string
}

func NewSpecSourceWriterVisitor(logger *log.Logger, outDir string, appName string, specDir string, remoteTypes map[string]*parser.ImplInfo, services map[string]*parser.ServiceInfo, pkgs map[string]string) *SpecSourceWriterVisitor {
	return &SpecSourceWriterVisitor{logger: logger, curDir: outDir, appName: appName, pkgName: "", specDir: specDir, remoteTypes: remoteTypes, services: services, pathpkgs: pkgs}
}

func (v *SpecSourceWriterVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.logger.Println("Starting SpecSourceWriter visit")
	// Copy over the input directory in to the spec package
	specDir := path.Join(v.curDir, "spec")
	err := os.MkdirAll(specDir, 0755)
	if err != nil {
		v.logger.Fatal(err)
	}
	err = copy.Copy(v.specDir, specDir)
	if err != nil {
		v.logger.Fatal(err)
	}
	v.DefaultVisitor.VisitMillenialNode(v, n)
	v.logger.Println("Ending SpecSourceWriter visit")
}

func (v *SpecSourceWriterVisitor) VisitDockerContainerNode(_ Visitor, n *DockerContainerNode) {
	oldPath := v.curDir
	new_dir := path.Join(oldPath, strings.ToLower(n.Name))
	err := os.MkdirAll(new_dir, 0755)
	if err != nil {
		v.logger.Fatal(err)
	}
	v.curDir = new_dir
	v.DefaultVisitor.VisitDockerContainerNode(v, n)
	v.curDir = oldPath
}

func (v *SpecSourceWriterVisitor) VisitProcessNode(_ Visitor, n *ProcessNode) {
	oldPath := v.curDir
	new_dir := path.Join(oldPath, strings.ToLower(n.Name))
	err := os.MkdirAll(new_dir, 0755)
	if err != nil {
		v.logger.Fatal(err)
	}
	v.curDir = new_dir
	v.pkgName = strings.ToLower(n.Name)
	v.DefaultVisitor.VisitProcessNode(v, n)
	v.pkgName = ""
	v.curDir = oldPath
}
