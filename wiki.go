package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"html/template"
	"regexp"
	"log"
)

const lenPath = len("/view/")

type page struct {
	Title	string
	Body	[]byte
}

var templates map[string]*template.Template = make(map[string]*template.Template)
var titleValidator *regexp.Regexp

func (p *page) save() error {
	fileName := "data/" + p.Title + ".txt"
	// WriteFile的第三个参数为八进制的0600
	// 表示仅当前用户拥有新创建文件的读写权限。(参考Unix手册 open(2) )
	return ioutil.WriteFile(fileName, p.Body, 0600)
}

func loadPage(title string) (*page, error) {
	fileName := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return &page{title, body}, nil
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[lenPath:]
		if !titleValidator.MatchString(title) {
			http.Error(w, "Invalid Page Title", http.StatusBadRequest)
			return
		}

		fn(w, r, title)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *page) {
	err := templates[tmpl].Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/" + title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &page{Title: title}
	}

	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &page{title, []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", frontPageHandler)

	log.Print("Starting server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func init() {
	var err error

	for _, tmpl := range []string{"edit", "view"} {
		templates[tmpl], err = template.ParseFiles("tmpl/" + tmpl + ".html")
		if err != nil {
			fmt.Printf("init failed, %v\n", err)
			return
		}
	}

	titleValidator = regexp.MustCompile(`^[a-zA-Z]+$`)
}