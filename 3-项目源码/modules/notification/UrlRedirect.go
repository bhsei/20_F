package notification

type ReqType int

const (
	POST ReqType = iota + 1
	GET
)

func UrlRegister(id int64, url string, request_type ReqType) error {
	return nil
}
