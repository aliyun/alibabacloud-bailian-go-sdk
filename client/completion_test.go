/*
 * All rights Reserved, Designed By Alibaba Group Inc.
 * Copyright: Copyright(C) 1999-2023
 * Company  : Alibaba Group Inc.

 * @brief test cases for completion sdk
 * @author  yuanci.ytb
 * @version 1.0.0
 * @date 2023-11-07
 */

package broadscope_bailian_test

import (
	"encoding/json"
	"fmt"
	client "github.com/aliyun/alibabacloud-bailian-go-sdk/client"
	"github.com/google/uuid"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

/**
官方大模型调用应用、自训练模型应用示例
*/
func TestCreateModelCompletion(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")

	agentKey := os.Getenv("AGENT_KEY")
	appId := os.Getenv("APP_ID")

	//尽量避免多次初始化
	tokenClient := client.AccessTokenClient{AccessKeyId: accessKeyId, AccessKeySecret: accessKeySecret, AgentKey: agentKey}
	token, err := tokenClient.GetToken()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	cc := client.CompletionClient{Token: token}

	request := &client.CompletionRequest{
		AppId: appId,
		Messages: []client.ChatCompletionMessage{
			{Role: "system", Content: "你是一名历史学家, 帮助回答各种历史问题和历史知识"},
			{Role: "user", Content: "帮我生成一篇200字的文章，描述一下春秋战国的经济和文化"},
		},
		Parameters: &client.CompletionRequestModelParameter{ResultFormat: "message"},
	}

	response, err := cc.CreateCompletion(request)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	if !response.Success {
		log.Printf("failed to create completion, requestId: %s, code: %s, message: %s\n",
			response.RequestId, response.Code, response.Message)
		return
	}

	log.Printf("requestId: %s, content : %s\n", response.RequestId, response.Data.Choices[0].Message.Content)

	usages := response.Data.Usage
	if usages != nil && len(usages) > 0 {
		usage := usages[0]
		log.Printf("modelId: %s, inputTokens: %d, outputTokens: %d\n", usage.ModelId, usage.InputTokens, usage.OutputTokens)
	}
}

/**
官方大模型调用应用、自训练模型应用-其他参数使用示例
*/
func TestCreateCompletionWithParams(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")

	agentKey := os.Getenv("AGENT_KEY")
	appId := os.Getenv("APP_ID")

	//尽量避免多次初始化
	tokenClient := client.AccessTokenClient{AccessKeyId: accessKeyId, AccessKeySecret: accessKeySecret, AgentKey: agentKey}
	token, err := tokenClient.GetToken()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	cc := client.CompletionClient{Token: token}

	//设置超时时间
	cc.Timeout = 30 * time.Second

	request := &client.CompletionRequest{
		AppId: appId,
		//设置模型参数topP的值
		TopP: 0.2,
		//开启历史上下文, sessionId需要采用uuid保证唯一性, 后续传入相同sessionId，百炼平台将自动维护历史上下文
		SessionId: strings.ReplaceAll(uuid.New().String(), "-", ""),
		//设置历史上下文, 由调用侧维护历史上下文, 如果同时传入sessionId和history, 优先使用调用者管理的对话上下文
		Messages: []client.ChatCompletionMessage{
			{Role: "system", Content: "你是一个旅行专家, 能够帮我们制定旅行计划"},
			{Role: "user", Content: "我想去北京"},
			{Role: "assistant", Content: "北京是一个非常值得去的地方"},
			{Role: "user", Content: "那边有什么推荐的旅游景点"},
		},
		Parameters: &client.CompletionRequestModelParameter{
			//设置模型参数topK
			TopK: 50,
			//设置模型参数seed
			Seed: 2222,
			//设置模型参数temperature
			Temperature: 0.7,
			//设置模型参数max tokens
			MaxTokens: 20,
			//设置停止词
			Stop: []string{"景点"},
			//设置内容返回结构为message
			ResultFormat: "message",
		},
	}

	//调用文本生成接口
	response, err := cc.CreateCompletion(request)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	if !response.Success {
		log.Printf("failed to create completion, requestId: %s, code: %s, message: %s\n",
			response.RequestId, response.Code, response.Message)
		return
	}

	log.Printf("requestId: %s, text: %s\n", response.RequestId, response.Data.Choices[0].Message.Content)
}

/**
流式响应使用示例
*/
func TestCreateStreamCompletion(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")

	agentKey := os.Getenv("AGENT_KEY")
	appId := os.Getenv("APP_ID")

	//尽量避免多次初始化
	tokenClient := client.AccessTokenClient{AccessKeyId: accessKeyId, AccessKeySecret: accessKeySecret, AgentKey: agentKey}
	token, err := tokenClient.GetToken()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	cc := client.CompletionClient{Token: token}

	request := &client.CompletionRequest{
		AppId: appId,
		Messages: []client.ChatCompletionMessage{
			{Role: "system", Content: "你是一名天文学家, 能够帮助小学生回答宇宙与天文方面的问题"},
			{Role: "user", Content: "宇宙中为什么会存在黑洞"},
		},
		Parameters: &client.CompletionRequestModelParameter{
			//返回choice message结果
			ResultFormat: "message",
			//开启增量输出模式，后面输出不会包含已经输出的内容
			IncrementalOutput: true,
		},
	}

	response, err := cc.CreateStreamCompletion(request)
	if err != nil {
		log.Printf("failed to create completion, err: %v\n", err)
		return
	}

	for result := range response {
		if !result.Success {
			log.Printf("get result with error, requestId: %s, code: %s, message: %s\n",
				result.RequestId, result.Code, result.Message)
		} else {
			fmt.Printf("%s", result.Data.Choices[0].Message.Content)
		}
	}
	fmt.Printf("\n")
}

/**
三方模型应用示例
*/
func TestCreateThirdModelCompletion(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")

	agentKey := os.Getenv("AGENT_KEY")
	appId := os.Getenv("APP_ID")

	//尽量避免多次初始化
	tokenClient := client.AccessTokenClient{AccessKeyId: accessKeyId, AccessKeySecret: accessKeySecret, AgentKey: agentKey}
	token, err := tokenClient.GetToken()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	cc := client.CompletionClient{Token: token}

	prompt := "那边有什么推荐的旅游景点"
	request := &client.CompletionRequest{
		AppId:  appId,
		Prompt: prompt,
		History: []client.ChatQaMessage{
			{User: "我想去北京", Bot: "北京是一个非常值得去的地方"},
		},
	}

	//调用文本生成接口
	response, err := cc.CreateCompletion(request)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	if !response.Success {
		log.Printf("failed to create completion, requestId: %s, code: %s, message: %s\n",
			response.RequestId, response.Code, response.Message)
		return
	}

	log.Printf("requestId: %s, text: %s\n", response.RequestId, response.Data.Text)
}

/**
检索增强应用示例
*/
func TestCreateRagAppCompletion(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")

	agentKey := os.Getenv("AGENT_KEY")
	appId := os.Getenv("APP_ID")

	//尽量避免多次初始化
	tokenClient := client.AccessTokenClient{AccessKeyId: accessKeyId, AccessKeySecret: accessKeySecret, AgentKey: agentKey}
	token, err := tokenClient.GetToken()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	cc := client.CompletionClient{Token: token}

	prompt := "API接口说明中, TopP参数改如何传递?"
	request := &client.CompletionRequest{
		AppId:  appId,
		Prompt: prompt,
		History: []client.ChatQaMessage{
			{User: "API接口如何使用", Bot: "API接口需要传入prompt、app id并通过post方法调用"},
		},
		// 返回文档检索的文档引用数据, 传入为simple或indexed
		DocReferenceType: client.DocReferenceTypeSimple,
		// 文档标签code列表
		DocTagCodes: []string{"471d*******3427", "881f*****0c232"},
	}

	//调用文本生成接口
	response, err := cc.CreateCompletion(request)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	if !response.Success {
		log.Printf("failed to create completion, requestId: %s, code: %s, message: %s\n",
			response.RequestId, response.Code, response.Message)
		return
	}

	log.Printf("requestId: %s, text: %s\n", response.RequestId, response.Data.Text)
	if response.Data.DocReferences != nil && len(response.Data.DocReferences) > 0 {
		log.Printf("Doc ref: %s\n", response.Data.DocReferences[0].DocName)
	}
}

/**
插件和流程编排应用示例
*/
func TestCreateFLowAppCompletion(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")

	agentKey := os.Getenv("AGENT_KEY")
	appId := os.Getenv("APP_ID")

	//尽量避免多次初始化
	tokenClient := client.AccessTokenClient{AccessKeyId: accessKeyId, AccessKeySecret: accessKeySecret, AgentKey: agentKey}
	token, err := tokenClient.GetToken()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	cc := client.CompletionClient{Token: token}

	bizParams := "{\"userId\": \"123\"}"
	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(bizParams), &data)
	if err != nil {
		log.Printf("failed to parse biz param, json: %s, err: %v\n", bizParams, err)
		return
	}

	prompt := "今天杭州的天气怎么样"
	request := &client.CompletionRequest{
		AppId:     appId,
		Prompt:    prompt,
		BizParams: data,
	}

	//调用文本生成接口
	response, err := cc.CreateCompletion(request)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	if !response.Success {
		log.Printf("failed to create completion, requestId: %s, code: %s, message: %s\n",
			response.RequestId, response.Code, response.Message)
		return
	}

	log.Printf("requestId: %s, text: %s\n", response.RequestId, response.Data.Text)
}

/**
智能问数应用示例
*/
func TestCreateNl2SqlCompletion(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")

	agentKey := os.Getenv("AGENT_KEY")
	appId := os.Getenv("APP_ID")

	//尽量避免多次初始化
	tokenClient := client.AccessTokenClient{AccessKeyId: accessKeyId, AccessKeySecret: accessKeySecret, AgentKey: agentKey}
	token, err := tokenClient.GetToken()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	cc := client.CompletionClient{Token: token}

	//自然语言转sql调用示例
	sqlSchema := "{" +
		"    \"sqlInput\": {" +
		"      \"synonym_infos\": \"国民生产总值: GNP|Gross National Product\"," +
		"      \"schema_infos\": [" +
		"        {" +
		"          \"columns\": [" +
		"            {" +
		"              \"col_caption\": \"地区\"," +
		"              \"col_name\": \"region\"" +
		"            }," +
		"            {" +
		"              \"col_caption\": \"年份\"," +
		"              \"col_name\": \"year\"" +
		"            }," +
		"            {" +
		"              \"col_caption\": \"国民生产总值\"," +
		"              \"col_name\": \"gross_national_product\"" +
		"            }" +
		"          ]," +
		"          \"table_id\": \"t_gross_national_product_1\"," +
		"          \"table_desc\": \"国民生产总值表\"" +
		"        }" +
		"      ]" +
		"    }" +
		"  }"
	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(sqlSchema), &data)
	if err != nil {
		log.Printf("failed to parse sql schema, json: %s, err: %v\n", sqlSchema, err)
		return
	}

	prompt := "浙江近五年GNP总和是多少"
	request := &client.CompletionRequest{
		AppId:     appId,
		Prompt:    prompt,
		BizParams: data,
	}

	//调用文本生成接口
	response, err := cc.CreateCompletion(request)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	if !response.Success {
		log.Printf("failed to create completion, requestId: %s, code: %s, message: %s\n",
			response.RequestId, response.Code, response.Message)
		return
	}

	log.Printf("requestId: %s, text: %s\n", response.RequestId, response.Data.Text)
}
