package module

import (
	"code.gitea.io/gitea/modules/log"
	"fmt"
	"sync"
)

type ReqType int

const (
	POST ReqType = iota + 1
	GET
)

var (
	redirectMap sync.Map = sync.Map{}
)

func UrlRegister(id int64, url string, request_type ReqType) error {
	key := fmt.Sprintf("%d%s", request_type, url)
	_, loaded := redirectMap.LoadOrStore(key, id)
	log.Info("register url: %s with type %d", url, request_type)
	if loaded {
		return fmt.Errorf("%s has been registered with type %d", url, id)
	}
	return nil
}

func UrlRedirectRequest(form map[string]string, data []byte, url string, request ReqType) (
	content_type string, payload []byte, redirected bool) {
	key := fmt.Sprintf("%d%s", request, url)
	spec_data, redirected := redirectMap.Load(key)
	if !redirected {
		return "", []byte{}, false
	}
	id, _ := spec_data.(int64)
	content_type, payload, redirected = Redirect(form, data, id)
	return
}
