package base

type Server struct {
	Port  int
	IP    string
	App   *App
	Start func()
}
