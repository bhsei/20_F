package module

import (
	"fmt"
	"sync"
)

type ReqType int

type UrlSpec struct {
	id      int64
	request ReqType
}

const (
	POST ReqType = iota + 1
	GET
)

var (
	redirectMap sync.Map = sync.Map{}
)

func UrlRegister(id int64, url string, request_type ReqType) error {
	_, loaded := redirectMap.LoadOrStore(url, &UrlSpec{
		id:      id,
		request: request_type,
	})
	if loaded {
		return fmt.Errorf("%s has been registered with id %d", url, id)
	}
	return nil
}

func UrlRedirectRequest(form map[string]string, data []byte, url_pattern string, request ReqType) (
	content_type string, payload []byte, redirected bool) {
	spec_data, redirected := redirectMap.Load(url_pattern)
	if !redirected {
		return
	}
	spec, _ := spec_data.(UrlSpec)
	if spec.request != request {
		redirected = false
		return
	}
	content_type, payload, redirected = Redirect(form, data, spec.id)
	return
}
