package config

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
)

func Load(file string) ([]Rule, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewBuffer(content))
	decoder.UseNumber()

	var rules []Rule
	err = decoder.Decode(&rules)
	return rules, errors.WithMessage(err, "Could not parse config file.")
}

func Persist(file string, rules []Rule) error {
	encoded, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return errors.WithMessage(err, "Could not serialize rules to json.")
	}

	err = ioutil.WriteFile(file, encoded, 0644)
	return errors.WithMessage(err, "Could not write serialized data to file.")
}
