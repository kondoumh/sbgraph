package file

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

func CreateDir(path string) error {
	if f, err := os.Stat(path); os.IsNotExist(err) || !f.IsDir() {
		if err := os.MkdirAll(path, 0775); err != nil {
			return err
		}
	}
	return nil
}

func WriteBytes(data []byte, fileName string, outDir string) error {
	dir, err := os.Stat(outDir)
	if os.IsNotExist(err) || !dir.IsDir() {
		return err
	}
	file, err := os.Create(outDir + "/" + fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	var pj bytes.Buffer
	json.Indent(&pj, []byte(data), "", " ")
	file.Write(pj.Bytes())
	return nil
}

func ReadBytes(fileName string, outDir string) ([]byte, error) {
	raw, err := ioutil.ReadFile(outDir + "/" + fileName)
	if err != nil {
		return nil, err
	}
	return raw, err
}