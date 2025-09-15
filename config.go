package service

import (
	"context"

	"dario.cat/mergo"
	"github.com/FlowSeer/fail"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// ConfigOption is a function that modifies ConfigOptions.
// It is used to configure how configuration is loaded.
type ConfigOption = func(*ConfigOptions)

// ConfigOptions holds options for reading configuration.
type ConfigOptions struct {
	// Files is a list of config file paths to load, in order.
	// Successive files override previous ones when the same key is present.
	Files []string
	// FilesPriority determines the priority of config files.
	// Lower values take precedence over higher values and are loaded last.
	// Defaults to 100.
	FilesPriority int
	// FilesRequired determines whether to fail if no config files are found or loaded successfully.
	// When false, errors are ignored.
	// Defaults to true.
	FilesRequired bool
	// EnvVars determines whether to read from environment variables.
	EnvVars bool
	// EnvVarsPriority determines the priority of environment variables.
	// Lower values take precedence over higher values and are loaded last.
	// Defaults to 1000.
	EnvVarsPriority int
	// EnvVarsPrefix is a string that sets the prefix for environment variables.
	EnvVarsPrefix string
	// TagName is the name of the struct field that will be used to populate the config struct.
	// Defaults to "json".
	TagName string
}

// DefaultConfigOptions returns a ConfigOptions struct with default values.
// By default, it enables environment variables and sets the prefix based on the service name extracted from the context.
func DefaultConfigOptions(ctx context.Context) *ConfigOptions {
	return &ConfigOptions{
		Files:           []string{},
		FilesPriority:   100,
		FilesRequired:   true,
		EnvVars:         true,
		EnvVarsPriority: 1000,
		EnvVarsPrefix:   NormalizeEnvName(Name(ctx)),
		TagName:         "json",
	}
}

// ReadConfig reads configuration into a struct of type T using the provided options.
// Returns a pointer to the struct and an error, if any.
func ReadConfig[T any](opts ...ConfigOption) (*T, error) {
	o := DefaultConfigOptions(context.Background())
	for _, opt := range opts {
		opt(o)
	}

	return ReadConfigWithOptions[T](o)
}

// ReadConfigFile reads configuration from the specified file path and applies any additional options.
// Returns a pointer to the struct and an error, if any.
func ReadConfigFile[T any](path string, opts ...ConfigOption) (*T, error) {
	return ReadConfig[T](append(opts, WithConfigFilePath(path))...)
}

// ReadConfigWithOptions reads configuration using the provided ConfigOptions struct.
// Returns a pointer to the struct and an error, if any.
func ReadConfigWithOptions[T any](opts *ConfigOptions) (*T, error) {
	return readConfig[T](opts)
}

// WithConfigFilePath returns a ConfigOption that appends the given file path to the list of config files.
func WithConfigFilePath(path string) ConfigOption {
	return func(o *ConfigOptions) {
		o.Files = append(o.Files, path)
	}
}

// WithConfigFilesPriority returns a ConfigOption that sets the priority of config files to the given value.
func WithConfigFilesPriority(priority int) ConfigOption {
	return func(o *ConfigOptions) {
		o.FilesPriority = priority
	}
}

// WithEnvVars returns a ConfigOption that enables or disables reading from environment variables.
func WithEnvVars(enabled bool) ConfigOption {
	return func(o *ConfigOptions) {
		o.EnvVars = enabled
	}
}

// WithEnvVarsPriority returns a ConfigOption that sets the priority of environment variables to the given value.
func WithEnvVarsPriority(priority int) ConfigOption {
	return func(o *ConfigOptions) {
		o.EnvVarsPriority = priority
	}
}

// WithEnvVarsPrefix returns a ConfigOption that sets the prefix for environment variables.
func WithEnvVarsPrefix(prefix string) ConfigOption {
	return func(o *ConfigOptions) {
		o.EnvVarsPrefix = prefix
	}
}

// WithTagName returns a ConfigOption that sets the tag name for struct fields.
// Empty strings are ignored.
func WithTagName(tagName string) ConfigOption {
	return func(o *ConfigOptions) {
		if tagName != "" {
			o.TagName = tagName
		}
	}
}

// readConfig implements the actual logic for reading configuration.
func readConfig[T any](opts *ConfigOptions) (_ *T, err error) {
	if opts == nil {
		opts = DefaultConfigOptions(context.Background())
	}

	var envCfg *T
	if opts.EnvVars {
		envCfg, err = readEnvConfig[T](opts)
		if err != nil {
			return nil, fail.Wrap(err, "failed to read environment variables")
		}
	}

	var fileCfgs []*T
	for _, path := range opts.Files {
		cfg, err := readFileConfig[T](path, opts)

		if err != nil {
			if opts.FilesRequired {
				return nil, fail.New().
					Attribute("path", path).
					Cause(err).
					Msg("failed to read config file")
			}
		} else {
			fileCfgs = append(fileCfgs, cfg)
		}
	}

	var allCfgs []*T
	if opts.EnvVarsPriority < opts.FilesPriority {
		allCfgs = append(allCfgs, append(fileCfgs, envCfg)...)
	} else {
		allCfgs = append(allCfgs, append([]*T{envCfg}, fileCfgs...)...)
	}

	var res T
	for _, cfg := range allCfgs {
		if err := mergo.Merge(&res, cfg, mergo.WithOverride); err != nil {
			return nil, fail.Wrap(err, "failed to merge config")
		}
	}

	return &res, nil
}

// readFileConfig reads configuration from the specified file path.
func readFileConfig[T any](path string, opts *ConfigOptions) (*T, error) {
	k := koanf.New(".")
	parsers := []koanf.Parser{
		yaml.Parser(),
		toml.Parser(),
		json.Parser(),
	}

	var (
		errs []error
		ok   bool
	)
	for _, parser := range parsers {
		if err := k.Load(file.Provider(path), parser); err != nil {
			errs = append(errs, err)
		} else {
			ok = true
			break
		}
	}

	if !ok {
		return nil, fail.New().
			CauseSlice(errs).
			Msg("failed to parse config file")
	}

	var t T
	if err := k.Unmarshal(".", &t); err != nil {
		return nil, fail.Wrap(err, "failed to unmarshal config file")
	}

	return &t, nil
}

// readEnvConfig reads configuration from environment variables.
func readEnvConfig[T any](opts *ConfigOptions) (*T, error) {
	k := koanf.New(".")
	_ = k
	return nil, nil
}
