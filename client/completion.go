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

func ToString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

type AccessTokenClient struct {
	AccessKeyId     string
	AccessKeySecret string
	AgentKey        string
	Endpoint        string
	TokenData       *client.CreateTokenResponseBodyData
}

func (c AccessTokenClient) String() string {
	return tea.Prettify(c)
}

func (c AccessTokenClient) GoString() string {
	return c.String()
}

func (c *AccessTokenClient) GetToken() (_token string, _err error) {
	timestamp := time.Now().Unix()
	//Token有效时间24小时, 本地缓存token, 以免每次请求token
	if c.TokenData == nil || (*c.TokenData.ExpiredTime-600) < timestamp {
		result, err := c.CreateToken()
		if err != nil {
			return "", err
		}

		c.TokenData = result
	}

	return *c.TokenData.Token, nil
}

func (c *AccessTokenClient) CreateToken() (_result *client.CreateTokenResponseBodyData, _err error) {
	if c.Endpoint == "" {
		c.Endpoint = BroadscopeBailianPopEndpoint
	}

	config := &openapi.Config{AccessKeyId: &c.AccessKeyId,
		AccessKeySecret: &c.AccessKeySecret,
		Endpoint:        &c.Endpoint}

	tokenClient, err := client.NewClient(config)
	if err != nil {
		return nil, err
	}

	request := &client.CreateTokenRequest{AgentKey: &c.AgentKey}
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

type ChatCompletionMessage struct {
	Role    string `json:"Role"`
	Content string `json:"Content"`
}

func (c ChatCompletionMessage) String() string {
	return tea.Prettify(c)
}

func (c ChatCompletionMessage) GoString() string {
	return c.String()
}

type CompletionRequest struct {
	RequestId        string                           `json:"RequestId"`
	AppId            string                           `json:"AppId"`
	Prompt           string                           `json:"Prompt"`
	SessionId        string                           `json:"SessionId,omitempty"`
	TopP             float32                          `json:"TopP,omitempty"`
	Stream           bool                             `json:"Stream,omitempty"`
	HasThoughts      bool                             `json:"HasThoughts,omitempty"`
	BizParams        map[string]interface{}           `json:"BizParams,omitempty"`
	DocReferenceType string                           `json:"DocReferenceType,omitempty"`
	History          []ChatQaMessage                  `json:"History,omitempty"`
	Messages         []ChatCompletionMessage          `json:"Messages,omitempty"`
	Parameters       *CompletionRequestModelParameter `json:"Parameters,omitempty"`
	DocTagIds        []int64                          `json:"DocTagIds,omitempty"`
	DocTagCodes      []string                         `json:"DocTagCodes,omitempty"`
}

func (cr CompletionRequest) String() string {
	return tea.Prettify(cr)
}

func (cr CompletionRequest) GoString() string {
	return cr.String()
}

type CompletionRequestModelParameter struct {
	TopK              int32    `json:"TopK,omitempty"`
	Seed              int32    `json:"Seed,omitempty"`
	UseRawPrompt      bool     `json:"UseRawPrompt,omitempty"`
	Temperature       float32  `json:"Temperature,omitempty"`
	MaxTokens         int32    `json:"MaxTokens,omitempty"`
	ResultFormat      string   `json:"ResultFormat,omitempty"`
	Stop              []string `json:"Stop,omitempty"`
	IncrementalOutput bool     `json:"IncrementalOutput,omitempty"`
}

func (cp CompletionRequestModelParameter) String() string {
	return tea.Prettify(cp)
}

func (cp CompletionRequestModelParameter) GoString() string {
	return cp.String()
}

type CompletionResponseDataThought struct {
	Thought           string `json:"Thought,omitempty"`
	ActionType        string `json:"ActionType,omitempty"`
	ActionName        string `json:"ActionName,omitempty"`
	Action            string `json:"Action,omitempty"`
	ActionInputStream string `json:"ActionInputStream,omitempty"`
	ActionInput       string `json:"ActionInput,omitempty"`
	Response          string `json:"Response,omitempty"`
	Observation       string `json:"Observation,omitempty"`
}

func (ct CompletionResponseDataThought) String() string {
	return tea.Prettify(ct)
}

func (ct CompletionResponseDataThought) GoString() string {
	return ct.String()
}

type CompletionResponseDataDocReference struct {
	IndexId string `json:"IndexId,omitempty"`
	Title   string `json:"Title,omitempty"`
	DocId   string `json:"DocId,omitempty"`
	DocName string `json:"DocName,omitempty"`
	DocUrl  string `json:"DocUrl,omitempty"`
	Text    string `json:"Text,omitempty"`
	BizId   string `json:"BizId,omitempty"`
}

func (cr CompletionResponseDataDocReference) String() string {
	return tea.Prettify(cr)
}

func (cr CompletionResponseDataDocReference) GoString() string {
	return cr.String()
}

type CompletionResponseDataUsage struct {
	InputTokens  int32  `json:"InputTokens"`
	OutputTokens int32  `json:"OutputTokens"`
	ModelId      string `json:"ModelId,omitempty"`
}

func (cu CompletionResponseDataUsage) String() string {
	return tea.Prettify(cu)
}

func (cu CompletionResponseDataUsage) GoString() string {
	return cu.String()
}

type CompletionResponseMessage struct {
	Role    string `json:"Role"`
	Content string `json:"Content"`
}

func (cm CompletionResponseMessage) String() string {
	return tea.Prettify(cm)
}

func (cm CompletionResponseMessage) GoString() string {
	return cm.String()
}

type CompletionResponseChoice struct {
	FinishReason string                     `json:"FinishReason,omitempty"`
	Message      *CompletionResponseMessage `json:"Message"`
}

func (rc CompletionResponseChoice) String() string {
	return tea.Prettify(rc)
}

func (rc CompletionResponseChoice) GoString() string {
	return rc.String()
}

type CompletionResponseData struct {
	ResponseId    string                               `json:"ResponseId"`
	SessionId     string                               `json:"SessionId,omitempty"`
	Text          string                               `json:"Text,omitempty"`
	Thoughts      []CompletionResponseDataThought      `json:"Thoughts,omitempty"`
	DocReferences []CompletionResponseDataDocReference `json:"DocReferences,omitempty"`
	Usage         []CompletionResponseDataUsage        `json:"Usage,omitempty"`
	Choices       []CompletionResponseChoice           `json:"Choices,omitempty"`
}

func (cd CompletionResponseData) String() string {
	return tea.Prettify(cd)
}

func (cd CompletionResponseData) GoString() string {
	return cd.String()
}

type CompletionResponse struct {
	Success   bool                    `json:"Success"`
	Code      string                  `json:"Code,omitempty"`
	Message   string                  `json:"Message,omitempty"`
	RequestId string                  `json:"RequestId,omitempty"`
	Data      *CompletionResponseData `json:"Data,omitempty"`
}

func (cr CompletionResponse) String() string {
	return tea.Prettify(cr)
}

func (cr CompletionResponse) GoString() string {
	return cr.String()
}

type CompletionClient struct {
	Token    string
	Endpoint string
	Timeout  time.Duration
}

func (cc CompletionClient) String() string {
	return tea.Prettify(cc)
}

func (cc CompletionClient) GoString() string {
	return cc.String()
}

func (cc *CompletionClient) CreateCompletionRequest(request *CompletionRequest, stream bool) (*http.Request, error) {
	if cc.Endpoint == "" {
		cc.Endpoint = BroadscopeBailianEndpoint
	}

	if request.RequestId == "" {
		requestId := strings.ReplaceAll(uuid.New().String(), "-", "")
		request.RequestId = requestId
	}

	if stream {
		request.Stream = stream
	}

	url := fmt.Sprintf("%s/v2/app/completions", cc.Endpoint)
	data, err := json.Marshal(*request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	authorization := fmt.Sprintf("Bearer %s", cc.Token)

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", authorization)

	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}

	return req, err
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
