package main

const (
	Prod string = "https://api.sgroup.qq.com"
	Test string = "https://sandbox.api.sgroup.qq.com"
)

var env = Test

func Url(endpoint string) string {
	return env + endpoint
}
