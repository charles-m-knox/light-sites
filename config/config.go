package config

import (
	"lightsites/constants"

	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type DirectoriesConfig struct {
	Assets    string `yaml:"assets"`
	Documents string `yaml:"documents"`
	Templates string `yaml:"templates"`
}

type RoutingConfig struct {
	RoutePrefix   string `yaml:"routePrefix"`
	AssetsPrefix  string `yaml:"assetsPrefix"`
	UrlFileSuffix string `yaml:"urlFileSuffix"`
}

type BodyConfig struct {
	ContainerClass string `yaml:"containerClass"`
	RowClass       string `yaml:"rowClass"`
	ColClass       string `yaml:"colClass"`
}

type NodeRule struct {
	Class string `yaml:"class"`
	Style string `yaml:"style"`
}

type Config struct {
	RefreshInterval time.Duration                `yaml:"refreshInterval"`
	Directories     DirectoriesConfig            `yaml:"directories"`
	Routing         RoutingConfig                `yaml:"routing"`
	CSSImports      []string                     `yaml:"cssImports"`
	BodyConfig      BodyConfig                   `yaml:"bodyConfig"`
	Rules           map[string]map[string]string `yaml:"rules"`
	ListenAddr      string                       `yaml:"listenAddr"`
}

// LoadConfig reads from a provided yaml-formatted configuration filename
func LoadConfig() (conf Config, err error) {
	// read from config file
	confData, err := ioutil.ReadFile(constants.ConfigFile)
	if err != nil {
		return conf, fmt.Errorf("failed to read config file %v: %v", constants.ConfigFile, err.Error())
	}

	err = yaml.Unmarshal(confData, &conf)
	if err != nil {
		return conf, fmt.Errorf("failed to parse config file %v: %v", constants.ConfigFile, err.Error())
	}

	return conf, nil
}

// GetDefaultConfig returns a basic sample configuration and
// is mainly used for unit testing
func GetDefaultConfig() Config {
	return Config{
		RefreshInterval: time.Duration(30 * time.Minute),
		Directories: DirectoriesConfig{
			Assets:    "./src/assets",
			Documents: "./src/content",
			Templates: "./src/templates",
		},
		Routing: RoutingConfig{
			RoutePrefix:   "/content/",
			AssetsPrefix:  "/assets/",
			UrlFileSuffix: ".html",
		},
		CSSImports: []string{
			"bootstrap.min.css",
			"custom.css",
		},
		BodyConfig: BodyConfig{
			ContainerClass: constants.ContainerClass,
			RowClass:       constants.RowClass,
			ColClass:       constants.ColClass,
		},
		Rules: map[string]map[string]string{
			constants.TableNode: {
				constants.ClassAttribute: constants.TableClasses,
			},
			constants.ImgNode: {
				constants.StyleAttribute: constants.ImgStyles,
			},
		},
		ListenAddr: ":8099",
	}
}
