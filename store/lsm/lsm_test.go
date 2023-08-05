package lsm_test

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func TestMakeNewFileFromMemtable(t *testing.T) {
	mem := map[string]string{
		"x": "X",
		"y": "Y",
	}
	file, err := os.OpenFile("./tmp.lsm", os.O_CREATE|os.O_RDWR, 0766)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	/*
	  文件内格式：
	  |k|v|...|index|footer|

	  |index|
	  |key1|offset|key2|offset|key3|offset|

	  |footer|
	  |index_start| 共64个字节
	*/
	bw := bufio.NewWriter(file)
	indexs := make(map[string]uint64)
	var currentNum uint64
	for k, v := range mem {
		bw.WriteString(k)
		currentNum += uint64(len(k))
		bw.WriteString(v)
		currentNum += uint64(len(v))
		indexs[k] = currentNum
	}
	var indexStartNum = currentNum
	for k, v := range indexs {
		bw.WriteString(k)
		bw.WriteString(fmt.Sprint(v))
	}
	bw.WriteString(fmt.Sprint(indexStartNum))
	bw.Flush()
}

func TestGetIndexFromFile(t *testing.T) {
	file, err := os.OpenFile("./tmp.lsm", os.O_RDWR, 0766)
	if err != nil {
		panic(err)
	}

}
