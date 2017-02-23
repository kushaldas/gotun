package main

import ("testing"
	"strings")


func TestReadCommands(t *testing.T) {
	data := ReadCommands("./testcommands.txt")
	inside := []string{"go test", "ls", ""}

	for i, v := range data {
		if  v != inside[i] {
			t.Error("Reading commands file failed.", v, inside[i], i)
		}
	}
}

func TestWriteIPinformation(t *testing.T) {
	m := make(map[string]string)
	m["ip"] = "127.0.0.1"
	writeIPinformation(m)
}

func TestSystem(t *testing.T) {
	data, err := system("ls ./")
	if err != nil {
		t.Error("We have error in running ls command.", data)
	}
	if !strings.Contains(data, "main.go") {
		t.Error("Running ls command failed", data)
	}
}