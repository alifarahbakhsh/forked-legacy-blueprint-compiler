package deploy

type DeployerGenerator interface {
	AddService(name string, depInfo *DeployInfo)
	AddChoice(name string, depInfo *DeployInfo)
	GenerateConfigFiles(out_dir string) error
}
