package main

import (
	"io/ioutil"
	"testing"

	"github.com/justinj/joinorder/datadriven"
)

// this will be too much effort for now.

const path = "/Users/justin/go/src/github.com/justinj/joinorder/testdata"

func parseSchema(input string) {

}

func TestData(t *testing.T) {
	t.Skip()
	files, err := ioutil.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		datadriven.RunTest(t, path+"/"+file.Name(), func(d *datadriven.TestData) string {
			switch d.Cmd {
			case "ikkbz":
				return d.Input
			}
			panic("unknown command " + d.Cmd)
		})
	}
}
