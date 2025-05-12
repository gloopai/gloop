package lib

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type conf struct{}

var Conf conf

/* 加载文件 */
func loadFileToBytes(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fileInfo.Size()
	buffer := make([]byte, fileSize)

	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func (c *conf) LoadTOML(path string, v interface{}) error {
	tomlData, err := loadFileToBytes(path)
	if err != nil {
		fmt.Println("Error loading file:", err)
		return err
	}

	fmt.Println("Loaded TOML data:", string(tomlData))

	// 解析 TOML 数据
	_, err = toml.Decode(string(tomlData), v)
	if err != nil {
		fmt.Println("Error decoding TOML:", err)
		return err
	}

	return nil
}
