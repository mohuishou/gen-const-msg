package main

import (
	"bytes"
	"flag"
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

var (
	out       string
	constType string
	md        bool
	mdOut     string
)

func init() {
	flag.StringVar(&out, "o", "", "output file path, default: filename_msg_gen.go")
	flag.StringVar(&constType, "t", "int", "the err code type")
	flag.BoolVar(&md, "m", false, "out markdown file")
	flag.StringVar(&mdOut, "md-out", "", "out markdown file path, default: filename_msg_gen.md")
	flag.Parse()
}

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

	if out == "" {
		out = strings.TrimSuffix(file, ".go") + suffix + ".go"
	}

	if mdOut == "" {
		mdOut = strings.TrimSuffix(file, ".go") + suffix + ".md"
	}

	// 生成代码文件
	checkErr(ioutil.WriteFile(out, code, 0644))

	if md {
		data, err := genMd(comments)
		checkErr(err)
		checkErr(ioutil.WriteFile(mdOut, data, 0644))
	}
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
	bs := buf.Bytes()
	for i, b := range bs {
		switch b {
		case '\t', '\n', '\r':
			bs[i] = ' '
		}
	}
	return string(bs)
}

// genMd 生成代码
func genMd(comments map[string]string) ([]byte, error) {
	var buf = bytes.NewBufferString("")

	data := map[string]interface{}{
		"pkg":      os.Getenv("GOPACKAGE"),
		"comments": comments,
	}

	t, err := template.New("").Parse(mdTpl)
	if err != nil {
		return nil, errors.Wrapf(err, "template init err")
	}

	err = t.Execute(buf, data)
	if err != nil {
		return nil, errors.Wrapf(err, "template data err")
	}

	return buf.Bytes(), nil
}

// gen 生成代码
func gen(comments map[string]string) ([]byte, error) {
	var (
		t   *template.Template
		err error
		buf = bytes.NewBufferString("")
	)

	data := map[string]interface{}{
		"pkg":       os.Getenv("GOPACKAGE"),
		"comments":  comments,
		"constType": constType,
	}

	t, err = template.New("").Parse(tpl)
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
		panic(fmt.Sprintf("err: %+v", err))
	}
}
