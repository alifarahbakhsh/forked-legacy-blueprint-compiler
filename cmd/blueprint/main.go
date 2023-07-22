package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/generators"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/generators/deploy"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"

	progressbar "github.com/schollz/progressbar/v3"
)

func main() {
	configPtr := flag.String("config", "", "Path to the configuration file")
	verbosePtr := flag.Bool("verbose", false, "Print log output")

	flag.Parse()

	configFile := *configPtr
	if configFile == "" {
		log.Fatal("Usage: go run main.go -config=<path to config.json>")
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	if !(*verbosePtr) {
		log.SetOutput(ioutil.Discard)
		logger.SetOutput(ioutil.Discard)
	}

	bar := progressbar.Default(17)

	config, err := parser.ParseConfig(configFile)
	if err != nil {
		logger.Fatal(err)
	}

	specParser := parser.NewSpecParser(config, logger)
	specParser.ParseSpec()
	bar.Add(1)

	wiringParser := parser.NewWiringParser(config, logger)
	wiringParser.ParseWiring()
	bar.Add(1)

	modregistry := generators.InitModifierRegistry(logger)
	bar.Add(1)

	generator := generators.NewGenerator(config, logger, specParser.Implementations, modregistry)
	generator.ConvertSerializedRep(wiringParser.RootNode)
	bar.Add(1)

	printVisitor := generators.NewPrintVisitor(logger)
	generator.RootNode.Accept(printVisitor)
	printVisitor.Print()
	bar.Add(1)

	depGraphVisitor := generators.NewDependencyGraphVisitor(logger, specParser.Implementations, specParser.Services)
	generator.RootNode.Accept(depGraphVisitor)
	depGraph := depGraphVisitor.DepGraph
	logger.Println("Dependency graph is as follows: \n" + depGraph.String())
	depGraph.TopoSort()
	bar.Add(1)

	// Apply source code modifiers + generate network layer node files
	fixRemoteTypeVisitor := generators.NewRemoteTypeVisitor(logger, specParser.RemoteTypes, specParser.PathPkgs, specParser.Implementations)
	generator.RootNode.Accept(fixRemoteTypeVisitor)
	bar.Add(1)

	clientCollectorVisitor := generators.NewClientCollectorVisitor(logger, specParser.Implementations, specParser.PathPkgs, config.SrcDir, specParser.RemoteTypes, specParser.Services)
	generator.RootNode.Accept(clientCollectorVisitor)
	bar.Add(1)

	generateVisitor := generators.NewGenerateSourceCodeVisitor(logger, modregistry, config.AppName, config.OutDir, specParser.RemoteTypes, clientCollectorVisitor.DefaultClientInfos, specParser.PathPkgs, specParser.Implementations, config.SrcDir, specParser.Enums)
	generator.RootNode.Accept(generateVisitor)
	bar.Add(1)

	generateClientVisitor := generators.NewGenerateClientSourceCodeVisitor(logger, modregistry, config.AppName, config.OutDir, specParser.RemoteTypes, clientCollectorVisitor.DefaultClientInfos, specParser.PathPkgs, specParser.Implementations, config.SrcDir, specParser.Enums, specParser.Services, depGraph)
	generator.RootNode.Accept(generateClientVisitor)
	bar.Add(1)

	// Addr + Port Resolution
	portAuthority := deploy.NewPortAuthority(logger)
	basicDeployVisitor := generators.NewBasicDeployVisitor(logger, config, portAuthority)
	generator.RootNode.Accept(basicDeployVisitor)
	bar.Add(1)

	// Apply deployment modifiers
	depModVisitor := generators.NewDeployModifierVisitor(logger, modregistry)
	generator.RootNode.Accept(depModVisitor)
	bar.Add(1)

	// Collect addresses and ports
	addrCollectorVisitor := generators.NewAddrCollectorVisitor(logger)
	generator.RootNode.Accept(addrCollectorVisitor)
	bar.Add(1)

	specWriterVisitor := generators.NewSpecSourceWriterVisitor(logger, config.OutDir, config.AppName, config.SrcDir, specParser.RemoteTypes, specParser.Services, specParser.PathPkgs)
	generator.RootNode.Accept(specWriterVisitor)
	bar.Add(1)

	localServicesInfoCollectorVisitor := generators.NewLocalServicesInfoCollectorVisitor(logger)
	generator.RootNode.Accept(localServicesInfoCollectorVisitor)
	bar.Add(1)

	depgenfactory := deploy.GetDepGenFactory()
	// Generate main functions, run scripts, container config files
	mainVisitor := generators.NewMainVisitor(logger, config.OutDir, specParser.PathPkgs, specParser.Implementations, config.SrcDir, depgenfactory, addrCollectorVisitor.Addrs, config.Inventory, generateVisitor.Frameworks, depGraph, localServicesInfoCollectorVisitor.LocalServiceInfos)
	generator.RootNode.Accept(mainVisitor)
	bar.Add(1)

	// Write Source Code to Files/Folders
	writerVisitor := generators.NewSourceCodeWriterVisitor(logger, config.OutDir, config.AppName, config.SrcDir, specParser.RemoteTypes, specParser.Services, specParser.PathPkgs)
	generator.RootNode.Accept(writerVisitor)
	bar.Add(1)

	fmt.Println("SUCCESS: Generated System!")
}
