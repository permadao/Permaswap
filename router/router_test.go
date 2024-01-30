package router

const (
	testPort = ":9876"
)

func testGenRouter() *Router {
	router := New("", "", 5, nil, "", "", "", nil, false)
	router.Run(testPort, "")
	return router
}
