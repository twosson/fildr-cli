package node

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"strings"
)

func readUintFromFile(path string) (uint64, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func bytesToString(byteArray []byte) string {
	n := bytes.IndexByte(byteArray, 0)
	if n < 0 {
		return string(byteArray)
	}
	return string(byteArray[:n])
}
