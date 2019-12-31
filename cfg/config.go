package cfg

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Port    string `json:"port"`
	Timeout int    `json:"timeout"`
	Mongo   struct {
		URL            string `json:"url"`
		Database       string `json:"database"`
		UserCollection string `json:"userCollection"`
		PostCollection string `json:"postCollection"`
	}
}

func GetConfig(path string) (c Config, err error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	err = json.Unmarshal(content, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}
