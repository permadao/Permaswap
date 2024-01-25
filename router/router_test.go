package router

const (
	testPort = ":9876"
)

func testGenRouter() *Router {
	router := New(5, nil, "", "", true, "")
	router.Run(testPort, "")
	return router
}
