package maker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Maker struct {
	rootPath string
}

func NewMaker(rootPath string) *Maker {
	return &Maker{rootPath: rootPath}
}

func (m *Maker) MakeController(name string) error {
	dir := filepath.Join(m.rootPath, "internal", "controller")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(dir, strings.ToLower(name)+".go")
	content := fmt.Sprintf(`package controller

import (
	"net/http"

	"github.com/gmcorenet/framework/kernel"
)

type %sController struct {
	kernel.BaseController
}

func New%sController() *%sController {
	return &%sController{}
}

func (c *%sController) Index(w http.ResponseWriter, r *http.Request, params map[string]string) {
	c.JSON(r.Context(), http.StatusOK, map[string]string{"message": "Hello from %s"})
}
`, name, name, name, name, name, name)

	return os.WriteFile(filename, []byte(content), 0644)
}

func (m *Maker) MakeService(name string) error {
	dir := filepath.Join(m.rootPath, "internal", "service")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(dir, strings.ToLower(name)+".go")
	content := fmt.Sprintf(`package service

type %sService struct {
}

func New%sService() *%sService {
	return &%sService{}
}
`, name, name, name, name)

	return os.WriteFile(filename, []byte(content), 0644)
}

func (m *Maker) MakeModel(name string) error {
	dir := filepath.Join(m.rootPath, "internal", "model")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(dir, strings.ToLower(name)+".go")
	content := fmt.Sprintf(`package model

type %s struct {
	ID   uint `+"`gorm:\"primaryKey\"`/"+`
}

func (%s) TableName() string {
	return "%ss"
}
`, name, name, strings.ToLower(name))

	return os.WriteFile(filename, []byte(content), 0644)
}

func (m *Maker) MakeMiddleware(name string) error {
	dir := filepath.Join(m.rootPath, "internal", "middleware")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(dir, strings.ToLower(name)+".go")
	content := fmt.Sprintf(`package middleware

import (
	"net/http"
)

func %s(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Insert middleware logic here when generating app code.
		next.ServeHTTP(w, r)
	})
}
`, name)

	return os.WriteFile(filename, []byte(content), 0644)
}

func (m *Maker) MakeCommand(name string) error {
	dir := filepath.Join(m.rootPath, "internal", "command")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(dir, strings.ToLower(name)+".go")
	content := fmt.Sprintf(`package command

import (
	"fmt"
)

type %sCommand struct {
}

func New%sCommand() *%sCommand {
	return &%sCommand{}
}

func (c *%sCommand) Name() string {
	return "%s"
}

func (c *%sCommand) Execute(args []string) error {
	fmt.Println("Executing %s command...")
	return nil
}
`, name, name, name, name, name, strings.ToLower(name), name, name)

	return os.WriteFile(filename, []byte(content), 0644)
}

func (m *Maker) MakeEntity(name string, fields []Field) error {
	dir := filepath.Join(m.rootPath, "internal", "model")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var fieldLines []string
	for _, f := range fields {
		fieldLines = append(fieldLines, fmt.Sprintf("\t%s %s `gorm:\"%s\"`", f.Name, f.Type, f.Tag))
	}

	filename := filepath.Join(dir, strings.ToLower(name)+".go")
	content := fmt.Sprintf(`package model

type %s struct {
	ID   uint `+"`gorm:\"primaryKey\"`\n"+`%s
}

func (%s) TableName() string {
	return "%ss"
}
`, name, strings.Join(fieldLines, "\n"), name, strings.ToLower(name))

	return os.WriteFile(filename, []byte(content), 0644)
}

type Field struct {
	Name string
	Type string
	Tag  string
}

func (m *Maker) MakeBundle(name string) error {
	dir := filepath.Join(m.rootPath, "bundles", strings.ToLower(name))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	files := map[string]string{
		"bundle.go": fmt.Sprintf(`package %s

import "github.com/gmcorenet/framework/kernel"

type %sBundle struct{}

func New%sBundle() *%sBundle {
	return &%sBundle{}
}

func (b *%sBundle) Name() string {
	return "%s"
}

func (b *%sBundle) Boot(ctx context.Context) error {
	return nil
}

func (b *%sBundle) Shutdown() error {
	return nil
}
`, name, name, name, name, name, name, name, name, name),
	}

	for filename, content := range files {
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (m *Maker) MakeCRUDController(name string) error {
	dir := filepath.Join(m.rootPath, "internal", "controller")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(dir, strings.ToLower(name)+"_controller.go")
	content := fmt.Sprintf(`package controller

import (
	"net/http"

	"github.com/gmcorenet/framework/kernel"
	"github.com/gmcorenet/bundle-crud/internal/crud"
)

type %sController struct {
	kernel.BaseController
	crud *crud.CRUD
}

func New%sController(c *crud.CRUD) *%sController {
	return &%sController{crud: c}
}

func (c *%sController) Index(w http.ResponseWriter, r *http.Request, params map[string]string) {
	c.JSON(r.Context(), http.StatusOK, map[string]string{"message": "List"})
}

func (c *%sController) Show(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id := params["id"]
	c.JSON(r.Context(), http.StatusOK, map[string]string{"message": "Show " + id})
}

func (c *%sController) Create(w http.ResponseWriter, r *http.Request, params map[string]string) {
	c.JSON(r.Context(), http.StatusOK, map[string]string{"message": "Create"})
}

func (c *%sController) Update(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id := params["id"]
	c.JSON(r.Context(), http.StatusOK, map[string]string{"message": "Update " + id})
}

func (c *%sController) Delete(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id := params["id"]
	c.JSON(r.Context(), http.StatusOK, map[string]string{"message": "Delete " + id})
}
`, name, name, name, name, name, name, name, name, name)

	return os.WriteFile(filename, []byte(content), 0644)
}
