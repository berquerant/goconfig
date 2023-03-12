# goconfig

Generate config pattern.

Given

``` go
package example
// ...
```

run `goconfig -field "Size int|ErrorHandling flag.ErrorHandling" -option` then generate

``` go
package example

import "flag"

type ConfigItem[T any] struct {
	modified     bool
	value        T
	defaultValue T
}

func (s *ConfigItem[T]) Set(value T) {
	s.modified = true
	s.value = value
}
func (s *ConfigItem[T]) Get() T {
	if s.modified {
		return s.value
	}
	return s.defaultValue
}
func (s *ConfigItem[T]) Default() T {
	return s.defaultValue
}
func (s *ConfigItem[T]) IsModified() bool {
	return s.modified
}
func NewConfigItem[T any](defaultValue T) *ConfigItem[T] {
	return &ConfigItem[T]{
		defaultValue: defaultValue,
	}
}

type Config struct {
	Size          *ConfigItem[int]
	ErrorHandling *ConfigItem[flag.ErrorHandling]
}
type ConfigBuilder struct {
	size          int
	errorHandling flag.ErrorHandling
}

func (s *ConfigBuilder) Size(v int) *ConfigBuilder {
	s.size = v
	return s
}
func (s *ConfigBuilder) ErrorHandling(v flag.ErrorHandling) *ConfigBuilder {
	s.errorHandling = v
	return s
}
func (s *ConfigBuilder) Build() *Config {
	return &Config{
		Size:          NewConfigItem(s.size),
		ErrorHandling: NewConfigItem(s.errorHandling),
	}
}

func NewConfigBuilder() *ConfigBuilder { return &ConfigBuilder{} }
func (s *Config) Apply(opt ...ConfigOption) {
	for _, x := range opt {
		x(s)
	}
}

type ConfigOption func(*Config)

func WithSize(v int) ConfigOption {
	return func(c *Config) {
		c.Size.Set(v)
	}
}
func WithErrorHandling(v flag.ErrorHandling) ConfigOption {
	return func(c *Config) {
		c.ErrorHandling.Set(v)
	}
}
```

in config.go in the same directory.

# Requirements

- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)
