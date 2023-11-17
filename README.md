###### 阿里云百炼Go SDK #####


##### 文本生成 #####

```
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

	prompt := "帮我生成一篇500字的文章，描述一下春秋战国的政治、军事和经济"

	request := &client.CompletionRequest{}
	request.SetAppId(appId)
	request.SetPrompt(prompt)

	response, err := cc.CreateCompletion(request)
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	if !response.Success {
		t.Errorf("%v\n", *response.Message)
		return
	}

	t.Logf("response : %s\n", *response)
}
```

##### 流式文本生成 #####

```
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
		t.Errorf("failed to complete, err: %v\n", err)
		return
	}

	for result := range response {
		if !result.Success {
			t.Errorf("get result with error, requestId: %s, code: %s, message: %s\n", *result.RequestId,
				*result.Code, *result.Message)
		} else {
			t.Logf("result: %s\n", *result)
		}
	}
}
```