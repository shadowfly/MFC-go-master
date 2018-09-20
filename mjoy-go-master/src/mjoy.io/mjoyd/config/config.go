////////////////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 The mjoy-go Authors.
//
// The mjoy-go is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// @File: config.go
// @Date: 2018/05/10 17:44:10
////////////////////////////////////////////////////////////////////////////////

package config

import (
	"github.com/naoina/toml"
	"reflect"
	"fmt"
	"unicode"
	"os"
	"bufio"
	"errors"
	"path/filepath"
	"io"
	"sync"
	"mjoy.io/mjoyd/defaults"
)

// interface for module config
type Iconfig interface {
	toml.UnmarshalerRec
	toml.MarshalerRec
	SetDefaultConfig() (error)
}

type Config struct{
	path string
	configs map[string] Iconfig
}

var c *Config
var once sync.Once
//singleton pattern for Config
func GetConfigInstance() *Config {
	once.Do(func(){
		c =&Config{
			path: defaults.DefaultTOMLConfigPath,
			configs: make(map[string]Iconfig),
		}
	})
	return c
}

func (c *Config) SetPath(path string) {
	c.path = path
}

//module register a config
func (c *Config) Register(name string, config Iconfig) error {
	if _, ok := c.configs[name]; ok {
		logger.Error("module",name,"is already registered")
		return errors.New("config is already registered")
	}

	path := filepath.Join(c.path, name + ".toml")

	if err := loadConfig(path, config); err != nil{
		return err
	}
	c.configs[name] = config
	//c.dumpAllconfig()
	return nil
}

func (c *Config) Unregister(name string) error {
	if _, ok := c.configs[name]; !ok {
		logger.Error("module",name,"is not registered")
		return errors.New("config is not registered")
	}
	delete(c.configs, name)
	logger.Info("module",name,"is unregistered !")
	return nil
}

func loadConfig(file string, config Iconfig) error {
	f, err := os.Open(file)
	if err != nil {
		logger.Info("config file", file, "is not exit, use default config")
		return config.SetDefaultConfig()
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(config)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

func (c *Config) dumpAllconfig() error {
	for name, config := range c.configs {
		out, err := tomlSettings.Marshal(config)
		if err != nil {
			return err
		}
		io.WriteString(os.Stdout, "\n**********"+ name +"***********\n")
		os.Stdout.Write(out)
	}
	return nil
}