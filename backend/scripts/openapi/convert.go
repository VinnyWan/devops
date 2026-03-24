package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"gopkg.in/yaml.v3"
)

const (
	inputPath      = "docs/swagger/swagger.json"
	outputJSONPath = "docs/openapi/openapi.json"
	outputYAMLPath = "docs/openapi/openapi.yaml"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	swaggerData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取 swagger 文件失败: %w", err)
	}

	var swaggerDoc openapi2.T
	if err := json.Unmarshal(swaggerData, &swaggerDoc); err != nil {
		return fmt.Errorf("解析 swagger JSON 失败: %w", err)
	}

	openapiDoc, err := openapi2conv.ToV3(&swaggerDoc)
	if err != nil {
		return fmt.Errorf("转换 OpenAPI3 失败: %w", err)
	}

	openapiJSON, err := json.MarshalIndent(openapiDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 OpenAPI3 JSON 失败: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputJSONPath), 0o755); err != nil {
		return fmt.Errorf("创建 OpenAPI 输出目录失败: %w", err)
	}
	if err := os.WriteFile(outputJSONPath, openapiJSON, 0o644); err != nil {
		return fmt.Errorf("写入 OpenAPI JSON 失败: %w", err)
	}

	openapiYAML, err := yaml.Marshal(openapiDoc)
	if err != nil {
		return fmt.Errorf("序列化 OpenAPI3 YAML 失败: %w", err)
	}
	if err := os.WriteFile(outputYAMLPath, openapiYAML, 0o644); err != nil {
		return fmt.Errorf("写入 OpenAPI YAML 失败: %w", err)
	}

	fmt.Printf("OpenAPI3 已生成: %s, %s\n", outputJSONPath, outputYAMLPath)
	return nil
}
