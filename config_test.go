package gofofa

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()

	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	tmpYAMLFile, err := os.CreateTemp("", "config-*.yaml")
	assert.Nil(t, err)
	defer os.Remove(tmpYAMLFile.Name())

	yamlContent := "categories:\n  - name: \"hard\"\n    filters:\n      - \"CONTAIN(category, '数据证书')\"\n\n  - name: \"soft\"\n    filters:\n      - \"CONTAIN(category, '其他支撑系统')\"\n\n  - name: \"buss\"\n    filters:\n      - \"CONTAIN(category, '电子邮件系统')\"\n      - \"CONTAIN(category, '其他企业应用')\"\ncustom_fields:\n  - name: \"custom\"\n    fields: ip,port,host,domain,protocol,status_code\n"
	_, err = tmpYAMLFile.WriteString(yamlContent)
	assert.Nil(t, err)
	tmpYAMLFile.Close()

	// 正确检测
	config, err := LoadConfig(tmpYAMLFile.Name())
	assert.Nil(t, err)
	assert.NotNil(t, config)

	// 错误检测
	_, err = LoadConfig("dsfhdksajfhsdkjfh")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read Config file failed")
}
