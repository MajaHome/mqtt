package server

type Backend struct {
}

func GetBackend() *Backend {
	return new(Backend)
}

func (b Backend) Close() {
}