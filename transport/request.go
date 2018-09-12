package transport

type Request struct {
	cmd string
	data map[string]interface{}
}
