package mockserver

var _services = make(Services, 0)

// SetServer 设置服务器集合实例
func SetServer(services Services) {
	_services = services
}

func GetServer() (services Services) {
	return _services
}
