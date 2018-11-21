/**
文件系统相关辅助方法
*/

package tools

import (
	"fmt"
	"os"
	"syscall"
)

// Exists 判断所给路径文件/文件夹是否存在
func Exists(file string) bool {
	_, err := os.Stat(file) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}

		return false
	}

	return true
}

// Writable 文件/文件夹是否可写
func Writable(file string) (bool, error) {
	if !Exists(file) {
		err := fmt.Errorf("file does not exist: `%s`", file)
		return false, err
	}

	err := syscall.Access(file, syscall.O_RDWR)
	if err != nil {
		return false, err
	}

	return true, nil
}

// IsDir 判断所给路径是否为文件夹
func IsDir(file string) bool {
	s, err := os.Stat(file)
	if err != nil {
		return false
	}

	return s.IsDir()
}

// IsFile 判断所给路径是否为文件
func IsFile(file string) bool {
	stat, err := os.Stat(file)
	if err != nil {
		return false
	}

	fm := stat.Mode()
	return fm.IsRegular()
}
