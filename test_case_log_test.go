package mockserver_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/mockserver"
)

func TestNewMemeryDB(t *testing.T) {
	_, err := mockserver.NewMemeryDB()
	require.NoError(t, err)
}

func TestAddTestCaseLog(t *testing.T) {
	testCase := mockserver.TestCaseLog{
		Host:           "domain.com",
		Method:         "POST",
		Path:           "/api/v1/interface",
		RequestHeader:  "",                                   // 实际输入请求头
		RequestQuery:   "",                                   // 实际输入请求查询
		RequestBody:    `{"name":"world"}`,                   // 实际输入请求体
		ResponseHeader: `["content-type: application/json"]`, // 实际响应头
		ResponseBody:   `{"grate":"hello world"}`,            // 实际响应体
		TestCaseID:     "grate",
		TestCaseInput:  `{"name":"world"}`,        // 当前测试用例期望的输入（作为服务时：可用于测试输入是否符合预期）
		TestCaseOutput: `{"grate":"hello world"}`, // 当前测试用例期望的输出（作为客户端时：可用于测试真实服务端返回是否符合预期）
	}
	db, err := mockserver.NewMemeryDB()
	require.NoError(t, err)
	err = mockserver.CreateTable(db)
	require.NoError(t, err)
	err = testCase.Add(db)
	require.NoError(t, err)
}

func TestLoadLastByTestCaseID(t *testing.T) {
	testCase := mockserver.TestCaseLog{
		Host:           "domain.com",
		Method:         "POST",
		Path:           "/api/v1/interface",
		RequestHeader:  "",                                   // 实际输入请求头
		RequestQuery:   "",                                   // 实际输入请求查询
		RequestBody:    `{"name":"world"}`,                   // 实际输入请求体
		ResponseHeader: `["content-type: application/json"]`, // 实际响应头
		ResponseBody:   `{"grate":"hello world"}`,            // 实际响应体
		TestCaseID:     "grate",
		TestCaseInput:  `{"name":"world"}`,        // 当前测试用例期望的输入（作为服务时：可用于测试输入是否符合预期）
		TestCaseOutput: `{"grate":"hello world"}`, // 当前测试用例期望的输出（作为客户端时：可用于测试真实服务端返回是否符合预期）
	}
	db, err := mockserver.NewMemeryDB()
	require.NoError(t, err)
	err = mockserver.CreateTable(db)
	require.NoError(t, err)
	err = testCase.Add(db)
	require.NoError(t, err)

	//相同testCaseID 第二次添加
	err = testCase.Add(db)
	require.NoError(t, err)

	selectedTestCase, err := mockserver.GetLastTestLogByTestCaseID(testCase.TestCaseID)
	require.NoError(t, err)
	require.Equal(t, 2, selectedTestCase.ID)

}
