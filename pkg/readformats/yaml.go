package readformats

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const maxCapacity = 2048 * 1024

type YAMLReader struct {
	Filename string
}

func NewYAMLReader(filename string) *YAMLReader {
	return &YAMLReader{Filename: filename}
}

func (y *YAMLReader) ReadFile() ([]byte, error) {
	file, err := os.Open(y.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %v", y.Filename, err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	buf := make([]byte, maxCapacity)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(buf, maxCapacity)
	lineNumber := 1
	for scanner.Scan() {
		buffer.Write(scanner.Bytes())
		buffer.WriteByte('\n')
		lineNumber++
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("something bad happened in the line %v: %v", lineNumber, err)
	}

	return buffer.Bytes(), nil
}

func (y *YAMLReader) UnmarshalFile(target interface{}) error {
	data, err := y.ReadFile()
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal YAML file %q: %v", y.Filename, err)
	}
	return nil
}
