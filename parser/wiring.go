package parser

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Copied from: https://gist.github.com/ivanzoid/129460aa08aff72862a534ebe0a9ae30
func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

type MillenialNode struct {
	Name     string          `json:"name"`
	Children []ContainerNode `json:"children"`
}

type ContainerNode struct {
	Name     string       `json:"name"`
	Children []DetailNode `json:"children"`
}

type ModifierNode struct {
	ModifierType   string         `json:"modifier_type"`
	ModifierParams []ArgumentNode `json:"modifier_params"`
}

type ArgumentNode struct {
	Name            string         `json:"name"`
	IsService       bool           `json:"isservice"`
	ClientModifiers []ModifierNode `json:"client_modifiers"`
	KeywordName     string         `json:"keyword_name"`
	Value           string         `json:"client_node"`
}

type DetailNode struct {
	Name            string         `json:"name"`
	Type            string         `json:"actual_type"`
	AbsType         string         `json:"abstract_type"`
	Arguments       []ArgumentNode `json:"arguments"`
	ClientModifiers []ModifierNode `json:"client_modifier"`
	ServerModifiers []ModifierNode `json:"server_modifiers"`
	Children        []DetailNode   `json:"children"`
}

type WiringParser struct {
	config   *Config
	logger   *log.Logger
	RootNode *MillenialNode
}

func NewWiringParser(config *Config, logger *log.Logger) *WiringParser {
	return &WiringParser{config: config, logger: logger, RootNode: nil}
}

func (w *WiringParser) ParseWiring() {
	wiring_file_output := fileNameWithoutExtension(w.config.WiringFile) + "_compiled.json"
	cmd := exec.Command("python", "wiring_translate.py", w.config.WiringFile, wiring_file_output)

	stdout, err := cmd.CombinedOutput()

	w.logger.Print(string(stdout))

	if err != nil {
		w.logger.Fatal(err)
	}

	// TODO: Parse Output

	wiringCompiledFile, err := os.Open(wiring_file_output)
	if err != nil {
		w.logger.Fatal(err)
	}
	defer wiringCompiledFile.Close()

	bytes, err := ioutil.ReadAll(wiringCompiledFile)
	if err != nil {
		w.logger.Fatal(err)
	}
	var rootNode MillenialNode
	json.Unmarshal(bytes, &rootNode)

	w.RootNode = &rootNode
}
