package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

var msg = map[string]string{}

const suffix = "_msg_gen.go"

// tpl temp
const tpl = `
// Code generated by github.com/mohuishou/gen-const-msg DO NOT EDIT

// Package {{.pkg}} const code comment msg
package {{.pkg}}

// noErrorMsg if code is not found, GetMsg will return this
const noErrorMsg = "unknown error"

// messages get msg from const comment
var messages = map[int]string{
	{{range $key, $value := .comments}}
	{{$key}}: "{{$value}}",{{end}}
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
`

func main() {
	file := os.Getenv("GOFILE")
	// 保存注释信息
	var comments = make(map[string]string)

	// 解析代码源文件，获取常量和注释之间的关系
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	checkErr(err)

	// Create an ast.CommentMap from the ast.File's comments.
	// This helps keeping the association between comments
	// and AST nodes.
	cmap := ast.NewCommentMap(fset, f, f.Comments)
	for node := range cmap {
		// 仅支持一条声明语句，一个常量的情况
		if spec, ok := node.(*ast.ValueSpec); ok && len(spec.Names) == 1 {
			// 仅提取常量的注释
			ident := spec.Names[0]
			if ident.Obj.Kind == ast.Con {
				// 获取注释信息
				comments[ident.Name] = getComment(ident.Name, spec.Doc)
			}
		}
	}

	code, err := gen(comments)
	checkErr(err)

	file = strings.TrimSuffix(file, ".go")
	file = file + suffix
	// 生成代码文件
	ioutil.WriteFile(file, code, 0644)
}

// getComment 获取注释信息
func getComment(name string, group *ast.CommentGroup) string {
	var buf bytes.Buffer

	// collect comments text
	// Note: CommentGroup.Text() does too much work for what we
	//       need and would only replace this innermost loop.
	//       Just do it explicitly.
	for _, comment := range group.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, fmt.Sprintf("// %s", name)))
		buf.WriteString(text)
	}

	// replace any invisibles with blanks
	bytes := buf.Bytes()
	for i, b := range bytes {
		switch b {
		case '\t', '\n', '\r':
			bytes[i] = ' '
		}
	}

	return string(bytes)
}

// gen 生成代码
func gen(comments map[string]string) ([]byte, error) {
	var buf = bytes.NewBufferString("")

	data := map[string]interface{}{
		"pkg":      os.Getenv("GOPACKAGE"),
		"comments": comments,
	}

	t, err := template.New("").Parse(tpl)
	if err != nil {
		return nil, errors.Wrapf(err, "template init err")
	}

	err = t.Execute(buf, data)
	if err != nil {
		return nil, errors.Wrapf(err, "template data err")
	}

	return format.Source(buf.Bytes())
}

// checkErr 检查err， 不为nil panic
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
