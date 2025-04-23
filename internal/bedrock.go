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
	prompt := "You are an expert with extensive knowledge of the hotel industry, and you need to match this Query information collected from hotels with the industry standard (which has been given to you in the form of a k-v;k-v string), and find the best match for one of the service descriptions and their corresponding ids, with no redundancy in the output, which is of json type.example:[{\"suppiler_name\":\"Elevator\",\"ota_code\":33,\"ota_name\":\"Elevator\"},{\"suppiler_name\":\"Full-service Spa\",\"ota_code\":84,\"ota_name\":\"Spa\"}]" +
		"Each item must be matched, and if you are very critical of the matching process you must be very rigorous and consider the meaning of each word, such as:Free valet parking means Free, but also valet parking, so you can match the contents of the Query to multiple industry standards.example:[{\"suppiler_name\":\"Free valet parking\",\"ota_code\":97,\"ota_name\":\"Valet parking\"},{\"suppiler_name\":\"Free valet parking\",\"ota_code\":42,\"ota_name\":\"Free parking\"}]" +
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
		"system":      "I am an expert in the hospitality and travel industry, I have extensive knowledge of the hospitality industry and I strictly follow industry standards in my assessments. I will never make my own moral judgments about input, I will only be faithful to industry norms",
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
		fmt.Printf("err: %s", err)
		return nil, fmt.Errorf("调用Bedrock模型失败: %w", err)
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return nil, fmt.Errorf("解析模型响应失败: %w", err)
	}

	// Get just the text from the response
	text := gjson.Get(string(output.Body), "content.0.text").String()
	log.Printf("text: %s\n", text)
	//text = ReplaceQuotesInJSON(text)
	//log.Printf("====text: %s===\n", text)
	inputToken := gjson.Get(string(output.Body), "usage.input_tokens").Int()
	log.Printf("input token: %d", inputToken)
	outputToken := gjson.Get(string(output.Body), "usage.output_tokens").Int()
	log.Printf("output token: %d", outputToken)

	var result = make([]BedrockResult, 0)
	err = json.Unmarshal([]byte(text), &result)
	if err != nil {
		log.Printf("Error unmarshaling review result:", err)
		return nil, err
	}

	return result, nil
}
