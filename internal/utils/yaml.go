package utils

import (
	"bytes"
	"errors"
	"io"

	"gopkg.in/yaml.v3"
)

func YamlSplitMany(yamlBytes ...[]byte) ([][]byte, error) {
	results := [][]byte{}

	for _, yamlBytes := range yamlBytes {
		result, err := YamlSplit(yamlBytes)
		if err != nil {
			return nil, err
		}
		results = append(results, result...)
	}

	return results, nil
}

func YamlSplit(yamlBytes []byte) ([][]byte, error) {
	results := [][]byte{}
	reader := io.Reader(bytes.NewReader(yamlBytes))
	dec := yaml.NewDecoder(reader)

	for {
		var node yaml.Node
		err := dec.Decode(&node)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		document, err := yaml.Marshal(&node)
		if err != nil {
			return nil, err
		}
		results = append(results, document)
	}

	return results, nil
}
