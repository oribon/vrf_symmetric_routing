package main

import (
	"os"
	"text/template"

	"gopkg.in/yaml.v2"
)

type config struct {
	Intf    string `yaml:"intf"`
	Worker0 struct {
		NodeName string `yaml:"nodeName"`
		IP       string `yaml:"ip"`
	} `yaml:"worker0"`
	Worker1 struct {
		NodeName string `yaml:"nodeName"`
		IP       string `yaml:"ip"`
	} `yaml:"worker1"`
	ExternalHostIP string `yaml:"externalHostIP"`
	SecondaryNetGW string `yaml:"secondaryNetGW"`
}

func main() {
	conf := &config{}

	confFile, err := os.ReadFile("conf.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(confFile, conf)
	if err != nil {
		panic(err)
	}

	t1, err := template.New("frr.conf.tpl").ParseFiles("frr.conf.tpl")
	if err != nil {
		panic(err)
	}
	f1, err := os.Create("./frr/frr.conf")
	if err != nil {
		panic(err)
	}
	defer f1.Close()
	err = t1.Execute(f1, conf)
	if err != nil {
		panic(err)
	}

	t2, err := template.New("metallb.yaml.tpl").ParseFiles("metallb.yaml.tpl")
	if err != nil {
		panic(err)
	}
	f2, err := os.Create("metallb.yaml")
	if err != nil {
		panic(err)
	}
	defer f2.Close()
	err = t2.Execute(f2, conf)
	if err != nil {
		panic(err)
	}

	t3, err := template.New("vrf-nncps.yaml.tpl").ParseFiles("vrf-nncps.yaml.tpl")
	if err != nil {
		panic(err)
	}
	f3, err := os.Create("vrf-nncps.yaml")
	if err != nil {
		panic(err)
	}
	defer f3.Close()
	err = t3.Execute(f3, conf)
	if err != nil {
		panic(err)
	}
}
