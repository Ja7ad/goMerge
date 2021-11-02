package goMerge

import (
	"fmt"
	"io/ioutil"
	"testing"
)

const numOfFile = 5

func Test_Merge(t *testing.T) {
	// create file with random data
	for i := 0; i < numOfFile; i++ {
		buf := make([]byte, 10000)
		err := ioutil.WriteFile(fmt.Sprintf("./test/test-%d.txt", i), buf, 0666)
		if err != nil {
			t.Error(err)
		}
	}

	err := Merge("./test/", ".txt", "./test/merged.txt", true)
	if err != nil {
		t.Error(err)
	}
}
