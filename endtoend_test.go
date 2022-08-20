package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	g := newGoConfig(t)
	defer g.close()

	for i, tc := range []struct {
		name              string
		fileName          string
		typeName          string
		configType        string
		configItemType    string
		configBuilderType string
		configOptionType  string
		needOption        bool
	}{
		{
			name:              "types",
			fileName:          "types.go",
			typeName:          "Size int,Rule,Reverse Rule,io.Reader",
			configType:        "Config",
			configItemType:    "Item",
			configBuilderType: "Builder",
			configOptionType:  "Option",
			needOption:        true,
		},
	} {
		i := i
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			g.compileAndRun(
				t,
				i,
				tc.fileName,
				tc.typeName,
				tc.configType,
				tc.configItemType,
				tc.configBuilderType,
				tc.configOptionType,
				tc.needOption,
			)
		})
	}
}

type goConfig struct {
	dir      string
	goConfig string
}

func newGoConfig(t *testing.T) *goConfig {
	t.Helper()
	s := &goConfig{}
	s.init(t)
	return s
}

func (s *goConfig) init(t *testing.T) {
	t.Helper()
	dir, err := os.MkdirTemp("", "goconfig")
	if err != nil {
		t.Fatal(err)
	}
	goConfig := filepath.Join(dir, "goconfig")
	// build goConfig
	if err := run("go", "build", "-o", goConfig); err != nil {
		t.Fatal(err)
	}
	s.dir = dir
	s.goConfig = goConfig
}

func (s *goConfig) close() {
	os.RemoveAll(s.dir)
}

type arrayArgs struct {
	args []string
}

func (s *arrayArgs) add(key, value string) {
	if value != "" {
		s.args = append(s.args, key, value)
	}
}

func (s *goConfig) compileAndRun(
	t *testing.T,
	caseNumber int,
	fileName string,
	typeName,
	configType,
	configItemType,
	configBuilderType,
	configOptionType string,
	needOption bool,
) {
	t.Helper()
	src := filepath.Join(s.dir, fileName)
	if err := copyFile(src, filepath.Join("testdata", fileName)); err != nil {
		t.Fatal(err)
	}
	goConfigSrc := filepath.Join(s.dir, fmt.Sprintf("config%d.go", caseNumber))
	// run goConfig
	var args arrayArgs
	args.add("-type", typeName)
	args.add("-output", goConfigSrc)
	args.add("-config", configType)
	args.add("-configItem", configItemType)
	args.add("-configBuilder", configBuilderType)
	args.add("-configOption", configOptionType)
	if needOption {
		args.args = append(args.args, "-option")
	}
	t.Logf("run: goconfig %s", strings.Join(args.args, " "))
	if err := run(s.goConfig, args.args...); err != nil {
		t.Fatal(err)
	}
	// run testfile with generated file
	if err := run("go", "run", goConfigSrc, src); err != nil {
		t.Fatal(err)
	}
}

func run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(to, from string) error {
	toFile, err := os.Create(to)
	if err != nil {
		return err
	}
	defer toFile.Close()
	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFile.Close()
	_, err = io.Copy(toFile, fromFile)
	return err
}
