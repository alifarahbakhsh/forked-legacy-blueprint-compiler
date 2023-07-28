package deploy

import (
	"os"
	"path"
	"strconv"
	"strings"
)

type DockerComposeDeployerGenerator struct {
	composeString string
}

func NewDockerComposeDeployerGenerator() DeployerGenerator {
	return &DockerComposeDeployerGenerator{}
}

func (d *DockerComposeDeployerGenerator) AddService(name string, depInfo *DeployInfo) {
	prefix := "  "
	pub_address := depInfo.Address
	d.composeString += prefix + strings.ToLower(name) + ":\n"
	d.composeString += prefix + prefix + "build:\n"
	d.composeString += prefix + prefix + prefix + "context: .\n"
	d.composeString += prefix + prefix + prefix + "dockerfile: ./" + depInfo.DockerPath + "/Dockerfile\n"
	d.composeString += prefix + prefix + "hostname: " + pub_address + "\n"
	if depInfo.NumReplicas > 0 {
		d.composeString += prefix + prefix + "deploy:\n"
		d.composeString += prefix + prefix + prefix + "replicas: " + strconv.Itoa(depInfo.NumReplicas) + "\n"
	}
	d.composeString += prefix + prefix + "ports:\n"
	d.composeString += prefix + prefix + prefix + "- \"" + strconv.Itoa(depInfo.Port) + ":" + strconv.Itoa(depInfo.Port) + "\"\n"
	for port1, port2 := range depInfo.PublicPorts {
		d.composeString += prefix + prefix + prefix + "- \"" + strconv.Itoa(port1) + ":" + strconv.Itoa(port2) + "\"\n"
	}
	if len(depInfo.EnvVars) != 0 {
		d.composeString += prefix + prefix + "environment:\n"
	}
	for key, val := range depInfo.EnvVars {
		d.composeString += prefix + prefix + prefix + "- " + key + "=" + val + "\n"
	}
	d.composeString += prefix + prefix + "restart: always\n\n"
}

func (d *DockerComposeDeployerGenerator) AddChoice(name string, depInfo *DeployInfo) {
	prefix := "  "
	pub_address := depInfo.Address
	d.composeString += prefix + strings.ToLower(name) + ":\n"
	d.composeString += prefix + prefix + "image: " + depInfo.ImageName + "\n"
	d.composeString += prefix + prefix + "hostname: " + pub_address + "\n"
	if len(depInfo.Volumes) != 0 {
		d.composeString += prefix + prefix + "volumes:\n"
		for _, vol := range depInfo.Volumes {
			d.composeString += prefix + prefix + prefix + "- " + vol + "\n"
		}
	}
	if len(depInfo.PublicPorts) != 0 {
		d.composeString += prefix + prefix + "ports:\n"
		for port1, port2 := range depInfo.PublicPorts {
			d.composeString += prefix + prefix + prefix + "- \"" + strconv.Itoa(port1) + ":" + strconv.Itoa(port2) + "\"\n"
		}
	}
	if len(depInfo.EnvVars) != 0 {
		d.composeString += prefix + prefix + "environment:\n"
		for key, val := range depInfo.EnvVars {
			d.composeString += prefix + prefix + prefix + "- " + key + "=" + val + "\n"
		}
	}
	if len(depInfo.Command) != 0 {
		d.composeString += prefix + prefix + "command:\n"
		d.composeString += prefix + prefix + prefix + "[" + strings.Join(depInfo.Command, ", ") + "]\n"
	}
	if len(depInfo.Entrypoint) != 0 {
		d.composeString += prefix + prefix + "entrypoint: [" + strings.Join(depInfo.Entrypoint, ", ") + "]\n"
	}
	d.composeString += prefix + prefix + "restart: always\n\n"
}

func (d *DockerComposeDeployerGenerator) GenerateConfigFiles(out_dir string) error {
	outfile := path.Join(out_dir, "docker-compose.yml")
	outf, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	d.composeString = "version: '3'\nservices:\n" + d.composeString
	_, err = outf.WriteString(d.composeString)
	return err
}
