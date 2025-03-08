package deploy

import (
	"fmt"
	"os"
	"path"
	"strconv"

	cp "github.com/otiai10/copy"
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

//? This part will need refactoring
type InstanceDepInfo struct {
	*DeployInfo
	ThirdParty bool
	// Name string
}

//-----------------------------------------------------------------

type AnsibleDeployerGenerator struct {
	mainPlaybook    string
	buildPlaybook   string
	imagePlaybook   string
	runnerPlaybooks []string
	inventory       string
	srcDir          string
	outDir          string
	instanceData    []InstanceDepInfo
	inventoryData   []parser.Node
}

//-----------------------------------------------------------------

type Env struct {
	Name  string
	Value string
}

type Port struct {
	Host      int
	Container int
}

type BuildImage struct {
	DockerPath string
	TagName    string
}

type Instance struct {
	Name       string
	ImgName    string
	ThirdParty bool
	Ports      []Port
	EnvVars    []Env
}

type Host struct {
	// Address string
	Idx      uint16
	Services []Instance
}
type HostMap map[string]*Host

func NewAnsibleDeployerGenerator() DeployerGenerator {
	return &AnsibleDeployerGenerator{}
}

func (adg *AnsibleDeployerGenerator) SetInventory(inventory []parser.Node) {
	adg.inventoryData = inventory
}

func (adg *AnsibleDeployerGenerator) AddService(name string, depInfo *DeployInfo) {
	depInfo.ImageName = name
	adg.instanceData = append(adg.instanceData, InstanceDepInfo{
		DeployInfo: depInfo,
		ThirdParty: false,
		// Name: name,
	})
}

func (adg *AnsibleDeployerGenerator) AddChoice(name string, depInfo *DeployInfo) {
	adg.instanceData = append(adg.instanceData, InstanceDepInfo{
		DeployInfo: depInfo,
		ThirdParty: true,
		// Name: name,
	})
}

func (adg *AnsibleDeployerGenerator) WriteFile(name, data string) error {

	err := os.MkdirAll(adg.outDir, 0755)
	if err != nil {
		return err
	}

	outfile := path.Join(adg.outDir, name)
	outf, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	_, err = outf.WriteString(data)

	return err
}

func (adg *AnsibleDeployerGenerator) CreateInventory(hosts HostMap) error {
	fmt.Println("Generating Inventory..")
	adg.inventory += "[all]\n"
	nameList := []string{}

	for k, _ := range hosts {
		nameList = append(nameList, k)
	}

	var registryNode string
	for _, v := range nameList {
		adg.inventory += v
		adg.inventory += " ansible_user=root\n\n"

		for _, node := range adg.inventoryData {
			if node.Hostname == v {
				if node.IsBuildNode {
					registryNode = v
				}
			}
		}
	}

	//! The last host is selected as a registry host unless we specify sth else in the config file
	adg.inventory += "[registry]\n"
	if registryNode != "" {
		adg.inventory += registryNode
	} else {
		adg.inventory += nameList[len(nameList)-1]
	}

	adg.inventory += " ansible_user=root\n\n"

	//TODO decide on how to handle these corner cases
	// adg += "[all:vars]\nansible_python_interpreter=/user/bin/python3"
	return adg.WriteFile("inventory", adg.inventory)
}

func (adg *AnsibleDeployerGenerator) CreateImagePlaybook(hosts HostMap) error {
	prefix := "  "
	fmt.Println("Generating Images Playbook..")
	adg.imagePlaybook += "---\n"
	adg.imagePlaybook += "- name: Building..\n"
	adg.imagePlaybook += prefix + "debug:\n" + prefix + prefix + "msg: \"Now building: {{image.name}} at {{image.rel_path}}\"\n\n"
	adg.imagePlaybook += "- name: Building image and pushing to private repo\n"
	adg.imagePlaybook += prefix + "community.docker.docker_image:\n"
	adg.imagePlaybook += prefix + prefix + "build:\n"
	adg.imagePlaybook += prefix + prefix + prefix + "path: \"{{dst.path}}\"\n"
	adg.imagePlaybook += prefix + prefix + prefix + "dockerfile: \"{{image.rel_path}}\"\n"
	adg.imagePlaybook += prefix + prefix + "name: \"{{image.name}}\"\n"
	adg.imagePlaybook += prefix + prefix + "repository: \"localhost:5000/{{image.name}}\"\n"
	adg.imagePlaybook += prefix + prefix + "push: yes\n"
	adg.imagePlaybook += prefix + prefix + "source: build\n\n"
	adg.imagePlaybook += "- name: Saving image name..\n"
	adg.imagePlaybook += prefix + "set_fact:\n"
	adg.imagePlaybook += prefix + prefix + "image_list: \"{{image_list}} + [ '{{inventory_hostname}}:5000/{{image.name}}' ]\""

	return adg.WriteFile("build_image.yaml", adg.imagePlaybook)
}

func (adg *AnsibleDeployerGenerator) CreateBuildPlaybook(images []BuildImage) error {

	// ! Prepare build playbook

	prefix := "  "
	fmt.Println("Generating Builds Playbook..")
	adg.buildPlaybook += "---\n"
	adg.buildPlaybook += "- hosts: registry\n"
	adg.buildPlaybook += prefix + "vars:\n"
	adg.buildPlaybook += prefix + prefix + "src_path: " + adg.srcDir + "\n" //TODO Clarify where the source path is provided
	adg.buildPlaybook += prefix + prefix + "dst_path: /tmp/build\n"         //TODO Clarify where the destination path is provided
	adg.buildPlaybook += prefix + prefix + "images:\n"
	for _, v := range images {
		adg.buildPlaybook += prefix + prefix + prefix + "- rel_path: " + v.DockerPath + "\n"
		adg.buildPlaybook += prefix + prefix + prefix + prefix + "name: " + v.TagName + "\n"
	}

	adg.buildPlaybook += "\n"
	adg.buildPlaybook += prefix + "tasks:\n"
	adg.buildPlaybook += prefix + prefix + "- name: Check if target has source files\n"
	adg.buildPlaybook += prefix + prefix + prefix + "stat:\n"
	adg.buildPlaybook += prefix + prefix + prefix + prefix + "path: /tmp/build\n"
	adg.buildPlaybook += prefix + prefix + prefix + "register: st\n"
	adg.buildPlaybook += prefix + prefix + "- name: Copy src for all builds\n"
	adg.buildPlaybook += prefix + prefix + prefix + "copy:\n"
	adg.buildPlaybook += prefix + prefix + prefix + prefix + "src: \"{{src_path}}\"\n"
	adg.buildPlaybook += prefix + prefix + prefix + prefix + "src: \"{{dst_path}}\"\n"
	adg.buildPlaybook += prefix + prefix + prefix + prefix + "mode: u+x, g+x\n"
	adg.buildPlaybook += prefix + prefix + prefix + prefix + "force: no\n"
	adg.buildPlaybook += prefix + prefix + prefix + "when: not st.stat.exists\n"
	adg.buildPlaybook += prefix + prefix + "- name: Building images\n"
	adg.buildPlaybook += prefix + prefix + prefix + "include_tasks: build_image.yaml\n"
	adg.buildPlaybook += prefix + prefix + prefix + "loop: \"{{images}}\"\n"
	adg.buildPlaybook += prefix + prefix + prefix + "loop_control:\n"
	adg.buildPlaybook += prefix + prefix + prefix + prefix + "loop_var: image\n"

	adg.buildPlaybook += "\n"
	adg.buildPlaybook += "- hosts: all\n"
	adg.buildPlaybook += prefix + "run_once: yes\n"
	adg.buildPlaybook += prefix + "gather_facts: no\n"
	adg.buildPlaybook += prefix + "tasks:\n"
	adg.buildPlaybook += prefix + prefix + "- name: List pullable images\n"
	adg.buildPlaybook += prefix + prefix + prefix + "debug:\n"
	adg.buildPlaybook += prefix + prefix + prefix + prefix + "msg: \"{{ hostvars[groups['registry'][0]].image_list }}\""

	return adg.WriteFile("build.yaml", adg.buildPlaybook)
}

func (adg *AnsibleDeployerGenerator) CreateMainPlaybook(hosts HostMap, noBuild bool) error {

	prefix := "  "
	fmt.Println("Generating Main Playbook..")
	adg.mainPlaybook += "---\n"

	if !noBuild {
		adg.mainPlaybook += "- name: Build Images\n"
		adg.mainPlaybook += prefix + "import_playbook: build.yaml\n\n"
	}

	for addr, _ := range hosts {
		//create entry for playbook with name 'addr.yaml'
		adg.mainPlaybook += "- name: Running " + addr + "\n"
		adg.mainPlaybook += prefix + "import_playbook: " + addr + ".yaml\n"
	}

	return adg.WriteFile("main.yaml", adg.mainPlaybook)
}

func (adg *AnsibleDeployerGenerator) CreateRunnerPlaybooks(hosts HostMap) error {

	fmt.Println("Generating runners..")
	prefix := "  "
	var err error
	// //! Iteration order in map is not defined; that's why we need to maintain host Indexes explicitly
	for addr, h := range hosts {
		//create entry for playbook with name 'addr.yaml'

		playbook := "---\n"

		playbook += "- hosts: all[" + strconv.FormatUint(uint64(h.Idx), 10) + "]\n"
		playbook += prefix + "vars:\n"
		playbook += prefix + prefix + "services:\n"

		for _, instance := range h.Services {

			playbook += prefix + prefix + prefix + "- name: " + instance.Name + "\n"
			playbook += prefix + prefix + prefix + prefix + "img_name: " + instance.ImgName + "\n"

			if instance.ThirdParty {
				playbook += prefix + prefix + prefix + prefix + "third_party_image: yes\n"
			} else {
				playbook += prefix + prefix + prefix + prefix + "third_party_image: no\n"
			}

			playbook += prefix + prefix + prefix + prefix + "ports:\n"

			for _, port := range instance.Ports {

				p1 := strconv.FormatUint(uint64(port.Host), 10)
				p2 := strconv.FormatUint(uint64(port.Container), 10)
				playbook += prefix + prefix + prefix + prefix + prefix + "- \"" + p1 + ":" + p2 + "\"\n"
			}
			// playbook += prefix + prefix + prefix + prefix + "env: " + instance.? + "\n"

			playbook += prefix + prefix + prefix + prefix + "env:\n"

			for _, env := range instance.EnvVars {
				playbook += prefix + prefix + prefix + prefix + prefix + env.Name + ": " + env.Value + "\n"
			}
			// playbook += prefix + prefix + prefix + prefix + "env: " + instance.? + "\n"

		}

		playbook += prefix + "tasks:\n"
		playbook += prefix + prefix + "- name: Pull and run\n"
		playbook += prefix + prefix + prefix + "block:\n"
		playbook += prefix + prefix + prefix + prefix + "- name: Running third-party images\n"
		playbook += prefix + prefix + prefix + prefix + prefix + "community.docker.docker_container:\n"
		playbook += prefix + prefix + prefix + prefix + prefix + prefix + "name: \"{{item.name}}\"\n"
		playbook += prefix + prefix + prefix + prefix + prefix + prefix + "image: \"{{item.img_name}}\"\n"
		playbook += prefix + prefix + prefix + prefix + prefix + prefix + "restart_policy: always\n"
		// playbook += prefix + prefix + prefix + prefix + prefix + prefix + "ports: \"{{item.ports}}\" \n"
		playbook += prefix + prefix + prefix + prefix + prefix + "loop: \"{{services}}\"\n"
		playbook += prefix + prefix + prefix + prefix + prefix + "when: item.third_party_image\n\n"

		playbook += prefix + prefix + prefix + prefix + "- name: Running own images\n"
		playbook += prefix + prefix + prefix + prefix + prefix + "community.docker.docker_container:\n"
		playbook += prefix + prefix + prefix + prefix + prefix + prefix + "name: \"{{item.name}}\"\n"
		playbook += prefix + prefix + prefix + prefix + prefix + prefix + "image: \"{{groups['registry'][0] + :5000/ + item.img_name}}\"\n"
		playbook += prefix + prefix + prefix + prefix + prefix + prefix + "restart_policy: always\n"
		// playbook += prefix + prefix + prefix + prefix + prefix + prefix + "ports: \"{{item.ports}}\" \n"
		playbook += prefix + prefix + prefix + prefix + prefix + "loop: \"{{services}}\"\n"
		playbook += prefix + prefix + prefix + prefix + prefix + "when: not item.third_party_image\n\n"

		adg.runnerPlaybooks = append(adg.runnerPlaybooks, playbook)

		err = adg.WriteFile(addr+".yaml", playbook)
		if err != nil {
			return err
		}
	}

	return nil
}

func (adg *AnsibleDeployerGenerator) CopySetupFiles() error {

	fmt.Println("Copying setup files..")
	pwd, _ := os.Getwd()
	setupFilesPath := path.Join(pwd, "generators/deploy/cluster_setup")
	outPath := path.Join(adg.outDir, "cluster_setup")
	err := cp.Copy(setupFilesPath, outPath)
	return err
}

//? When replication is supported, this function is where the changes should take place
func (adg *AnsibleDeployerGenerator) GenerateConfigFiles(out_dir string) error {

	adg.outDir = path.Join(out_dir, "ansible")
	err := os.RemoveAll(adg.outDir)
	if err != nil {
		return err
	}

	adg.srcDir = out_dir
	plac := adg.instanceData

	hosts := make(HostMap)
	images := []BuildImage{}

	//! Preprocess the input to generate ansible-friendly data structures
	hostIdx := uint16(0)
	for _, pl := range plac {

		if !pl.ThirdParty {
			images = append(images, BuildImage{
				DockerPath: pl.DockerPath,
				TagName:    pl.ImageName,
			})
		}

		var ports []Port
		var envs []Env
		for h, c := range pl.PublicPorts {

			ports = append(ports, Port{
				Host:      h,
				Container: c,
			})
		}

		for e1, e2 := range pl.EnvVars {
			envs = append(envs, Env{
				Name:  e1,
				Value: e2,
			})
		}

		//! The following lines group instances based on hosts
		if _, ok := hosts[pl.Hostname]; !ok {

			hosts[pl.Hostname] = &Host{
				Idx: hostIdx,
				Services: []Instance{
					Instance{
						Name:       pl.Address,
						ImgName:    pl.ImageName,
						Ports:      ports,
						EnvVars:    envs,
						ThirdParty: pl.ThirdParty,
					},
				},
			}
			hostIdx++

		} else {

			hosts[pl.Hostname].Services = append(hosts[pl.Hostname].Services, Instance{
				Name:       pl.Address,
				ImgName:    pl.ImageName,
				Ports:      ports,
				EnvVars:    envs,
				ThirdParty: pl.ThirdParty,
			})
		}
	}

	err = adg.CreateInventory(hosts)
	if err != nil {
		return nil
	}

	if len(images) != 0 {
		err = adg.CreateImagePlaybook(hosts)
		if err != nil {
			return err
		}

		err = adg.CreateBuildPlaybook(images)
		if err != nil {
			return err
		}
	}

	err = adg.CreateMainPlaybook(hosts, len(images) == 0)
	if err != nil {
		return err
	}

	err = adg.CreateRunnerPlaybooks(hosts)
	if err != nil {
		return err
	}

	err = adg.CopySetupFiles()
	if err != nil {
		return err
	}

	fmt.Println("DONE")
	return nil
}
