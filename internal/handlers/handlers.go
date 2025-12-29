package handlers

type URLService interface {
	AddURL(url string) (string, error)
	GetURL(id string) (string, error)
}

type Handlers struct {
	urlService URLService
}

func New(url URLService) *Handlers {
	return &Handlers{urlService: url}
}
