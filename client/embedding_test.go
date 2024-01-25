/*
 * All rights Reserved, Designed By Alibaba Group Inc.
 * Copyright: Copyright(C) 1999-2023
 * Company  : Alibaba Group Inc.

 * @brief embedding test
 * @author  yuanci.ytb
 * @version 1.0.0
 * @date 2024-01-23
 */

package broadscope_bailian_test

import (
	"fmt"
	apiClient "github.com/alibabacloud-go/bailian-20230601/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	client "github.com/aliyun/alibabacloud-bailian-go-sdk/client"
	"os"
	"testing"
)

func TestCreateTextEmbeddings(t *testing.T) {
	accessKeyId := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")
	agentKey := os.Getenv("AGENT_KEY")
	endpoint := client.BroadscopeBailianPopEndpoint

	config := &openapi.Config{AccessKeyId: &accessKeyId,
		AccessKeySecret: &accessKeySecret,
		Endpoint:        &endpoint}

	openapiClient, err := apiClient.NewClient(config)
	if err != nil {
		fmt.Printf("failed to new client, err: %v\n", err)
		return
	}

	request := &apiClient.CreateTextEmbeddingsRequest{AgentKey: &agentKey}
	text := "今天天气怎么样"
	request.SetInput([]*string{&text})
	result, err := openapiClient.CreateTextEmbeddings(request)
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

		errMessage := fmt.Sprintf("Failed to create embeddings, reason: %s RequestId: %s", *resultBody.Message, *requestId)
		fmt.Printf("%v\n", errMessage)
		return
	}

	embeddings := resultBody.Data.Embeddings
	if embeddings != nil {
		for _, embedding := range embeddings {
			fmt.Printf("index: %d, embeding : [", embedding.TextIndex)
			for _, num := range embedding.Embedding {
				fmt.Printf("%.16f, ", *num)
			}
			fmt.Printf("]")
		}
	}
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
