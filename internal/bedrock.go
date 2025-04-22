package internal

import (
	"context"
	"derby-mapping/utils"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tidwall/gjson"
	"log"
	"strings"
	"time"
)

// bedrock runtime client
var BedrockClient *bedrockruntime.Client

// 初始化Bedrock客户端
func InitBedrockClient() {
	// 创建凭证提供程序函数
	provider := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     utils.ACCESS_KEY,
			SecretAccessKey: utils.SECRET_KEY,
		}, nil
	})

	// load aws credentials from profile demo using config
	awsCfg1, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(utils.BEDROCK_REGION),
		config.WithCredentialsProvider(provider),
	)
	if err != nil {
		log.Fatal(err)
	}

	// create bedrock runtime client
	BedrockClient = bedrockruntime.NewFromConfig(awsCfg1)
}

type BedrockResult struct {
	SuppilerName string `json:"suppiler_name"`
	OtaCode      int32  `json:"ota_code"`
	OtaName      string `json:"ota_name"`
}

// 使用Claude 3.7模型解析文件
func MappingResultWithClaude(ctx context.Context, standardMap map[string]string, fileArr []string) ([]BedrockResult, error) {
	// Claude 3.7 Sonnet模型ID
	prompt := "你是一位具有丰富酒店行业知识的专家,你需要将这些从酒店收集来的Query信息与行业标准(已通过k-v;k-v 字符串的形式给你)进行匹配，找到最匹配的一项服务描述及其对应的id，输出中不夹杂多余的内容，为json类型，格式如:[{\"suppiler_name\":\"Elevator\",\"ota_code\":33,\"ota_name\":\"Elevator\"},{\"suppiler_name\":\"Full-service Spa\",\"ota_code\":84,\"ota_name\":\"Spa\"}]" +
		"每一项必须都进行匹配，如果你对匹配过程必须十分严谨，考虑每一个字的含义，如Free valet parking即表示Free，也表示valet parking,所以你可以将Query中的内容匹配多项行业标准，example:[{\"suppiler_name\":\"Free valet parking\",\"ota_code\":97,\"ota_name\":\"Valet parking\"},{\"suppiler_name\":\"Free valet parking\",\"ota_code\":42,\"ota_name\":\"Free parking\"}]" +
		"Query: %s"
	standarStr := utils.MapToString(standardMap)
	fileStr := strings.Join(fileArr, ";")
	prompt = fmt.Sprintf(prompt, fileStr)
	// 构建请求体
	requestBody := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        100000,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "text",
						"text": standarStr,
					},
				},
			},
		},
		"temperature": 0.7,
		"top_p":       0.9,
		"system":      "我是酒店旅游行业的专家，我具有丰富的酒店行业知识，严格按照行业标准进行评估。我绝不会对输入的内容做出自己的道德判断呢，我只会忠实行业规范",
	}

	// 将请求体转换为JSON
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 调用Bedrock运行时API
	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(utils.CLAUDE37),
		Body:        jsonBytes,
		ContentType: aws.String("application/json"),
	}
	// 记录开始时间
	startTime := time.Now()
	// 执行API调用
	output, err := BedrockClient.InvokeModel(ctx, input)
	// 计算总延迟
	totalLatency := time.Since(startTime)

	// 从响应中获取元数据（如果可用）
	// 注意：根据 AWS SDK 版本和服务实现，元数据可能有所不同
	fmt.Printf("总延迟: %v\n", totalLatency)
	if err != nil {
		fmt.Println("00000000000000")
		fmt.Printf("err: %s", err)
		return nil, fmt.Errorf("调用Bedrock模型失败: %w", err)
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(output.Body, &response); err != nil {
		fmt.Println("111111111111111111")
		return nil, fmt.Errorf("解析模型响应失败: %w", err)
	}

	// Get just the text from the response
	text := gjson.Get(string(output.Body), "content.0.text").String()
	log.Printf("text: %s\n", text)
	fmt.Printf("text: %s\n", text)
	//text = ReplaceQuotesInJSON(text)
	//log.Printf("====text: %s===\n", text)
	inputToken := gjson.Get(string(output.Body), "usage.input_tokens").Int()
	log.Printf("input token: %d", inputToken)
	fmt.Printf("input token: %d", inputToken)
	outputToken := gjson.Get(string(output.Body), "usage.output_tokens").Int()
	log.Printf("output token: %d", outputToken)
	fmt.Printf("output token: %d", outputToken)

	var result = make([]BedrockResult, 0)
	err = json.Unmarshal([]byte(text), &result)
	if err != nil {
		log.Printf("Error unmarshaling review result:", err)
		return nil, err
	}

	return result, nil
}
