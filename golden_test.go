package main

import (
	"fmt"
	"go/format"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testcase struct {
	name              string
	typeName          string
	configType        string
	configItemType    string
	configBuilderType string
	configOptionType  string
	needOption        bool
	want              string
}

func (tc testcase) test(t *testing.T) {
	g := newGenerator(
		tc.typeName,
		tc.configType,
		tc.configItemType,
		tc.configBuilderType,
		tc.configOptionType,
		tc.needOption,
	)
	g.generate()
	got, err := format.Source(g.bytes())
	assert.Nil(t, err)
	w, err := format.Source([]byte(tc.want))
	assert.Nil(t, err)
	assert.Equal(t, string(w), string(got))
}

// generate a test case when only one field (V typeName) is given.
func generateSimpleTestcase(typeName string) testcase {
	return testcase{
		name:              typeName,
		typeName:          fmt.Sprintf("V %s", typeName),
		configType:        "Config",
		configItemType:    "Item",
		configBuilderType: "Builder",
		configOptionType:  "Option",
		want:              fmt.Sprintf(simpleTestWantTemplate, typeName),
	}
}

const simpleTestWantTemplate = `type Item[T any] struct {
       modified     bool
       value        T
       defaultValue T
}

func (s *Item[T]) Set(value T) {
       s.modified = true
       s.value = value
}
func (s *Item[T]) Get() T {
       if s.modified {
               return s.value
       }
       return s.defaultValue
}
func (s *Item[T]) Default() T {
       return s.defaultValue
}
func (s *Item[T]) IsModified() bool {
       return s.modified
}
func NewItem[T any](defaultValue T) *Item[T] {
       return &Item[T]{
               defaultValue: defaultValue,
       }
}

type Config struct {
       V *Item[%[1]s]
}
type Builder struct {
       v %[1]s
}

func (s *Builder) V(v %[1]s) *Builder {
       s.v = v
       return s
}
func (s *Builder) Build() *Config {
       return &Config{
               V: NewItem(s.v),
       }
}

func NewBuilder() *Builder { return &Builder{} }
`

func TestGolden(t *testing.T) {
	dir, err := os.MkdirTemp("", "goconfig")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	simpleTestcaseTypenames := []string{
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
		"flag.ErrorHandler",
		"chan chan map[string]int",
	}

	simpleTestcases := make([]testcase, len(simpleTestcaseTypenames))
	for i, typeName := range simpleTestcaseTypenames {
		simpleTestcases[i] = generateSimpleTestcase(typeName)
	}

	compositeTestcases := []testcase{
		{
			name:              "minimum",
			typeName:          "I int",
			configType:        "Config",
			configItemType:    "Item",
			configBuilderType: "Builder",
			configOptionType:  "Option",
			want: `type Item[T any] struct {
       modified     bool
       value        T
       defaultValue T
}

func (s *Item[T]) Set(value T) {
       s.modified = true
       s.value = value
}
func (s *Item[T]) Get() T {
       if s.modified {
               return s.value
       }
       return s.defaultValue
}
func (s *Item[T]) Default() T {
       return s.defaultValue
}
func (s *Item[T]) IsModified() bool {
       return s.modified
}
func NewItem[T any](defaultValue T) *Item[T] {
       return &Item[T]{
               defaultValue: defaultValue,
       }
}

type Config struct {
       I *Item[int]
}
type Builder struct {
       i int
}

func (s *Builder) I(v int) *Builder {
       s.i = v
       return s
}
func (s *Builder) Build() *Config {
       return &Config{
               I: NewItem(s.i),
       }
}

func NewBuilder() *Builder { return &Builder{} }
`,
		},
		{
			name:              "pkg",
			typeName:          "ErrorHandling flag.ErrorHandling",
			configType:        "Config",
			configItemType:    "Item",
			configBuilderType: "Builder",
			configOptionType:  "Option",
			want: `type Item[T any] struct {
       modified     bool
       value        T
       defaultValue T
}

func (s *Item[T]) Set(value T) {
       s.modified = true
       s.value = value
}
func (s *Item[T]) Get() T {
       if s.modified {
               return s.value
       }
       return s.defaultValue
}
func (s *Item[T]) Default() T {
       return s.defaultValue
}
func (s *Item[T]) IsModified() bool {
       return s.modified
}
func NewItem[T any](defaultValue T) *Item[T] {
       return &Item[T]{
               defaultValue: defaultValue,
       }
}

type Config struct {
       ErrorHandling *Item[flag.ErrorHandling]
}
type Builder struct {
       errorHandling flag.ErrorHandling
}

func (s *Builder) ErrorHandling(v flag.ErrorHandling) *Builder {
       s.errorHandling = v
       return s
}
func (s *Builder) Build() *Config {
       return &Config{
               ErrorHandling: NewItem(s.errorHandling),
       }
}

func NewBuilder() *Builder { return &Builder{} }
`,
		},
		{
			name:              "alias",
			typeName:          "Handler flag.ErrorHandling",
			configType:        "Config",
			configItemType:    "Item",
			configBuilderType: "Builder",
			configOptionType:  "Option",
			want: `type Item[T any] struct {
       modified     bool
       value        T
       defaultValue T
}

func (s *Item[T]) Set(value T) {
       s.modified = true
       s.value = value
}
func (s *Item[T]) Get() T {
       if s.modified {
               return s.value
       }
       return s.defaultValue
}
func (s *Item[T]) Default() T {
       return s.defaultValue
}
func (s *Item[T]) IsModified() bool {
       return s.modified
}
func NewItem[T any](defaultValue T) *Item[T] {
       return &Item[T]{
               defaultValue: defaultValue,
       }
}

type Config struct {
       Handler *Item[flag.ErrorHandling]
}
type Builder struct {
       handler flag.ErrorHandling
}

func (s *Builder) Handler(v flag.ErrorHandling) *Builder {
       s.handler = v
       return s
}
func (s *Builder) Build() *Config {
       return &Config{
               Handler: NewItem(s.handler),
       }
}

func NewBuilder() *Builder { return &Builder{} }
`,
		},
		{
			name:              "types",
			typeName:          "B bool|Handler flag.ErrorHandling|ErrorHandling flag.ErrorHandling",
			configType:        "Config",
			configItemType:    "Item",
			configBuilderType: "Builder",
			configOptionType:  "Option",
			want: `type Item[T any] struct {
       modified     bool
       value        T
       defaultValue T
}

func (s *Item[T]) Set(value T) {
       s.modified = true
       s.value = value
}
func (s *Item[T]) Get() T {
       if s.modified {
               return s.value
       }
       return s.defaultValue
}
func (s *Item[T]) Default() T {
       return s.defaultValue
}
func (s *Item[T]) IsModified() bool {
       return s.modified
}
func NewItem[T any](defaultValue T) *Item[T] {
       return &Item[T]{
               defaultValue: defaultValue,
       }
}

type Config struct {
       B             *Item[bool]
       Handler       *Item[flag.ErrorHandling]
       ErrorHandling *Item[flag.ErrorHandling]
}
type Builder struct {
       b             bool
       handler       flag.ErrorHandling
       errorHandling flag.ErrorHandling
}

func (s *Builder) B(v bool) *Builder {
       s.b = v
       return s
}
func (s *Builder) Handler(v flag.ErrorHandling) *Builder {
       s.handler = v
       return s
}
func (s *Builder) ErrorHandling(v flag.ErrorHandling) *Builder {
       s.errorHandling = v
       return s
}
func (s *Builder) Build() *Config {
       return &Config{
               B:             NewItem(s.b),
               Handler:       NewItem(s.handler),
               ErrorHandling: NewItem(s.errorHandling),
       }
}

func NewBuilder() *Builder { return &Builder{} }
`,
		},
	}

	testcases := append(simpleTestcases, compositeTestcases...)

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, tc.test)
	}
}
