package mockserver

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

//TestCaseLog 日志记录，方便测试方获取服务内部数据对比
type TestCaseLog struct {
	ID             int    `json:"id" db:"id"`
	Host           string `json:"host" db:"host"`
	Method         string `json:"method" db:"method"`
	Path           string `json:"path" db:"path"`
	RequestId      string `json:"requestId" db:"request_id"`
	RequestHeader  string `json:"requestHeader" db:"request_header"`   // 实际输入请求头
	RequestQuery   string `json:"requestQuery" db:"request_query"`     // 实际输入请求查询
	RequestBody    string `json:"requestBody" db:"request_body"`       // 实际输入请求体
	ResponseHeader string `json:"responseHeader" db:"response_header"` // 实际响应头
	ResponseBody   string `json:"responseBody" db:"response_body"`     // 实际响应体
	TestCaseID     string `json:"testCaseId" db:"test_case_id"`
	Description    string `json:"description" db:"description"`
	TestCaseInput  string `json:"testCaseInput" db:"test_case_input"`   // 当前测试用例期望的输入（作为服务时：可用于测试输入是否符合预期）
	TestCaseOutput string `json:"testCaseOutput" db:"test_case_output"` // 当前测试用例期望的输出（作为客户端时：可用于测试真实服务端返回是否符合预期）
	CreatedAt      string `json:"createdAt" db:"created_at"`            // 当前测试用例期望的输出（作为客户端时：可用于测试真实服务端返回是否符合预期）
}

func (log TestCaseLog) Table() string {
	return "test_case_log"
}

func (log TestCaseLog) Add(db *sql.DB) (err error) {
	sql := fmt.Sprintf("insert into `%s`(`host`,`method`,`path`,`request_id`,`request_header`,`request_query`,`request_body`,`response_header`,`response_body`,`test_case_id`,`description`,`test_case_input`,`test_case_output`)values('%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s')",
		log.Table(),
		log.Host,
		log.Method,
		log.Path,
		log.RequestId,
		log.RequestHeader,
		log.RequestQuery,
		log.RequestBody,
		log.ResponseHeader,
		log.ResponseBody,
		log.TestCaseID,
		log.Description,
		log.TestCaseInput,
		log.TestCaseOutput,
	)
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}
	return
}

func (log *TestCaseLog) LoadLastByTestCaseID(db *sql.DB) (err error) {
	if log.TestCaseID == "" {
		err = errors.Errorf("TestCaseLog.Load.TestCaseID required")
		return err
	}
	sql := fmt.Sprintf("select * from `%s` where `test_case_id`='%s' order by `id` desc limit 1 ", log.Table(), log.TestCaseID)
	logs := make([]TestCaseLog, 0)
	rows, err := db.Query(sql)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = sqlx.StructScan(rows, &logs)
	if err != nil {
		return err
	}
	if len(logs) > 0 {
		*log = logs[0]
	}
	return nil
}

func GetLastTestLogByTestCaseID(testCaseID string) (testCaseLog *TestCaseLog, err error) {
	testCaseLog = &TestCaseLog{
		TestCaseID: testCaseID,
	}
	err = testCaseLog.LoadLastByTestCaseID(_db)
	if err != nil {
		return nil, err
	}
	return testCaseLog, nil
}

var _db *sql.DB

//go:embed ddl_mysql.sql
var ddl_mysql string

//go:embed ddl_sqlite3.sql
var ddl_sqlite3 string
var dbInit sync.Once

func CreateTable(db *sql.DB) (err error) {
	dbInit.Do(func() {
		_db = db
	})
	var ddl string
	switch strings.ToLower(DBDrivername) {
	case "mysql":
		ddl = ddl_mysql

	default:
		ddl = ddl_sqlite3
	}
	_, err = db.Exec(ddl)
	if err != nil {
		return err
	}
	return nil
}

var DBDrivername = "sqlite3"

//NewMemeryDB 创建内存数据库
func NewMemeryDB() (db *sql.DB, err error) {
	dsn := `:memory:`
	db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func UseMemoryDB() (err error) {
	db, err := NewMemeryDB()
	if err != nil {
		return err
	}
	err = CreateTable(db)
	if err != nil {
		return err
	}
	return nil
}
