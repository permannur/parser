package entity

type ResponseStruct struct {
	Error error
	Items []ResponseItem
}

type ResponseItem struct {
	Host string
	Url  string
}
