package mockserver_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/mockserver"
)

//InitMockServer 初始化mock服务
func InitMockServer() (err error) {
	mockserver.SetUseMockServer(func(r *http.Request) bool {
		return true
	}) // 启用模拟服务器
	var mockSerices = MockSerices()
	mockserver.SetServer(mockSerices)
	err = mockserver.UseMemoryDB()
	if err != nil {
		return err
	}
	return nil
}

const (
	GetOrderByID21167382 = "GetOrderByID21167382"
)

func TestDealPlaceOrderEvent(t *testing.T) {
	err := InitMockServer()
	require.NoError(t, err)
	t.Run("getOrderById", func(t *testing.T) {
		orderID := "21167382"
		order, err := GetOrderByID(orderID)
		require.NoError(t, err)
		testCase, err := mockserver.GetLastTestLogByTestCaseID(GetOrderByID21167382)
		require.NoError(t, err)
		exceptedB := []byte(testCase.TestCaseOutput)
		responseB := []byte(order.String())
		// 期望发送的短信内容
		diffPathB, _ := mockserver.DiffJson(responseB, exceptedB)
		diffPath := string(diffPathB)
		require.Equal(t, diffPath, "{}")
	})
}

type Order struct {
	OrderID string `json:"orderId"`
	Price   string `json:"price"`
}

func (o Order) String() (s string) {
	b, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	s = string(b)
	return s
}

func GetOrderByID(orderID string) (order *Order, err error) {
	var body = strings.NewReader(fmt.Sprintf(`{"head":{"interface":"orderDetail"},"params":{"orderId":"%s"}}`, orderID))
	r, err := http.NewRequest(http.MethodPost, "https://order.domain.com?hello=world", body)
	if err != nil {
		return nil, err
	}
	r.Header.Set(mockserver.Header_Request_id, "15451aaierew")
	var resp *http.Response
	if mockserver.CanUseMockServer(r) {
		resp, err = mockserver.HttpMockServer(r)
	} else {
		client := &http.Client{}
		resp, err = client.Do(r)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	order = &Order{}
	err = json.Unmarshal(data, order)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func MockSerices() (mockServices mockserver.Services) {
	mockServices = make(mockserver.Services, 0)
	orderService := mockserver.Service{
		Description: "订单服务",
		Host:        "https://order.domain.com",
		Apis: mockserver.Apis{
			mockserver.Api{
				Path:                   "/",
				Method:                 http.MethodPost,
				Request2inputFeatureFn: mockserver.Request2Feature,
				InputFeature: mockserver.Feature{
					"body.head.interface": "orderDetail",
				},
				TestCases: mockserver.TestCases{
					mockserver.TestCase{
						ID:          GetOrderByID21167382,
						Description: "模拟订单详情接口",
						InputFeature: mockserver.Feature{
							"body.params.orderId": "21167382",
						},
						InputFn: func(t mockserver.TestCase) (header http.Header, out []byte, err error) {
							header = make(http.Header)
							header.Add("content-type", "application/json")
							out = []byte(`{"head":{"interface":"orderDetail"},"params":{"orderId":"21167382"}}`)
							return
						},
						OutputFn: func(t mockserver.TestCase) (header http.Header, out []byte, err error) {
							header = make(http.Header)
							header.Add("content-type", "application/json")
							out = []byte(`{"orderId":"21167382","price":"124"}`)
							return
						},
					},
				},
			},
		},
	}
	mockServices.AddReplace(orderService)
	return mockServices
}
