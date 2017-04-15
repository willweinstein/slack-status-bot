package main

import (
	"io/ioutil"
	"os"
)

const Storage_Path = "_private/"

func Storage_Init() {
	_, err := os.Stat(Storage_Path)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(Storage_Path, 0777)
		} else {
			panic(err)
		}
	}
}

func Storage_Get(key string) (string) {
	f, err := ioutil.ReadFile(Storage_Path + key)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		} else {
			panic(err)
		}
	}
	return string(f)
}

func Storage_Set(key string, value string) {
	ioutil.WriteFile(Storage_Path + key, []byte(value), 0777)
}