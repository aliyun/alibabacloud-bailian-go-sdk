/*
 * All rights Reserved, Designed By Alibaba Group Inc.
 * Copyright: Copyright(C) 1999-2023
 * Company  : Alibaba Group Inc.

 * @brief broadscope completion client
 * @author  yuanci.ytb
 * @version 1.0.0
 * @date 2023-08-04
 */

package broadscope_bailian

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alibabacloud-go/bailian-20230601/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	BroadscopeBailianPopEndpoint = "bailian.cn-beijing.aliyuncs.com"
	BroadscopeBailianEndpoint    = "https://bailian.aliyuncs.com"
	DocReferenceTypeSimple       = "simple"
	DocReferenceTypeIndexed      = "indexed"
)

var (
	SSEEventData  = []byte("data: ")
	SSEEventError = []byte(`data: {"error":`)
	SSEEventDone  = "[DONE]"
)

type AccessTokenClient struct {
	AccessKeyId     *string
	AccessKeySecret *string
	AgentKey        *string
	Endpoint        *string
	TokenData       *client.CreateTokenResponseBodyData
}

func ToString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (c AccessTokenClient) String() string {
	return tea.Prettify(c)
}

func (c AccessTokenClient) GoString() string {
	return c.String()
}

func (c *AccessTokenClient) SetAccessKeyId(v string) *AccessTokenClient {
	c.AccessKeyId = &v
	return c
}

func (c *AccessTokenClient) GetAccessKeyId() string {
	if c.AccessKeyId == nil {
		return ""
	}
	return *c.AccessKeyId
}

func (c *AccessTokenClient) SetAccessKeySecret(v string) *AccessTokenClient {
	c.AccessKeySecret = &v
	return c
}

func (c *AccessTokenClient) GetAccessKeySecret() string {
	if c.AccessKeySecret == nil {
		return ""
	}
	return *c.AccessKeySecret
}

func (c *AccessTokenClient) SetAgentKey(v string) *AccessTokenClient {
	c.AgentKey = &v
	return c
}

func (c *AccessTokenClient) GetAgentKey() string {
	if c.AgentKey == nil {
		return ""
	}
	return *c.AgentKey
}

func (c *AccessTokenClient) SetEndpoint(v string) *AccessTokenClient {
	c.Endpoint = &v
	return c
}

func (c *AccessTokenClient) GetEndpoint() string {
	if c.Endpoint == nil {
		return ""
	}
	return *c.Endpoint
}

func (c *AccessTokenClient) SetTokenData(tokenData *client.CreateTokenResponseBodyData) *AccessTokenClient {
	c.TokenData = tokenData
	return c
}

func (c *AccessTokenClient) GetTokenData(tokenData *client.CreateTokenResponseBodyData) *client.CreateTokenResponseBodyData {
	return c.TokenData
}

func (c *AccessTokenClient) GetToken() (_token string, _err error) {
	timestamp := time.Now().Unix()
	//Token有效时间24小时, 本地缓存token, 以免每次请求token
	if c.TokenData == nil || (*c.TokenData.ExpiredTime-600) < timestamp {
		result, err := c.CreateToken()
		if err != nil {
			return "", err
		}

		c.SetTokenData(result)
	}

	return *c.TokenData.Token, nil
}

func (c *AccessTokenClient) CreateToken() (_result *client.CreateTokenResponseBodyData, _err error) {
	if c.Endpoint == nil {
		c.SetEndpoint(BroadscopeBailianPopEndpoint)
	}

	config := &openapi.Config{AccessKeyId: c.AccessKeyId,
		AccessKeySecret: c.AccessKeySecret,
		Endpoint:        c.Endpoint}

	tokenClient, err := client.NewClient(config)
	if err != nil {
		return nil, err
	}

	request := &client.CreateTokenRequest{AgentKey: c.AgentKey}
	result, err := tokenClient.CreateToken(request)
	if err != nil {
		return nil, err
	}

	resultBody := result.Body
	if !(*resultBody.Success) {
		requestId := resultBody.RequestId
		if requestId == nil {
			requestId = result.Headers["x-acs-request-id"]
		}

		errMessage := fmt.Sprintf("Failed to create token, reason: %s RequestId: %s",
			ToString(resultBody.Message), ToString(requestId))
		return nil, errors.New(errMessage)
	}

	return resultBody.Data, nil
}

type CompletionClient struct {
	Token    *string
	Endpoint *string
	Timeout  time.Duration
}

func (cc CompletionClient) String() string {
	return tea.Prettify(cc)
}

func (cc CompletionClient) GoString() string {
	return cc.String()
}

func (cc *CompletionClient) SetToken(v string) *CompletionClient {
	cc.Token = &v
	return cc
}

func (cc *CompletionClient) GetToken() string {
	if cc.Token == nil {
		return ""
	}
	return *cc.Token
}

func (cc *CompletionClient) SetEndpoint(v string) *CompletionClient {
	cc.Endpoint = &v
	return cc
}

func (cc *CompletionClient) GetEndpoint() string {
	if cc.Endpoint == nil {
		return ""
	}
	return *cc.Endpoint
}

func (cc *CompletionClient) SetTimeout(v time.Duration) *CompletionClient {
	cc.Timeout = v
	return cc
}

func (cc *CompletionClient) GetTimeout() time.Duration {
	return cc.Timeout
}

type ChatQaMessage struct {
	User string `json:"User"`
	Bot  string `json:"Bot"`
}

func (c ChatQaMessage) String() string {
	return tea.Prettify(c)
}

func (c ChatQaMessage) GoString() string {
	return c.String()
}

type CompletionRequest struct {
	RequestId        *string                          `json:"RequestId"`
	AppId            *string                          `json:"AppId"`
	Prompt           *string                          `json:"Prompt"`
	SessionId        *string                          `json:"SessionId,omitempty"`
	TopP             float32                          `json:"TopP,omitempty"`
	Stream           bool                             `json:"Stream,omitempty"`
	HasThoughts      bool                             `json:"HasThoughts,omitempty"`
	BizParams        *map[string]interface{}          `json:"BizParams,omitempty"`
	DocReferenceType *string                          `json:"DocReferenceType,omitempty"`
	History          []*ChatQaMessage                 `json:"History,omitempty"`
	Parameters       *CompletionRequestModelParameter `json:"Parameters,omitempty"`
	DocTagIds        []int64                          `json:"DocTagIds,omitempty"`
}

func (cr CompletionRequest) String() string {
	return tea.Prettify(cr)
}

func (cr CompletionRequest) GoString() string {
	return cr.String()
}

func (cr *CompletionRequest) SetRequestId(v string) *CompletionRequest {
	cr.RequestId = &v
	return cr
}

func (cr *CompletionRequest) GetRequestId() string {
	if cr.RequestId == nil {
		return ""
	}
	return *cr.RequestId
}

func (cr *CompletionRequest) SetAppId(v string) *CompletionRequest {
	cr.AppId = &v
	return cr
}

func (cr *CompletionRequest) GetAppId() string {
	if cr.AppId == nil {
		return ""
	}
	return *cr.AppId
}

func (cr *CompletionRequest) SetPrompt(v string) *CompletionRequest {
	cr.Prompt = &v
	return cr
}

func (cr *CompletionRequest) GetPrompt() string {
	if cr.Prompt == nil {
		return ""
	}
	return *cr.Prompt
}

func (cr *CompletionRequest) SetSessionId(v string) *CompletionRequest {
	cr.SessionId = &v
	return cr
}

func (cr *CompletionRequest) GetSessionId() string {
	if cr.SessionId == nil {
		return ""
	}
	return *cr.SessionId
}

func (cr *CompletionRequest) SetTopP(v float32) *CompletionRequest {
	cr.TopP = v
	return cr
}

func (cr *CompletionRequest) GetTopP() float32 {
	return cr.TopP
}

func (cr *CompletionRequest) SetHasThoughts(v bool) *CompletionRequest {
	cr.HasThoughts = v
	return cr
}

func (cr *CompletionRequest) GetHasThoughts() bool {
	return cr.HasThoughts
}

func (cr *CompletionRequest) SetStream(v bool) *CompletionRequest {
	cr.Stream = v
	return cr
}

func (cr *CompletionRequest) GetStream() bool {
	return cr.Stream
}

func (cr *CompletionRequest) SetBizParams(v *map[string]interface{}) *CompletionRequest {
	cr.BizParams = v
	return cr
}

func (cr *CompletionRequest) GetBizParams() *map[string]interface{} {
	return cr.BizParams
}

func (cr *CompletionRequest) SetDocReferenceType(v string) *CompletionRequest {
	cr.DocReferenceType = &v
	return cr
}

func (cr *CompletionRequest) GetDocReferenceType() string {
	if cr.DocReferenceType == nil {
		return ""
	}
	return *cr.DocReferenceType
}

func (cr *CompletionRequest) SetHistory(v []*ChatQaMessage) *CompletionRequest {
	cr.History = v
	return cr
}

func (cr *CompletionRequest) GetHistory() []*ChatQaMessage {
	return cr.History
}

func (cr *CompletionRequest) SetParameters(v *CompletionRequestModelParameter) *CompletionRequest {
	cr.Parameters = v
	return cr
}

func (cr *CompletionRequest) GetParameters() *CompletionRequestModelParameter {
	return cr.Parameters
}

func (cr *CompletionRequest) SetDocTagIds(v []int64) *CompletionRequest {
	cr.DocTagIds = v
	return cr
}

func (cr *CompletionRequest) GetDocTagIds() []int64 {
	return cr.DocTagIds
}

type CompletionRequestModelParameter struct {
	TopK         int32 `json:"TopK,omitempty"`
	Seed         int32 `json:"Seed,omitempty"`
	UseRawPrompt bool  `json:"UseRawPrompt,omitempty"`
}

func (cp CompletionRequestModelParameter) String() string {
	return tea.Prettify(cp)
}

func (cp CompletionRequestModelParameter) GoString() string {
	return cp.String()
}

func (cp *CompletionRequestModelParameter) SetTopK(v int32) *CompletionRequestModelParameter {
	cp.TopK = v
	return cp
}

func (cp *CompletionRequestModelParameter) GetTopK() int32 {
	return cp.TopK
}

func (cp *CompletionRequestModelParameter) SetSeed(v int32) *CompletionRequestModelParameter {
	cp.Seed = v
	return cp
}

func (cp *CompletionRequestModelParameter) GetSeed() int32 {
	return cp.Seed
}

func (cp *CompletionRequestModelParameter) SetUseRawPrompt(v bool) *CompletionRequestModelParameter {
	cp.UseRawPrompt = v
	return cp
}

func (cp *CompletionRequestModelParameter) GetUseRawPrompt() bool {
	return cp.UseRawPrompt
}

type CompletionResponseDataThought struct {
	Thought           *string `json:"Thought,omitempty"`
	ActionType        *string `json:"ActionType,omitempty"`
	ActionName        *string `json:"ActionName,omitempty"`
	Action            *string `json:"Action,omitempty"`
	ActionInputStream *string `json:"ActionInputStream,omitempty"`
	ActionInput       *string `json:"ActionInput,omitempty"`
	Response          *string `json:"Response,omitempty"`
	Observation       *string `json:"Observation,omitempty"`
}

func (ct CompletionResponseDataThought) String() string {
	return tea.Prettify(ct)
}

func (ct CompletionResponseDataThought) GoString() string {
	return ct.String()
}

func (ct *CompletionResponseDataThought) SetThought(v string) *CompletionResponseDataThought {
	ct.Thought = &v
	return ct
}

func (ct *CompletionResponseDataThought) GetThought() string {
	if ct.Thought == nil {
		return ""
	}
	return *ct.Thought
}

func (ct *CompletionResponseDataThought) SetActionType(v string) *CompletionResponseDataThought {
	ct.ActionType = &v
	return ct
}

func (ct *CompletionResponseDataThought) GetActionType() string {
	if ct.ActionType == nil {
		return ""
	}
	return *ct.ActionType
}

func (cr *CompletionResponseDataThought) SetActionName(v string) *CompletionResponseDataThought {
	cr.ActionName = &v
	return cr
}

func (cr *CompletionResponseDataThought) GetActionName() string {
	if cr.ActionName == nil {
		return ""
	}
	return *cr.ActionName
}

func (cr *CompletionResponseDataThought) SetAction(v string) *CompletionResponseDataThought {
	cr.Action = &v
	return cr
}

func (cr *CompletionResponseDataThought) GetAction() string {
	if cr.Action == nil {
		return ""
	}
	return *cr.Action
}

func (cr *CompletionResponseDataThought) SetActionInputStream(v string) *CompletionResponseDataThought {
	cr.ActionInputStream = &v
	return cr
}

func (cr *CompletionResponseDataThought) GetActionInputStream() string {
	if cr.ActionInputStream == nil {
		return ""
	}
	return *cr.ActionInputStream
}

func (cr *CompletionResponseDataThought) SetActionInput(v string) *CompletionResponseDataThought {
	cr.ActionInput = &v
	return cr
}

func (cr *CompletionResponseDataThought) GetActionInput() string {
	if cr.ActionInput == nil {
		return ""
	}
	return *cr.ActionInput
}

func (cr *CompletionResponseDataThought) SetResponse(v string) *CompletionResponseDataThought {
	cr.Response = &v
	return cr
}

func (cr *CompletionResponseDataThought) GetResponse() string {
	if cr.Response == nil {
		return ""
	}
	return *cr.Response
}

func (cr *CompletionResponseDataThought) SetObservation(v string) *CompletionResponseDataThought {
	cr.Observation = &v
	return cr
}

func (cr *CompletionResponseDataThought) GetObservation() string {
	if cr.Observation == nil {
		return ""
	}
	return *cr.Observation
}

type CompletionResponseDataDocReference struct {
	IndexId *string `json:"IndexId,omitempty"`
	Title   *string `json:"Title,omitempty"`
	DocId   *string `json:"DocId,omitempty"`
	DocName *string `json:"DocName,omitempty"`
	DocUrl  *string `json:"DocUrl,omitempty"`
	Text    *string `json:"Text,omitempty"`
	BizId   *string `json:"BizId,omitempty"`
}

func (cr CompletionResponseDataDocReference) String() string {
	return tea.Prettify(cr)
}

func (cr CompletionResponseDataDocReference) GoString() string {
	return cr.String()
}

func (cr *CompletionResponseDataDocReference) SetIndexId(v string) *CompletionResponseDataDocReference {
	cr.IndexId = &v
	return cr
}

func (cr *CompletionResponseDataDocReference) GetIndexId() string {
	if cr.IndexId == nil {
		return ""
	}
	return *cr.IndexId
}

func (cr *CompletionResponseDataDocReference) SetTitle(v string) *CompletionResponseDataDocReference {
	cr.Title = &v
	return cr
}

func (cr *CompletionResponseDataDocReference) GetTitle() string {
	if cr.Title == nil {
		return ""
	}
	return *cr.Title
}

func (cr *CompletionResponseDataDocReference) SetDocId(v string) *CompletionResponseDataDocReference {
	cr.DocId = &v
	return cr
}

func (cr *CompletionResponseDataDocReference) GetDocId() string {
	if cr.DocId == nil {
		return ""
	}
	return *cr.DocId
}

func (cr *CompletionResponseDataDocReference) SetDocName(v string) *CompletionResponseDataDocReference {
	cr.DocName = &v
	return cr
}

func (cr *CompletionResponseDataDocReference) GetDocName() string {
	if cr.DocName == nil {
		return ""
	}
	return *cr.DocName
}

func (cr *CompletionResponseDataDocReference) SetDocUrl(v string) *CompletionResponseDataDocReference {
	cr.DocUrl = &v
	return cr
}

func (cr *CompletionResponseDataDocReference) GetDocUrl() string {
	if cr.DocUrl == nil {
		return ""
	}
	return *cr.DocUrl
}

func (cr *CompletionResponseDataDocReference) SetText(v string) *CompletionResponseDataDocReference {
	cr.Text = &v
	return cr
}

func (cr *CompletionResponseDataDocReference) GetText() string {
	if cr.Text == nil {
		return ""
	}
	return *cr.Text
}

func (cr *CompletionResponseDataDocReference) SetBizId(v string) *CompletionResponseDataDocReference {
	cr.BizId = &v
	return cr
}

func (cr *CompletionResponseDataDocReference) GetBizId() string {
	if cr.BizId == nil {
		return ""
	}
	return *cr.BizId
}

type CompletionResponseData struct {
	ResponseId    *string                               `json:"ResponseId"`
	SessionId     *string                               `json:"SessionId,omitempty"`
	Text          *string                               `json:"Text,omitempty"`
	Thoughts      []*CompletionResponseDataThought      `json:"Thoughts,omitempty"`
	DocReferences []*CompletionResponseDataDocReference `json:"DocReferences,omitempty"`
}

func (cd CompletionResponseData) String() string {
	return tea.Prettify(cd)
}

func (cd CompletionResponseData) GoString() string {
	return cd.String()
}

func (cd *CompletionResponseData) SetResponseId(v string) *CompletionResponseData {
	cd.ResponseId = &v
	return cd
}

func (cd *CompletionResponseData) GetResponseId() string {
	if cd.ResponseId == nil {
		return ""
	}
	return *cd.ResponseId
}

func (cd *CompletionResponseData) SetSessionId(v string) *CompletionResponseData {
	cd.SessionId = &v
	return cd
}

func (cd *CompletionResponseData) GetSessionId() string {
	if cd.SessionId == nil {
		return ""
	}
	return *cd.SessionId
}

func (cd *CompletionResponseData) SetText(v string) *CompletionResponseData {
	cd.Text = &v
	return cd
}

func (cd *CompletionResponseData) GetText() string {
	if cd.Text == nil {
		return ""
	}
	return *cd.Text
}

func (cd *CompletionResponseData) SetThoughts(v []*CompletionResponseDataThought) *CompletionResponseData {
	cd.Thoughts = v
	return cd
}

func (cd *CompletionResponseData) GetThoughts() []*CompletionResponseDataThought {
	return cd.Thoughts
}

func (cd *CompletionResponseData) SetDocReferences(v []*CompletionResponseDataDocReference) *CompletionResponseData {
	cd.DocReferences = v
	return cd
}

func (cd *CompletionResponseData) GetDocReferences() []*CompletionResponseDataDocReference {
	return cd.DocReferences
}

type CompletionResponse struct {
	Success   bool                    `json:"Success"`
	Code      *string                 `json:"Code,omitempty"`
	Message   *string                 `json:"Message,omitempty"`
	RequestId *string                 `json:"RequestId,omitempty"`
	Data      *CompletionResponseData `json:"Data,omitempty"`
}

func (cr CompletionResponse) String() string {
	return tea.Prettify(cr)
}

func (cr CompletionResponse) GoString() string {
	return cr.String()
}

func (cr *CompletionResponse) SetSuccess(v bool) *CompletionResponse {
	cr.Success = v
	return cr
}

func (cr *CompletionResponse) GetSuccess() bool {
	return cr.Success
}

func (cr *CompletionResponse) SetCode(v string) *CompletionResponse {
	cr.Code = &v
	return cr
}

func (cr *CompletionResponse) GetCode() string {
	if cr.Code == nil {
		return ""
	}
	return *cr.Code
}

func (cr *CompletionResponse) SetMessage(v string) *CompletionResponse {
	cr.Message = &v
	return cr
}

func (cr *CompletionResponse) GetMessage() string {
	if cr.Message == nil {
		return ""
	}
	return *cr.Message
}

func (cr *CompletionResponse) SetRequestId(v string) *CompletionResponse {
	cr.RequestId = &v
	return cr
}

func (cr *CompletionResponse) GetRequestId() string {
	if cr.RequestId == nil {
		return ""
	}
	return *cr.RequestId
}

func (cr *CompletionResponse) SetData(v *CompletionResponseData) *CompletionResponse {
	cr.Data = v
	return cr
}

func (cr *CompletionResponse) GetData() *CompletionResponseData {
	return cr.Data
}

func (cc *CompletionClient) CreateCompletion(request *CompletionRequest) (_response *CompletionResponse, _err error) {
	req, err := cc.CreateCompletionRequest(request, false)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Timeout: cc.Timeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		errMessage := fmt.Sprintf("Failed to complete request, code: %d, message: %s", resp.StatusCode, string(body))
		return nil, errors.New(errMessage)
	}

	response := &CompletionResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (cc *CompletionClient) CreateStreamCompletion(request *CompletionRequest) (_response chan *CompletionResponse, _err error) {
	req, err := cc.CreateCompletionRequest(request, true)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		errMessage := fmt.Sprintf("Failed to complete request, code: %d, message: %s", resp.StatusCode, string(body))
		return nil, errors.New(errMessage)
	}

	result, err := cc.ReadStream(resp)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (cc *CompletionClient) CreateCompletionRequest(request *CompletionRequest, stream bool) (*http.Request, error) {
	if cc.Endpoint == nil {
		endpoint := BroadscopeBailianEndpoint
		cc.Endpoint = &endpoint
	}

	if request.RequestId == nil {
		requestId := strings.ReplaceAll(uuid.New().String(), "-", "")
		request.RequestId = &requestId
	}

	if stream {
		request.Stream = stream
	}

	url := fmt.Sprintf("%s/v2/app/completions", ToString(cc.Endpoint))
	data, err := json.Marshal(*request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	authorization := fmt.Sprintf("Bearer %s", ToString(cc.Token))

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", authorization)

	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}

	return req, err
}

func (cc *CompletionClient) ReadStream(response *http.Response) (chan *CompletionResponse, error) {
	ch := make(chan *CompletionResponse)
	reader := bufio.NewReader(response.Body)

	go func() {
		defer response.Body.Close()
		defer close(ch)

		for {
			rawLine, err := reader.ReadBytes('\n')
			if err == io.EOF {
				return
			}

			if err != nil {
				log.Printf("failed to read line, err: %v\n", err)
				return
			}

			line := bytes.TrimSpace(rawLine)
			if line == nil || len(line) == 0 {
				continue
			}

			if bytes.HasPrefix(line, SSEEventError) || !bytes.HasPrefix(line, SSEEventData) {
				log.Printf("got invalid event, line: %s\n", line)
				continue
			}

			dataLine := bytes.TrimPrefix(line, SSEEventData)
			if string(dataLine) == SSEEventDone {
				return
			}

			response := &CompletionResponse{}
			err = json.Unmarshal(dataLine, response)
			if err != nil {
				log.Printf("failed to unmarshal line: %v\n", err)
				continue
			}

			ch <- response
		}
	}()

	return ch, nil
}
