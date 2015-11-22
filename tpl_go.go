package main

var goTpl = `package {{.Pkg}}

var {{.Var}} = map[string]string{
{{range $k, $content := .Buf}}
	"{{$k}}": ` + "`" + `{{dueReverseQuote $content}}` + "`" + `,{{end}}
}`
