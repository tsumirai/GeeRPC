package codec

import "io"

type Header struct{
	ServiceMethod string // format "Service.Method"
	Seq uint64 //请求的序号，用来区分不同的请求
	Error string
}