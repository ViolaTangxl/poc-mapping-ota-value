package utils

import (
	"fmt"
	"os"
	"strings"
)

// 敏感信息通过环境变量或配置文件加载
var (
	ACCESS_KEY      = getEnvOrDefault("AWS_ACCESS_KEY", "")
	SECRET_KEY      = getEnvOrDefault("AWS_SECRET_KEY", "")
	FEISHU_WEB_HOOK = getEnvOrDefault("FEISHU_WEB_HOOK", "")
	BEDROCK_REGION  = getEnvOrDefault("BEDROCK_REGION", "us-east-1")
	CLAUDE37        = getEnvOrDefault("CLAUDE_MODEL", "us.anthropic.claude-3-7-sonnet-20250219-v1:0")
)

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// MapToString 将map[string]string转换为"k-value;k-value"格式的字符串
func MapToString(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	pairs := make([]string, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s-%s", k, v))
	}

	return strings.Join(pairs, ";")
}
