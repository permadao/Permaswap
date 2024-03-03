package router

const (
	testPort = ":9876"
)

func testGenRouter() *Router {
	config := &Config{
		ChainId: 5,
	}
	router := New(config, nil, nil, nil, false)
	router.Run(testPort, "")
	return router
}
