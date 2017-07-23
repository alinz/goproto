package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type protoLexer struct {
	lex *Lexer
}

func sequence(lex *Lexer, value string) bool {
	for _, v := range value {
		if lex.Next() != v {
			return false
		}
	}

	return true
}

func (p *protoLexer) Parse() []string {
	var imports []string

	for {
		p.lex.AcceptRun(" \t\n")
		p.lex.Ignore()

		if p.lex.Peek() == 0 {
			break
		}

		// ignore comments
		if p.lex.Peek() == '/' && p.lex.PeekNth(2) == '/' {
			p.lex.AcceptRunUntil("\n")
			p.lex.Next()
			p.lex.Ignore()

			continue
		}

		if !sequence(p.lex, "import") {
			p.lex.AcceptRunUntil("\n")
			p.lex.Next()
			p.lex.Ignore()
			continue
		}

		p.lex.AcceptRun(" \t\n")
		p.lex.Ignore()

		if !p.lex.Accept("\"") {
			panic("error")
		}
		p.lex.Ignore()

		p.lex.AcceptRunUntil("\"")

		imports = append(imports, p.lex.CurrentString())

		p.lex.Next()
		p.lex.Ignore()
	}

	return imports
}

func getImports(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	proto := protoLexer{
		lex: New(file, 10),
	}

	// fmt.Println("parsing", filepath)

	return proto.Parse(), nil
}

func getAllProtoPaths(target string) ([]string, error) {
	fileList := []string{}

	// the following tw lines adds the "/" to the end of path
	// this is helpful to create relative path
	target = filepath.Join(target, "/")
	target = target + "/"

	err := filepath.Walk(target, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".proto") && !strings.HasPrefix(path, "vendor/") {
			fileList = append(fileList, strings.Replace(path, target, "", -1))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileList, nil
}

func run(command string, args ...string) {
	//var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(command, args...)
	//cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(stderr.String() + "\n" + fmt.Sprint(err))
		return
	}
}

func ParseCompile(prefix string) {
	sources, err := getAllProtoPaths(".")
	if err != nil {
		panic(err)
	}

	for _, source := range sources {
		imports, err := getImports(source)
		if err != nil {
			panic(err)
		}

		filemap := ""
		for _, impr := range imports {
			filemap += fmt.Sprintf(",M%s=%s", impr, filepath.Join(prefix, filepath.Dir(impr)))
		}

		protoc := fmt.Sprintf("-I . ./%s --go_out=plugins=grpc%s:.", source, filemap)

		binary, lookErr := exec.LookPath("protoc")
		if lookErr != nil {
			panic(lookErr)
		}

		run(binary, strings.Split(protoc, " ")...)

		// optional
		binary, lookErr = exec.LookPath("protoc-go-inject-tag")
		if lookErr == nil {
			compiledSource := strings.Replace(source, ".proto", ".pb.go", 1)
			run(binary, fmt.Sprintf("-input=%s", compiledSource))
		}
	}
}
