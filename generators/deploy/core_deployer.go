package deploy

import (
	"errors"
)

type DeployInfo struct {
	Address     string
	Hostname    string
	Port        int
	DockerPath  string
	ImageName   string
	EnvVars     map[string]string
	PublicPorts map[int]int
	NumReplicas int
	Command     []string
	Entrypoint  []string
	Volumes     []string
}

func NewDeployInfo() *DeployInfo {
	return &DeployInfo{EnvVars: make(map[string]string), PublicPorts: make(map[int]int)}
}

type NoOpDeployerGenerator struct{}

func (n *NoOpDeployerGenerator) AddService(name string, depInfo *DeployInfo) {}
func (n *NoOpDeployerGenerator) AddChoice(name string, depInfo *DeployInfo)  {}
func (n *NoOpDeployerGenerator) GenerateConfigFiles(out_dir string) error    { return nil }

func NewNoOpDeployerGenerator() DeployerGenerator {
	return &NoOpDeployerGenerator{}
}

type DeployerGeneratorFactory struct {
	Generators     map[string]DeployerGenerator
	GeneratorFuncs map[string]func() DeployerGenerator
}

var factory *DeployerGeneratorFactory

func GetDepGenFactory() *DeployerGeneratorFactory {
	if factory == nil {
		gen_funcs := make(map[string]func() DeployerGenerator)
		gen_funcs["docker"] = NewDockerComposeDeployerGenerator
		gen_funcs["kubernetes"] = NewKubernetesDeployerGenerator
		gen_funcs["ansible"] = NewAnsibleDeployerGenerator
		gen_funcs["noop"] = NewNoOpDeployerGenerator
		factory = &DeployerGeneratorFactory{Generators: make(map[string]DeployerGenerator), GeneratorFuncs: gen_funcs}
	}
	return factory
}

func (df *DeployerGeneratorFactory) GetGenerator(framework string) (DeployerGenerator, error) {
	if gen, ok := df.Generators[framework]; !ok {
		if gen_func, ok2 := df.GeneratorFuncs[framework]; ok2 {
			generator := gen_func()
			df.Generators[framework] = generator
			return generator, nil
		} else {
			return nil, errors.New("Framework " + framework + " was not registered as a valid DeployerGenerator")
		}
	} else {
		return gen, nil
	}
	return nil, nil
}
