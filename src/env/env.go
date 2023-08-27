package env

const (
	Prod string = "https://api.sgroup.qq.com"
	Test string = "https://sandbox.api.sgroup.qq.com"
)

var Env = Test

func Url(endpoint string) string {
	return Env + endpoint
}
