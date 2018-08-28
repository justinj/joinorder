package main

import (
	"io/ioutil"
	"testing"

	"github.com/justinj/joinorder/datadriven"
	"github.com/justinj/joinorder/queries"
)

// this will be too much effort for now.

const path = "/Users/justin/go/src/github.com/justinj/joinorder/testdata"

func parseSchema(input string) {

}

func TestData(t *testing.T) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		datadriven.RunTest(t, path+"/"+file.Name(), func(d *datadriven.TestData) string {
			switch d.Cmd {
			case "run":
				s := queries.QueryByName("bushy")
				orderer := NewIKKBZOrderer(s)
				return orderer.Order().String() + "\n"
			}
			panic("unknown command " + d.Cmd)
		})
	}
}
