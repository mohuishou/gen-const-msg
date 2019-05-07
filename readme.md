# go-const-msg

提取源码文件中的常量注释，生成对应的 msg 信息，常用于`err code`当中

## install

go install github.com/mohuishou/gen-const-msg

## Usage

> 详情请查看[example](./example)

`errcode.go`

```golang
package example

//go:generate  gen-const-msg
const (
	// ErrParams err params
	ErrParams = 400
	// ErrServer Internal Server Error
	ErrServer = 500
)
```

在文件夹中执行

```bash
go generate ./...
```

同目录下生成新文件`errcode_msg_gen.go`

```golang
package example

// This file is automatically generated by go generate
// Do not modify the contents of this file

// noErrorMsg if code is not found, GetMsg will return this
const noErrorMsg = "unknown error"

// messages get msg from const comment
var messages = map[int]string{

	ErrParams: "err params",
	ErrServer: "Internal Server Error",
}

// GetMsg get error msg
func GetMsg(code int) string {
	var (
		msg string
		ok  bool
	)
	if msg, ok = messages[code]; !ok {
		msg = noErrorMsg
	}
	return msg
}

```