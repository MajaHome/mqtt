package server

type Backend struct {
}

var DEBUG bool = false

func GetBackend(debug bool) *Backend {
	DEBUG = debug
	return &Backend{}
}

func (b *Backend) Close() {
}

func (b *Backend) Authorize(login string, pass string) bool {

	return true
}

func (b *Backend) Subscribe() bool {

	return true
}

func (b *Backend) UnSubscribe() bool {

	return true
}

func (b *Backend) Publish() bool {

	return true
}
