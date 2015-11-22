package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	NAME = "file2string"
)

var (
	pkg      string
	filePath string
	varName  string
)

var (
	// buffers store all files contents
	buffers = make(map[string]string)

	// template funcMaps
	funcMaps = map[string]interface{}{
		"dueReverseQuote": func(text string) string {
			buffer := new(bytes.Buffer)
			for {
				pos := strings.IndexRune(text, '`')
				if pos == -1 {
					buffer.WriteString(text)
					break
				}
				buffer.WriteString(text[:pos] + "` + \"`\" + `")
				text = text[pos+1:]
			}
			return buffer.String()
		},
	}
)

func init() {
	flag.StringVar(&pkg, "pkg", "main", "define the package of file")
	flag.StringVar(&filePath, "o", "text.go", "define the generated file path")
	flag.StringVar(&varName, "var", "text", "define the variable name")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if err := checkArgs(args); err != nil {
		fmt.Println(err)
		return
	}

	createDir(filePath)

	tmpFileName := writeTmpFile(filePath, args)
	defer os.Remove(tmpFileName)

	writeTargetFile(filePath, tmpFileName)
}

func checkArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Please specifiy files.")
	}

	for _, arg := range args {
		if isFileNotExist(arg) {
			return fmt.Errorf("File %s not exists.", arg)
		}

		// check whether a textual file
		buf, err := ioutil.ReadFile(arg)
		if err != nil {
			panic(err)
		}

		typ := http.DetectContentType(buf)
		if !strings.HasPrefix(typ, "text") {
			return fmt.Errorf("File %s isn't textual file.", arg)
		}

		buffers[arg] = string(buf)
	}

	return nil
}

func createDir(filePath string) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0777); err != nil {
		panic(err)
	}

	if !isFileNotExist(filePath) {
		fmt.Print("File has existed, do you want to rewrite the file? (y/n): ")
	loop:
		for {
			var input string
			fmt.Scan(&input)
			switch input {
			case "y", "Y":
				break loop
			case "n", "N":
				fmt.Println("Abort.")
				os.Exit(0)
			default:
				fmt.Print("Please input y or n: ")
			}
		}
	}
}

func isFileNotExist(filePath string) bool {
	f, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return true
	}
	defer f.Close()

	fInfo, err := f.Stat()
	if err != nil {
		panic(err)
	}

	if fInfo.IsDir() {
		fmt.Printf("%s is a directory.\n", filePath)
		os.Exit(0)
	}

	return false
}

func writeTmpFile(filePath string, args []string) string {
	f, err := ioutil.TempFile(".", "."+NAME+"_tmp_")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	data := map[string]interface{}{
		"Pkg": pkg,
		"Var": varName,
		"Buf": buffers,
	}

	tpl := template.Must(template.New("tpl_go").Funcs(funcMaps).Parse(goTpl))
	err = tpl.Execute(f, data)
	if err != nil {
		panic(err)
	}

	return f.Name()
}

func writeTargetFile(filePath, tmpFileName string) {
	fw, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	fr, err := os.Open(tmpFileName)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(fw, fr)
	if err != nil {
		panic(err)
	}
}
