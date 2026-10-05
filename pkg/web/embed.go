package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed static/* templates/*
var contentFS embed.FS

// GetStaticFS 返回静态资源文件系统
func GetStaticFS() http.FileSystem {
	sub, err := fs.Sub(contentFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

// GetTemplate 解析并返回所有 HTML 模板
func GetTemplates() (*template.Template, error) {
	return template.ParseFS(contentFS, "templates/*.html")
}

// GetStaticFile 读取单个静态嵌入文件内容
func GetStaticFile(name string) ([]byte, error) {
	return contentFS.ReadFile("static/" + name)
}

