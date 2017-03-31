package main

import ("testing"
	"io/ioutil"
	"os"
	"strings"
)

func TestOpenStack(t *testing.T) {
	if testing.Short() {
        	t.Skip("skipping integration test")
    	}
	tmpfile, _ := ioutil.TempFile("", "randomtestrun")
	defer os.Remove(tmpfile.Name())
	s := os.Stdout
	os.Stdout = tmpfile
	retcode := starthere("testcommands", "./", "", "")
	os.Stdout = s
	b, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Error("Error in reading logs", err)
	}
	data := string(b)
	if retcode != 0 {
		t.Error("Test failed.", retcode, data)
	}
	if !strings.Contains(data, "Executing:  cat /etc/os-release") {
		t.Error("Missing cat /etc/os-release", data)
	}

}

func TestAWS(t *testing.T) {
	if testing.Short() {
        	t.Skip("skipping integration test")
    	}
	tmpfile, _ := ioutil.TempFile("", "randomtestrun")
	defer os.Remove(tmpfile.Name())
	s := os.Stdout
	os.Stdout = tmpfile
	retcode := starthere("testaws", "./", "", "")
	os.Stdout = s
	b, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Error("Error in reading logs", err)
	}
	data := string(b)
	if retcode != 0 {
		t.Error("Test failed.", retcode, data)
	}
	if !strings.Contains(data, "Executing:  cat /etc/os-release") {
		t.Error("Missing cat /etc/os-release", data)
	}

}
