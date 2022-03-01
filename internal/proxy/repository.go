package proxy

import ()

type Repository interface {
	//isSecure: http -> false, https -> true
	Save(request string, response string, reqJson string, resJson string, isSecure bool)
}
