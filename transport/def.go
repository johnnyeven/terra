package transport

type driver interface {
	MakeRequest() *Request
	Request(request *Request)
}
