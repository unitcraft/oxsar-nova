package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// writeYAMLSorted сериализует map[string]V в YAML с явно заданным
// порядком ключей. Используется конвертерами import-datasheets для
// получения детерминированного output (важно для git diff и code review).
func writeYAMLSorted[V any](path string, m map[string]V, sortFn func([]string), topKey string) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	if sortFn != nil {
		sortFn(keys)
	}
	node := &yaml.Node{Kind: yaml.MappingNode}
	innerNode := &yaml.Node{Kind: yaml.MappingNode}
	for _, k := range keys {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: k}
		valNode := &yaml.Node{}
		if err := valNode.Encode(m[k]); err != nil {
			return err
		}
		innerNode.Content = append(innerNode.Content, keyNode, valNode)
	}
	node.Content = []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: topKey},
		innerNode,
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return err
	}
	return enc.Close()
}
