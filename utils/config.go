package utils

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"
)

type ListEntry struct {
	Name    string
	Command string
}

type Config struct {
	file string
	data []ListEntry
}

func NewConfig() (*Config, error) {
	conf := &Config{
		file: "config.json",
		data: make([]ListEntry, 0),
	}
	err := conf.load()
	if err != nil {
		return nil, err
	}

	err = conf.Save()
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func (c *Config) load() error {
	f, err := os.OpenFile(c.file, os.O_RDONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(&c.data)
	if err != nil && errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func (c *Config) Save() error {
	f, err := os.OpenFile(c.file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(&c.data)
}

func (c *Config) GetList() []ListEntry {
	return c.data
}

func (c *Config) AddEntry(entry ListEntry) {
	c.data = append(c.data, entry)
	c.Save()
}

func (c *Config) RemoveEntry(e *ListEntry) {
	c.data = slices.DeleteFunc(c.data, func(ce ListEntry) bool {
		return ce.Name == e.Name && ce.Command == e.Command
	})
	c.Save()
}
