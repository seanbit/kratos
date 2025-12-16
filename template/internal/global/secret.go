package global

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// 自定义错误类型
var (
	ErrNotDirectoryOrYaml = errors.New("path is neither a directory nor a yaml file")
	ErrSecretNotFound     = errors.New("secret.yaml not found")
	ErrInvalidPath        = errors.New("invalid path")
	ErrNotYamlFile        = errors.New("not a yaml file")
)

// SecretReader 封装secret读取逻辑
type SecretReader struct {
	// 可以添加配置项，比如允许的文件扩展名等
	allowedExtensions []string
}

// NewSecretReader 创建新的SecretReader实例
func NewSecretReader() *SecretReader {
	return &SecretReader{
		allowedExtensions: []string{".yaml", ".yml"},
	}
}

// WithExtensions 自定义允许的文件扩展名
func (sr *SecretReader) WithExtensions(extensions []string) *SecretReader {
	sr.allowedExtensions = extensions
	return sr
}

// GetSecretFilePath 获取secret.yaml文件路径
func (sr *SecretReader) GetSecretFilePath(path, fileName string) (string, error) {
	if !strings.HasSuffix(fileName, ".yaml") {
		fileName = fmt.Sprintf("%s.yaml", fileName)
	}
	// 规范化路径
	cleanPath := filepath.Clean(path)

	// 检查路径是否存在
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return "", fmt.Errorf("%w: %s", ErrInvalidPath, cleanPath)
	}

	// 获取路径信息
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to get path info: %w", err)
	}

	var secretPath string

	if fileInfo.IsDir() {
		// 如果是目录，读取目录下的secret.yaml
		secretPath = filepath.Join(cleanPath, fileName)
	} else {
		// 如果是文件，检查是否是yaml文件
		if !sr.isAllowedFile(cleanPath) {
			return "", fmt.Errorf("%w: %s", ErrNotYamlFile, cleanPath)
		}
		// 读取同级别的secret.yaml
		secretPath = filepath.Join(filepath.Dir(cleanPath), fileName)
	}

	// 检查secret.yaml是否存在
	if _, err := os.Stat(secretPath); os.IsNotExist(err) {
		return "", fmt.Errorf("%w: %s", ErrSecretNotFound, secretPath)
	}
	return secretPath, nil
}

// ReadSecret 读取secret.yaml文件内容
func (sr *SecretReader) ReadSecret(path, fileName string) ([]byte, error) {
	secretPath, err := sr.GetSecretFilePath(path, fileName)
	if err != nil {
		return nil, err
	}
	// 读取secret.yaml内容
	content, err := os.ReadFile(secretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret file: %w", err)
	}

	return content, nil
}

// isAllowedFile 检查文件是否是允许的类型
func (sr *SecretReader) isAllowedFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, allowedExt := range sr.allowedExtensions {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

// ReadSecretYaml 简化版的对外函数
func ReadSecretYaml(path, fileName string) ([]byte, error) {
	reader := NewSecretReader()
	return reader.ReadSecret(path, fileName)
}

func LoadSecretFromFile(flagconf, fileName string) error {
	reader := NewSecretReader()
	secretPath, err := reader.GetSecretFilePath(flagconf, fileName)
	if err != nil {
		return err
	}

	file, err := os.Open(secretPath)
	if err != nil {
		return errors.Wrap(err, "Open file error")
	}
	defer file.Close() // Ensure the file is closed when the function exits

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			return errors.Wrapf(err, "Skipping invalid line: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.TrimLeft(value, "\"")
		value = strings.TrimRight(value, "\"")

		err = os.Setenv(key, value)
		if err != nil {
			return errors.Wrap(err, "Setenv error")
		}
	}

	// Check for any errors encountered during scanning
	if err = scanner.Err(); err != nil {
		return errors.Wrap(err, "Error reading file")
	}

	fmt.Println("All local environment variables from the file have been set")
	return nil
}
