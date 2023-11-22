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
	apiClient "github.com/alibabacloud-go/bailian-20230601/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	client "github.com/aliyun/alibabacloud-bailian-go-sdk/client"
	"github.com/google/uuid"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCreateCompletion(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")

	agentKey := os.Getenv("AGENT_KEY")
	appId := os.Getenv("APP_ID")

	//尽量避免多次初始化
	tokenClient := client.AccessTokenClient{AccessKeyId: &accessKeyId, AccessKeySecret: &accessKeySecret, AgentKey: &agentKey}
	token, err := tokenClient.GetToken()
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	cc := client.CompletionClient{Token: &token}
	prompt := "帮我生成一篇200字的文章，描述一下春秋战国的政治、军事和经济"

	request := &client.CompletionRequest{}
	request.SetAppId(appId)
	request.SetPrompt(prompt)

	response, err := cc.CreateCompletion(request)
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	if !response.Success {
		t.Errorf("failed to create completion, requestId: %s, code: %s, message: %s\n",
			response.GetRequestId(), response.GetCode(), response.GetMessage())
		return
	}

	t.Logf("requestId: %s, text : %s\n", response.GetRequestId(), response.GetData().GetText())
}

func TestCreateStreamCompletion(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")

	agentKey := os.Getenv("AGENT_KEY")
	appId := os.Getenv("APP_ID")

	//尽量避免多次初始化
	tokenClient := client.AccessTokenClient{AccessKeyId: &accessKeyId, AccessKeySecret: &accessKeySecret, AgentKey: &agentKey}
	token, err := tokenClient.GetToken()
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	cc := client.CompletionClient{Token: &token}
	prompt := "帮我生成一篇500字的文章，描述一下春秋战国的政治、军事和经济"

	request := &client.CompletionRequest{}
	request.SetAppId(appId)
	request.SetPrompt(prompt)

	response, err := cc.CreateStreamCompletion(request)
	if err != nil {
		t.Errorf("failed to create completion, err: %v\n", err)
		return
	}

	for result := range response {
		if !result.Success {
			t.Errorf("get result with error, requestId: %s, code: %s, message: %s\n", result.GetRequestId(),
				result.GetCode(), result.GetMessage())
		} else {
			t.Logf("requestId: %s, text: %s\n", result.GetRequestId(), result.GetData().GetText())
		}
	}
}

func TestCreateCompletionWithParams(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")

	agentKey := os.Getenv("AGENT_KEY")
	appId := os.Getenv("APP_ID")

	//尽量避免多次初始化
	tokenClient := client.AccessTokenClient{AccessKeyId: &accessKeyId, AccessKeySecret: &accessKeySecret, AgentKey: &agentKey}
	token, err := tokenClient.GetToken()
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	cc := client.CompletionClient{Token: &token}

	//设置超时时间
	cc.SetTimeout(30 * time.Second)

	prompt := "云南近5年GNP总和是多少"

	request := &client.CompletionRequest{}
	request.SetAppId(appId)
	request.SetPrompt(prompt)

	//设置模型参数topP的值
	request.SetTopP(0.2)

	//开启历史上下文, sessionId需要采用uuid保证唯一性, 后续传入相同sessionId，百炼平台将自动维护历史上下文
	sessionId := strings.ReplaceAll(uuid.New().String(), "-", "")
	request.SetSessionId(sessionId)

	//设置历史上下文, 由调用侧维护历史上下文, 如果同时传入sessionId和history, 优先使用调用者管理的对话上下文
	message1 := &client.ChatQaMessage{User: "我想去北京", Bot: "北京的天气很不错"}
	message2 := &client.ChatQaMessage{User: "北京有什么旅游景点", Bot: "北京有故宫、天坛、长城等"}
	chatHistory := []*client.ChatQaMessage{message1, message2}
	request.SetHistory(chatHistory)

	//设置模型参数topK，seed
	modelParameter := &client.CompletionRequestModelParameter{TopK: 50, Seed: 2222, UseRawPrompt: true}
	request.SetParameters(modelParameter)

	//设置文档标签tagId，设置后，文档检索召回时，仅从tagIds对应的文档范围进行召回
	request.SetDocTagIds([]int64{100, 101})

	//返回文档检索的文档引用数据
	request.SetDocReferenceType(client.DocReferenceTypeSimple)

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
		t.Errorf("failed to parse sql schema, json: %s, err: %v\n", sqlSchema, err)
		return
	}
	request.SetBizParams(&data)

	//调用文本生成接口
	response, err := cc.CreateCompletion(request)
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	if !response.Success {
		t.Errorf("failed to create completion, requestId: %s, code: %s, message: %s\n", response.GetRequestId(),
			response.GetCode(), response.GetMessage())
		return
	}

	t.Logf("requestId: %s, text: %s\n", response.GetRequestId(), response.GetData().GetText())
}

func TestCreateToken(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")
	agentKey := os.Getenv("AGENT_KEY")
	endpoint := client.BroadscopeBailianPopEndpoint

	config := &openapi.Config{AccessKeyId: &accessKeyId,
		AccessKeySecret: &accessKeySecret,
		Endpoint:        &endpoint}

	tokenClient, err := apiClient.NewClient(config)
	if err != nil {
		fmt.Printf("failed to new client, err: %v\n", err)
		return
	}

	request := &apiClient.CreateTokenRequest{AgentKey: &agentKey}
	result, err := tokenClient.CreateToken(request)
	if err != nil {
		fmt.Printf("failed to create token, err: %v\n", err)
		return
	}

	resultBody := result.Body
	if !(*resultBody.Success) {
		requestId := resultBody.RequestId
		if requestId == nil {
			requestId = result.Headers["x-acs-request-id"]
		}

		errMessage := fmt.Sprintf("Failed to create token, reason: %s RequestId: %s", *resultBody.Message, *requestId)
		fmt.Printf("%v\n", errMessage)
	}

	fmt.Printf("token: %s, expiredTime : %d\n", *resultBody.Data.Token, *resultBody.Data.ExpiredTime)
}
