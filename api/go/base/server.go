package base

type Server struct {
	Port  int
	IP    string
	Start func()
}
