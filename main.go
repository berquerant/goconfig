package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	fields            = flag.String("field", "", "list of fields by '|'; must be set")
	configType        = flag.String("config", "Config", "type name of config")
	configItemType    = flag.String("configItem", "ConfigItem", "type name of config item")
	configBuilderType = flag.String("configBuilder", "ConfigBuilder", "type name of config builder")
	configOptionType  = flag.String("configOption", "ConfigOption", "type name of config option")
	needOption        = flag.Bool("option", false, "generate option functions as WithXXX style")
	goImports         = flag.String("goimports", "goimports", "goimports executable")
	output            = flag.String("output", "", "output file name; default srcdir/config.go")
)

const usage = `Usage of goconfig:
  goconfig [flags] -field F [directory]

F is list of "fieldName typeName" separated by "|".

Environment variables:
  GOCONFIG_DEBUG
    If set, enable debug logs.

  GOCONFIG_STDOUT
    If set, write result to stdout.

Flags:`

func Usage() {
	fmt.Fprintln(os.Stderr, usage)
	flag.PrintDefaults()
}

var (
	debug            = false
	redirectToStdout = false
)

func parseEnv() {
	debug = os.Getenv("GOCONFIG_DEBUG") != ""
	if debug {
		log.Println("Debug enabled")
	}
	redirectToStdout = os.Getenv("GOCONFIG_STDOUT") != ""
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("goconfig: ")
	flag.Usage = Usage
	flag.Parse()
	parseEnv()

	if len(*fields) == 0 {
		log.Fatal("field option must be set")
	}

	g := newGenerator(
		*fields,
		*configType,
		*configItemType,
		*configBuilderType,
		*configOptionType,
		*needOption,
	)
	g.parsePackage(flag.Args())

	g.Printf("// Code generated by \"goconfig %s\"; DO NOT EDIT.\n", strings.Join(os.Args[1:], " "))
	g.Println()
	g.Printf("package %s\n", g.pkgName)
	g.Println()

	g.generate()

	if err := writeResult(g.bytes()); err != nil {
		log.Panic(err)
	}
}

func writeResult(src []byte) error {
	if redirectToStdout {
		return writeResultToStdout(src)
	}
	return writeResultToDestfile(src)
}

func writeResultToStdout(src []byte) error {
	f, err := os.CreateTemp("", "goconfig")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(f.Name())

	if err := writeResultAndFormat(src, f.Name()); err != nil {
		return err
	}
	if _, err := f.Seek(0, os.SEEK_SET); err != nil {
		return err
	}
	if _, err := io.Copy(os.Stdout, f); err != nil {
		return err
	}
	return nil
}

func writeResultToDestfile(src []byte) error {
	return writeResultAndFormat(src, destFilename())
}

func writeResultAndFormat(src []byte, fileName string) error {
	if err := os.WriteFile(fileName, src, 0600); err != nil {
		return fmt.Errorf("failed to write to %s: %w", fileName, err)
	}
	gi := &goImporter{
		goImports:  *goImports,
		targetFile: fileName,
	}
	if err := gi.doImport(); err != nil {
		return fmt.Errorf("failed to goimport: %w", err)
	}
	return nil
}

func destFilename() string {
	if *output != "" {
		return *output
	}
	return filepath.Join(destDir(), "config.go")
}

func destDir() string {
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}
	if len(args) == 1 && isDirectory(args[0]) {
		return args[0]
	}
	return filepath.Dir(args[0])
}

func isDirectory(p string) bool {
	x, err := os.Stat(p)
	if err != nil {
		log.Fatalf("directory: %v", err)
	}
	return x.IsDir()
}

type goImporter struct {
	goImports  string
	targetFile string
}

func (s *goImporter) doImport() error {
	cmd := exec.Command(s.goImports, "-w", s.targetFile)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func newGenerator(
	fields,
	configType,
	configItemType,
	configBuilderType,
	configOptionType string,
	needOption bool,
) *generator {
	item := &configItem{
		typeName:    configItemType,
		constructor: fmt.Sprintf("New%s", configItemType),
	}
	conf := &config{
		typeName:   configType,
		configItem: item,
		fields:     parseConfigFields(fields),
	}
	builder := &configBuilder{
		typeName:    configBuilderType,
		config:      conf,
		constructor: fmt.Sprintf("New%s", configBuilderType),
	}
	option := &configOption{
		typeName: configOptionType,
		config:   conf,
	}
	var b bytes.Buffer
	return &generator{
		buf:        b,
		item:       item,
		conf:       conf,
		builder:    builder,
		option:     option,
		needOption: needOption,
	}
}

type generator struct {
	buf        bytes.Buffer
	pkgName    string
	item       *configItem
	conf       *config
	builder    *configBuilder
	option     *configOption
	needOption bool
}

func (s *generator) Printf(format string, v ...any) { fmt.Fprintf(&s.buf, format, v...) }
func (s *generator) Println(v ...any)               { fmt.Fprintln(&s.buf, v...) }

func (s *generator) parsePackage(patterns []string) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName,
	}, patterns...)
	if err != nil {
		log.Fatalf("load: %v", err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("%d packages found", len(pkgs))
	}
	s.pkgName = pkgs[0].Name
	debugf("Found package: %s", s.pkgName)
}

func (s *generator) generate() {
	s.Printf(s.item.generate())
	s.Printf(s.conf.generate())
	s.Printf(s.builder.generate())
	if s.needOption {
		s.Printf(s.option.generate())
	}
}

func (s *generator) bytes() []byte { return s.buf.Bytes() }

type stringBuilder struct {
	strings.Builder
}

func (s *stringBuilder) write(format string, v ...any) {
	s.WriteString(fmt.Sprintf("%s\n", fmt.Sprintf(format, v...)))
}

type configItem struct {
	typeName    string
	constructor string
}

func (s *configItem) generate() string {
	recv := fmt.Sprintf("(s *%s[T])", s.typeName)
	return fmt.Sprintf(`type %[1]s[T any] struct {
  modified bool
  value T
  defaultValue T
}
func %[2]s Set(value T) {
  s.modified = true
  s.value = value
}
func %[2]s Get() T {
  if s.modified {
    return s.value
  }
  return s.defaultValue
}
func %[2]s Default() T {
  return s.defaultValue
}
func %[2]s IsModified() bool {
  return s.modified
}
func %[3]s[T any](defaultValue T) *%[1]s[T] {
  return &%[1]s[T]{
    defaultValue: defaultValue,
  }
}
`, s.typeName, recv, s.constructor)
}

func capitalize(v string) string {
	return fmt.Sprintf("%s%s", strings.ToUpper(string(v[0])), v[1:])
}

func decapitalize(v string) string {
	return fmt.Sprintf("%s%s", strings.ToLower(string(v[0])), v[1:])
}

func parseConfigField(field string) (*configField, error) {
	xs := strings.SplitN(field, " ", 2)
	if len(xs) != 2 {
		return nil, fmt.Errorf("field must have fieldName and typeName: %s", field)
	}

	fieldName := xs[0]
	typeName := xs[1]

	// validate typename
	if _, err := parser.ParseExpr(typeName); err != nil {
		return nil, fmt.Errorf("failed to parse field %s: %w", field, err)
	}

	return &configField{
		fieldName: capitalize(fieldName), // as public field
		typeName:  typeName,
	}, nil
}

func parseConfigFields(fields string) []*configField {
	ss := strings.Split(fields, "|")
	fs := make([]*configField, len(ss))
	for i, s := range ss {
		debugf("Parse field[%d]: %s", i, s)
		f, err := parseConfigField(s)
		if err != nil {
			log.Fatalf("Failed to parse field[%d]: %v", i, err)
		}
		debugf("Parse field[%d]: %s -> fieldName = %s typeName = %s", i, s, f.fieldName, f.typeName)
		fs[i] = f
	}
	return fs
}

type configField struct {
	typeName  string
	fieldName string
}

type config struct {
	typeName   string
	configItem *configItem
	fields     []*configField
}

func (s *config) generate() string {
	var b stringBuilder
	b.write("type %s struct {", s.typeName)
	for _, f := range s.fields {
		t := fmt.Sprintf("*%s[%s]", s.configItem.typeName, f.typeName) // config item type is generic
		b.write("%s %s", f.fieldName, t)
	}
	b.write("}") // struct
	return b.String()
}

type configOption struct {
	typeName string
	config   *config
}

func (s *configOption) generateConfigApply() string {
	return fmt.Sprintf(`func (s *%[1]s) Apply(opt ...%[2]s) {
  for _, x := range opt {
    x(s)
  }
}`, s.config.typeName, s.typeName)
}

func (s *configOption) generate() string {
	var b stringBuilder
	b.write(s.generateConfigApply())
	b.write("type %s func(*%s)", s.typeName, s.config.typeName)
	for _, f := range s.config.fields {
		withSig := fmt.Sprintf("func With%s(v %s) %s", f.fieldName, f.typeName, s.typeName)
		b.write(`%[1]s {
  return func(c *%[2]s) {
    c.%[3]s.Set(v)
  }
}`, withSig, s.config.typeName, f.fieldName)
	}
	return b.String()
}

type configBuilder struct {
	typeName    string
	constructor string
	config      *config
}

func (s *configBuilder) fieldName(i int) string {
	return decapitalize(s.config.fields[i].fieldName)
}

func (s *configBuilder) generateConstructor() string {
	return fmt.Sprintf(`func %[1]s() *%[2]s { return &%[2]s{} }`, s.constructor, s.typeName)
}

func (s *configBuilder) generateType() string {
	var b stringBuilder
	b.write("type %s struct {", s.typeName)
	for i, f := range s.config.fields {
		b.write("%s %s", s.fieldName(i), f.typeName)
	}
	b.write("}") // struct
	return b.String()
}

func (s *configBuilder) generateMethods() string {
	var b stringBuilder
	for i, f := range s.config.fields {
		b.write(`func (s *%[1]s) %[2]s(v %[3]s) *%[1]s {
  s.%[4]s = v
  return s
}`, s.typeName, f.fieldName, f.typeName, s.fieldName(i))
	}
	// Build()
	b.write("func (s *%s) Build() *%s {", s.typeName, s.config.typeName)
	b.write("return &%s{", s.config.typeName)
	for i, f := range s.config.fields {
		b.write("%s: %s(s.%s),", f.fieldName, s.config.configItem.constructor, s.fieldName(i))
	}
	b.write("}") // return
	b.write("}")
	return b.String()
}

func (s *configBuilder) generate() string {
	var b stringBuilder
	b.write(s.generateType())
	b.write(s.generateMethods())
	b.write(s.generateConstructor())
	return b.String()
}

func debugf(format string, v ...any) {
	if debug {
		log.Printf(format, v...)
	}
}
