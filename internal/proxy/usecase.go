package proxy

import ()

type Usecase interface {
	Save(request string, response string, isSecure bool)
}
