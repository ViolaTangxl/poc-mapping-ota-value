package main

import (
	"context"
	"derby-mapping/internal"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func main() {
	internal.InitBedrockClient()

	// 设置路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)

	// 启动服务器
	port := "8080"
	fmt.Printf("服务器启动在 http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// indexHandler 处理主页请求
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filepath.Join("templates", "index.html"))
}
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("开始上传请求......")

	// 设置CORS头部
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理OPTIONS请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 检查请求方法
	if r.Method != "POST" {
		http.Error(w, "只允许POST请求", http.StatusMethodNotAllowed)
		return
	}

	// 解析multipart form数据
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "解析表单数据失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 1. 解析第一个文件 (JSON格式)
	file1, _, err := r.FormFile("txtFile1")
	if err != nil {
		http.Error(w, "无法获取第一个上传文件: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file1.Close()

	// 2. 解析第二个文件 (行业标准)
	file2, _, err := r.FormFile("txtFile2")
	if err != nil {
		http.Error(w, "无法获取第二个上传文件: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file2.Close()

	// 3. 读取第一个文件内容
	fileContent1, err := io.ReadAll(file1)
	if err != nil {
		http.Error(w, "读取第一个文件错误: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. 读取第二个文件内容
	fileContent2, err := io.ReadAll(file2)
	if err != nil {
		http.Error(w, "读取第二个文件错误: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. 解析第一个文件的JSON数据
	var items []string
	if err := json.Unmarshal(fileContent1, &items); err != nil {
		http.Error(w, "无法解析JSON内容: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 6. 使用 map 来存储唯一的项，实现去重
	uniqueItems := make(map[string]bool)
	for _, item := range items {
		// 提取引号内的内容并去除空白
		trimmedItem := strings.TrimSpace(item)
		if trimmedItem != "" {
			uniqueItems[trimmedItem] = true
		}
	}

	// 7. 将去重后的数据转换为切片
	var result []string
	for item := range uniqueItems {
		result = append(result, item)
	}

	// 8. 解析第二个文件 (行业标准)
	content2 := string(fileContent2)
	lines2 := strings.Split(content2, "\n")

	// 9. 创建行业标准映射 (内容 -> ID)
	standardMap := make(map[string]string)
	for _, line := range lines2 {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // 跳过空行
		}

		// 分割行，获取ID和内容
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 {
			id := strings.TrimSpace(parts[0])
			content := strings.TrimSpace(parts[1])
			standardMap[content] = id
		}
	}
	// TODO 先进行字符完全匹配，然后再将无法完全匹配的内容调用模型

	// 调用模型进行匹配
	bedrockResult, err := internal.MappingResultWithClaude(context.Background(), standardMap, result)
	if err != nil {
		http.Error(w, "调用模型失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 10. 匹配结果
	//matches := make(map[string]string)
	//for _, item := range result {
	//	if id, exists := standardMap[item]; exists {
	//		matches[item] = id
	//	}
	//}
	//// 11. 使用fmt输出结果
	//fmt.Println("处理后的结果:")
	//for _, item := range result {
	//	fmt.Println(item)
	//}
	//
	//fmt.Println("\n匹配的行业标准ID:")
	//for item, id := range matches {
	//	fmt.Printf("%s -> %s\n", item, id)
	//}

	// 12. 向客户端返回成功消息
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  "success",
		"message": "文件处理成功",
		"count":   len(result),
		"matches": bedrockResult,
	}
	json.NewEncoder(w).Encode(response)
}
