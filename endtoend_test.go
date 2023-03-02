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

type endToEndTestcase struct {
	name              string
	fileName          string
	field             string
	configType        string
	configItemType    string
	configBuilderType string
	configOptionType  string
	needOption        bool
	typePrefix        string
}

func (tc *endToEndTestcase) test(t *testing.T, caseNumber int, g *goConfig) {
	g.compileAndRun(
		t,
		caseNumber,
		tc.fileName,
		tc.field,
		tc.configType,
		tc.configItemType,
		tc.configBuilderType,
		tc.configOptionType,
		tc.needOption,
		tc.typePrefix,
	)
}

func TestEndToEnd(t *testing.T) {
	const testdataDir = "testdata"
	g := newGoConfig(t, testdataDir)
	defer g.close()

	simpleTestcaseTypeNames := []string{
		"int",
		"string",
		"[]int",
		"[1]int",
		"map[string]int",
		"chan string",
		"chan<- string",
		"<-chan string",
		"func()",
		"[][]int",
		"[]map[string]int",
		"map[string][]int",
		"chan []int",
		"func() error",
		"func(int)",
		"func(int) error",
		"func(int) (string, error)",
		"func(int, string) (map[string]int, error)",
		"*int",
		"*[]int",
		"chan chan map[string]int",
	}

	const simpleEndToEndTestcasePrefix = "simple_"
	removeSimpleEndToEndTestcaseSources := func() {
		files, err := filepath.Glob(filepath.Join(testdataDir, fmt.Sprintf("%s*", simpleEndToEndTestcasePrefix)))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			if err := os.Remove(file); err != nil {
				t.Logf("Failed to remove %s", file)
			}
		}
	}
	removeSimpleEndToEndTestcaseSources()
	simpleTestcases := make([]*endToEndTestcase, len(simpleTestcaseTypeNames))
	for i, typeName := range simpleTestcaseTypeNames {
		tc, err := generateSimpleEndToEndTestcase(
			testdataDir,
			fmt.Sprintf("%s_%d", simpleEndToEndTestcasePrefix, i),
			typeName,
		)
		if err != nil {
			t.Fatal(err)
		}
		simpleTestcases[i] = tc
	}
	defer removeSimpleEndToEndTestcaseSources()

	compositeTestcases := []*endToEndTestcase{
		{
			name:              "types",
			fileName:          "types.go",
			field:             "Size int|Rule Rule|Reverse Rule|Reader io.Reader",
			configType:        "Config",
			configItemType:    "Item",
			configBuilderType: "Builder",
			configOptionType:  "Option",
			needOption:        true,
			typePrefix:        "",
		},
		{
			name:              "types-prefix",
			fileName:          "types_prefix.go",
			field:             "Size int|Rule Rule|Reverse Rule|Reader io.Reader",
			configType:        "Config",
			configItemType:    "Item",
			configBuilderType: "Builder",
			configOptionType:  "Option",
			needOption:        true,
			typePrefix:        "Prefix",
		},
	}

	testcases := append(simpleTestcases, compositeTestcases...)

	for i, tc := range testcases {
		i := i
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, i, g)
		})
	}
}

type goConfig struct {
	testdataDir string
	dir         string
	goConfig    string
}

func newGoConfig(t *testing.T, testdataDir string) *goConfig {
	t.Helper()
	s := &goConfig{
		testdataDir: testdataDir,
	}
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
	field,
	configType,
	configItemType,
	configBuilderType,
	configOptionType string,
	needOption bool,
	typePrefix string,
) {
	t.Helper()

	src := filepath.Join(s.dir, fileName)
	if err := copyFile(src, filepath.Join(s.testdataDir, fileName)); err != nil {
		t.Fatal(err)
	}
	goConfigSrc := filepath.Join(s.dir, fmt.Sprintf("config%d.go", caseNumber))
	// run goConfig
	var args arrayArgs
	args.add("-field", field)
	args.add("-output", goConfigSrc)
	args.add("-config", configType)
	args.add("-configItem", configItemType)
	args.add("-configBuilder", configBuilderType)
	args.add("-configOption", configOptionType)
	args.add("-prefix", typePrefix)
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

// generate a test case when only one field (V typeName) is given.
func generateSimpleEndToEndTestcase(dir, name, typeName string) (*endToEndTestcase, error) {
	fileName := fmt.Sprintf("%s.go", name)
	filePath := filepath.Join(dir, fileName)
	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, simpleEndToEndTestcaseSourceTemplate, typeName, "%#v"); err != nil {
		return nil, err
	}
	return &endToEndTestcase{
		name:              name,
		fileName:          fileName,
		field:             fmt.Sprintf("V %s", typeName),
		configType:        "Config",
		configItemType:    "Item",
		configBuilderType: "Builder",
		configOptionType:  "Option",
		needOption:        true,
		typePrefix:        "",
	}, nil
}

const simpleEndToEndTestcaseSourceTemplate = `package main
import "fmt"
func check(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}
func main() {
  var d %[1]s
  c := NewBuilder().V(d).Build()
  check(!c.V.IsModified(), "not modified")
  c.Apply(WithV(d))
  check(c.V.IsModified(), "modified")
  fmt.Printf("%[2]s\n", c.V)
}
`
