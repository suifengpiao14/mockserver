package mockserver

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/kvstruct"
	"github.com/suifengpiao14/logchan/v2"
)

type Service struct {
	Description string `json:"description"`
	Host        string `json:"host"`
	Apis        Apis   `json:"apis"`
}

func (s *Service) Init() (err error) {
	u, err := url.Parse(s.Host)
	if err != nil {
		err = errors.WithMessagef(err, "Service.Init Host:%s", s.Host)
		return err
	}
	s.Host = u.Host
	s.Apis.Init(s)
	return nil
}

type Services []Service

func (ss Services) Init() (err error) {
	for i := range ss {
		err = ss[i].Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (services *Services) AddReplace(subServices ...Service) {
	for _, imcomeService := range subServices {
		err := imcomeService.Init()
		if err != nil {
			panic(err)
		}
		exists := false
		for i, service := range *services {
			if strings.EqualFold(imcomeService.Host, service.Host) {
				exists = true
				(*services)[i] = service
				break
			}
		}
		if !exists {
			*services = append(*services, imcomeService)
		}
	}
}

func (ss Services) GetService(host string) (service *Service, err error) {
	for _, s := range ss {
		if strings.EqualFold(s.Host, host) {
			return &s, nil
		}
	}
	err = errors.Errorf("not found service by host:%s", host)
	return nil, err
}

func (ss Services) GetTestCase(testCaseID string) (testCase *TestCase, err error) {
	for _, s := range ss {
		for _, api := range s.Apis {
			for _, testCase := range api.TestCases {
				if strings.EqualFold(testCase.ID, testCaseID) {
					return &testCase, nil
				}
			}
		}
	}
	err = errors.Errorf("not found testCase  by id:%s", testCaseID)
	return nil, err
}

type Api struct {
	Path                   string    `json:"path"`
	Method                 string    `json:"method"`
	InputFeature           Feature   `json:"requestFeature"` // api请求的特征量，比如_interface,特定header等
	TestCases              TestCases `json:"testCases"`      //测试用例集合
	Request2inputFeatureFn func(r *http.Request) (inputFeature Feature, err error)
	service                *Service
}

var (
	ERROR_API_Request2inputFeatureFn_IS_NULL = errors.New("api.Request2inputFeatureFn is null")
)

func (api Api) GetService() (service *Service) {
	return api.service
}

// http请求转换为输入特征
func (api Api) Request2InputFeature(r *http.Request) (inputFeatures Feature, err error) {
	inputFeatures = api.InputFeature
	if r == nil {
		return inputFeatures, nil
	}

	if api.Request2inputFeatureFn == nil {
		err = errors.WithMessagef(ERROR_API_Request2inputFeatureFn_IS_NULL, "api:%s", api.Route())
		return nil, err
	}
	req, err := CopyRequest(r) // 复制一份，避免body被读取后不可再读等问题
	if err != nil {
		err = errors.WithMessagef(err, "CopyRequest,route:%s", api.Route())
		return nil, err
	}
	inputFeatures, err = api.Request2inputFeatureFn(req)
	if err != nil {
		err = errors.WithMessagef(err, "api.Request2inputFeatureFn,route:%s", api.Route())
		return nil, err
	}
	return inputFeatures, nil
}

//Handle 提取请求特征，匹配测试用例，生成返回数据
func (api Api) Handle(w http.ResponseWriter, r *http.Request) (err error) {
	cpyR, _ := CopyRequest(r)
	logInfo := MockServiceLog{
		Request: cpyR,
		Api:     api,
	}
	var testCaseRef *TestCase
	var out []byte
	defer func() {
		logInfo.err = err
		if testCaseRef != nil {
			logInfo.TestCase = *testCaseRef
		}
		logInfo.ResponseBody = out
		logchan.SendLogInfo(&logInfo)
	}()
	inputFeatures, err := api.Request2InputFeature(r)
	if err != nil {
		return err
	}
	testCaseRef, err = api.TestCases.GetByInputFeature(inputFeatures)
	if err != nil {
		return err
	}

	out, err = testCaseRef.GetOutput()
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out)
	return nil
}

func (api Api) Route() (identify string) {
	return route(api.Path, api.Method)
}

func route(path string, method string) (identify string) {
	method = strings.ToUpper(method)
	identify = fmt.Sprintf("%s-%s", path, method)
	return identify
}

type Apis []Api

func (apis Apis) Init(s *Service) {
	for i := range apis {
		apis[i].service = s
		apis[i].TestCases.Init(&apis[i])
	}
}

func (apis Apis) GetApi(r *http.Request) (api *Api, err error) {
	path, method := r.URL.Path, r.Method
	routeFeature := route(path, method)
	for _, api := range apis {
		if strings.EqualFold(api.Route(), routeFeature) {
			allFeature, err := api.Request2InputFeature(r)
			if err != nil {
				return nil, err
			}
			if allFeature.Contains(api.InputFeature) { // 判断api特征量
				return &api, nil
			}
		}
	}
	err = errors.Errorf("not found api by path:%s,method:%s", path, method)
	return nil, err
}

func (apis *Apis) AddReplace(subApis ...Api) {
	for _, incomeApi := range subApis {
		exists := false
		for i, api := range *apis {
			if strings.EqualFold(incomeApi.Route(), api.Route()) {
				exists = true
				(*apis)[i] = api
				break
			}
		}
		if !exists {
			*apis = append(*apis, incomeApi)
		}
	}
}

func Json2Feature(josnStr string) (feature Feature) {
	feature = make(Feature)
	kvs := kvstruct.JsonToKVS(josnStr, "")
	for _, kv := range kvs {
		feature[kv.Key] = kv.Value
	}
	return feature
}

func RequestBody2Feature(r *http.Request) (feature Feature, err error) {
	if r.Body == nil {
		return nil, nil
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(b)) //重新填写
	feature = Json2Feature(string(b))
	return feature, nil
}

//Feature 输入输出特征参数
type Feature map[string]string

func (feature Feature) Contains(subFeature Feature) (yes bool) {
	for k, v := range subFeature {
		if !strings.EqualFold(feature[k], v) { // 指定的key值不相同，返回false
			return false
		}
	}
	return true
}

func (feature Feature) String() (s string) {
	arr := make([]string, 0)
	for k, v := range feature {
		kv := fmt.Sprintf("%s:%s", k, v)
		arr = append(arr, kv)
	}
	sort.Strings(arr) //排序确保每次输出一致
	s = strings.Join(arr, ",")
	return s
}

func (feature *Feature) Merge(subFeature Feature) {
	for k, v := range subFeature {
		(*feature)[k] = v
	}
}

type TestCase struct {
	ID            string                                   `json:"id"`            // 测试用例唯一标识
	Description   string                                   `json:"description"`   //描述
	InputFeature  Feature                                  `json:"inputFeature"`  //输入特征
	OutputFeature Feature                                  `json:"outputFeature"` //输出特征
	InputFn       func(t TestCase) (out []byte, err error) // 生成输入内容函数
	OutputFn      func(t TestCase) (out []byte, err error) // 生成输出内容函数
	api           *Api
	//subscribers   []func(request *http.Request, response *http.Response)//   暂时通过日志实现该功能
}

func (tc TestCase) GetApi() (api *Api) {
	return api
}

// //Subscribe 订阅输入输出数据 暂时通过日志实现该功能
// func (tc *TestCase) Subscribe(subscriber func(request *http.Request, response *http.Response)) {
// 	tc.subscribers = append(tc.subscribers, subscriber)
// }

// //Publish 将请求响应发布  暂时通过日志实现该功能
// func (tc TestCase) Publish(request *http.Request, response *http.Response) {
// 	for _, subscriber := range tc.subscribers {
// 		r, _ := CopyRequest(request)
// 		res, _ := CopyResponse(response)
// 		subscriber(r, res)
// 	}
// }

func (tc TestCase) GetInput() (out []byte, err error) {
	if tc.InputFn == nil {
		return nil, nil
	}
	out, err = tc.InputFn(tc)
	if err != nil {
		err = errors.WithMessagef(err, "TestCase.InputFn,testCaseId:%s", tc.ID)
		return nil, err
	}
	return out, nil
}

func (tc TestCase) GetOutput() (out []byte, err error) {
	if tc.OutputFn == nil {
		return nil, nil
	}
	out, err = tc.OutputFn(tc)
	if err != nil {
		err = errors.WithMessagef(err, "TestCase.OutputFn,testCaseId:%s", tc.ID)
		return nil, err
	}
	return out, nil
}

type TestCases []TestCase

func (testCases TestCases) Init(api *Api) {
	for i := range testCases {
		testCases[i].api = api
	}
}

func (tcs TestCases) GetByInputFeature(requestFeature Feature) (testCase *TestCase, err error) {
	for _, tc := range tcs {
		if requestFeature.Contains(tc.InputFeature) {
			return &tc, nil
		}
	}
	err = errors.Errorf("not found test case by inputFeature:%s", requestFeature.String())
	return nil, err
}

func (tcs *TestCases) AddReplace(subApis ...TestCase) {
	for _, incomeTestCase := range subApis {
		exists := false
		for i, t := range *tcs {
			if strings.EqualFold(incomeTestCase.InputFeature.String(), t.InputFeature.String()) {
				exists = true
				(*tcs)[i] = t
				break
			}
		}
		if !exists {
			*tcs = append(*tcs, incomeTestCase)
		}
	}
}

const (
	HTTP_MOCK_SERVER = "MOCKSERVER"
)

var _UseMockServer bool

// SetUseMockServer 启用模拟服务
func SetUseMockServer() {
	_UseMockServer = true
}

func CanUseMockServer(r *http.Request) bool {
	return _UseMockServer
}

// CopyRequest 复制HTTP请求
func CopyRequest(r *http.Request) (*http.Request, error) {
	// 创建一个新的请求
	req := &http.Request{}

	// 复制请求方法、URL和协议版本
	req.Method = r.Method
	req.URL = r.URL
	req.Proto = r.Proto
	req.ProtoMajor = r.ProtoMajor
	req.ProtoMinor = r.ProtoMinor

	// 复制请求头
	req.Header = make(http.Header)
	for k, vv := range r.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	if r.Body == nil {
		return req, nil
	}

	// 复制请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body)) //重新填写
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	return req, nil
}

// CopyResponse 复制http.Response对象
func CopyResponse(resp *http.Response) (newResp *http.Response, err error) {
	// 创建一个新的http.Response对象
	// 注意：这里只复制了部分字段，实际情况中可能需要根据需要复制更多字段
	// 复制响应体
	var bodyBytes []byte
	if resp.Body != nil {
		// 读取响应体的内容
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // 重新赋值
	}

	response := &http.Response{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		Proto:         resp.Proto,
		ProtoMajor:    resp.ProtoMajor,
		ProtoMinor:    resp.ProtoMinor,
		Header:        make(http.Header),
		Body:          io.NopCloser(bytes.NewReader(bodyBytes)), // 创建一个空的响应体
		ContentLength: resp.ContentLength,
	}

	// 复制响应头
	for k, v := range resp.Header {
		response.Header[k] = v
	}

	return response, nil
}

func HttpMockServer(req *http.Request) (resp *http.Response, err error) {
	service, err := GetServer().GetService(req.Host)
	if err != nil {
		return nil, err
	}
	api, err := service.Apis.GetApi(req)
	if err != nil {
		return nil, err
	}
	w := httptest.NewRecorder()
	err = api.Handle(w, req)
	if err != nil {
		return nil, err
	}
	resp = w.Result()
	return resp, nil
}
