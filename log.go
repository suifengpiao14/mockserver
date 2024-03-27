package mockserver

import (
	"net/http"

	"github.com/suifengpiao14/logchan/v2"
)

type LogName string

func (l LogName) String() string {
	return string(l)
}

const (
	LOG_INFO_MOCK_SERVICE LogName = "LogInfoMockService"
)

type MockServiceLog struct {
	Request      *http.Request
	RequestBody  []byte
	ResponseBody []byte
	Service      Service
	Api          Api
	TestCase     TestCase
	logchan.EmptyLogInfo
	err error
}

func (l MockServiceLog) GetName() (logName logchan.LogName) {
	return LOG_INFO_MOCK_SERVICE
}

func (l MockServiceLog) Error() (err error) {
	return l.err
}
func (l *MockServiceLog) BeforeSend() {
	if l.Api.Route() == "" {
		api := l.TestCase.GetApi()
		if api != nil {
			l.Api = *api
		}
	}
	if l.Api.Route() == "" {
		return
	}
	if l.Service.Host == "" && l.Api.Route() != "" {
		service := l.Api.GetService()
		if service != nil {
			l.Service = *service
		}
	}

}
