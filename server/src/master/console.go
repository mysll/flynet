package master

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"server/libs/log"

	"github.com/gorilla/mux"
)

var (
	ListDir = 0x0001
	routers = map[string]HandlerFunc{
		"/index":             indexHandle,
		"/login":             loginHandler,
		"/logout":            logoutHandler,
		"/view/{appid}/info": infoHandler,
		"/sys/shutdown":      shutdownHandler,
		"/":                  indexHandle,
	}

	templates = make(map[string]*template.Template)

	tplPath = ""

	master *Master
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type DiskInfo struct {
	Percent int
	Used    int
	Total   int
}

func indexHandle(handler Handler) {
	renderBaseHtml(handler, "base.html", "index.html", GetInfo())
}

func safeHandler(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*defer func() {
			if err, ok := recover().(error); ok {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()*/
		handler := NewHandler(w, r)
		_, ok := currentUser(r)

		if !ok && r.URL.Path != "/login" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		fn(handler)
	}
}

func renderHtml(handler Handler, tmpl string, data map[string]interface{}) (err error) {
	return renderBaseHtml(handler, "", tmpl, data)
}

func renderBaseHtml(handler Handler, base string, tmpl string, data map[string]interface{}) (err error) {
	var t *template.Template

	if base != "" {
		baseBytes, err := ioutil.ReadFile(tplPath + base)
		if err != nil {
			panic(err)
		}
		t, err = templates[tmpl].Clone()
		if err != nil {
			panic(err)
		}
		t, err = t.Parse(string(baseBytes))
		if err != nil {
			panic(err)
		}
	} else {
		t = templates[tmpl]
	}

	user, ok := currentUser(handler.Request)

	if ok {
		data["username"] = user.Username
		data["servers"] = master.app
	}

	err = t.Execute(handler.ResponseWriter, data)
	check(err)
	return
}

func loadTemplate(dir string) {
	tplPath = dir + "/"
	fileInfoArr, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
		return
	}
	var templateName, templatePath string
	for _, fileInfo := range fileInfoArr {
		templateName = fileInfo.Name()
		if ext := path.Ext(templateName); ext != ".html" {
			continue
		}
		templatePath = tplPath + templateName
		t := template.Must(template.ParseFiles(templatePath))
		templates[path.Base(templateName)] = t
	}
}

func isExists(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		return false
	}

	return true
}

func staticDirHandler(prefix string, staticDir string, flags int) {
	http.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		file := staticDir + r.URL.Path[len(prefix)-1:]
		if (flags & ListDir) == 0 {
			if exists := isExists(file); !exists {
				http.NotFound(w, r)
				return
			}
		}
		http.ServeFile(w, r, file)
	})
}

func StartConsoleServer(m *Master) {
	master = m
	log.LogMessage("console root:", m.Template)
	loadTemplate(m.Template)
	r := mux.NewRouter()
	staticDirHandler("/static/", m.Template+"/static", 1)
	staticDirHandler("/img/", m.Template+"/img", 1)
	staticDirHandler("/js/", m.Template+"/js", 1)
	staticDirHandler("/css/", m.Template+"/css", 1)
	staticDirHandler("/font/", m.Template+"/font", 1)
	for k, h := range routers {
		r.HandleFunc(k, safeHandler(h))
	}

	http.Handle("/", r)
	go http.ListenAndServe(fmt.Sprintf(":%d", m.ConsolePort), nil)
}
