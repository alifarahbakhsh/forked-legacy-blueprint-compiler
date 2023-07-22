package deploy

import (
	"os"
	"os/exec"
	"log"
)

type KubernetesDeployerGenerator struct {
	*DockerComposeDeployerGenerator
}

func NewKubernetesDeployerGenerator() DeployerGenerator {
	return &KubernetesDeployerGenerator{&DockerComposeDeployerGenerator{}}
}

func (k *KubernetesDeployerGenerator) GenerateConfigFiles(out_dir string) error {
	err := k.DockerComposeDeployerGenerator.GenerateConfigFiles(out_dir)
	if err != nil {
		return err
	}

	cur_dir, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.Chdir(out_dir)
	if err != nil {
		return err
	}

	cmd := exec.Command("kompose", "convert")
	stdout, err := cmd.CombinedOutput()
	log.Println(string(stdout))

	if err != nil {
		return err
	}
	err = os.Chdir(cur_dir)
	if err != nil {
		return err
	}
	return nil
}