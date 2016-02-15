package master

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jimmykuu/wtforms"
	"net/http"
)

var (
	store        *sessions.CookieStore
	CookieSecret = "05e0ba2eca9411e18155109add4b8aac"
)

type User struct {
	Username string
}

type Status struct {
	Type       string
	Id         string
	Host       string
	Port       int
	ClientPort int
	Ready      bool
}

// 返回当前用户
func currentUser(r *http.Request) (*User, bool) {
	session, _ := store.Get(r, "user")
	username, ok := session.Values["username"]

	if !ok {
		return nil, false
	}
	user := User{username.(string)}
	return &user, true
}

// URL: /login
// 处理用户登录,如果登录成功,设置Cookie
func loginHandler(handler Handler) {

	form := wtforms.NewForm(
		wtforms.NewTextField("username", "用户名", "", &wtforms.Required{}),
		wtforms.NewPasswordField("password", "密码", &wtforms.Required{}),
	)

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {

			if form.Value("username") == "sll" &&
				form.Value("password") == "123456" {
			} else {
				form.AddError("password", "密码和用户名不匹配")

				renderHtml(handler, "login.html", map[string]interface{}{"form": form})
				return
			}

			session, _ := store.Get(handler.Request, "user")
			session.Values["username"] = form.Value("username")
			session.Save(handler.Request, handler.ResponseWriter)

			http.Redirect(handler.ResponseWriter, handler.Request, "/", http.StatusFound)

			return
		}
	}

	renderHtml(handler, "login.html", map[string]interface{}{"form": form})
}

// URL: /signout
// 用户登出,清除Cookie
func logoutHandler(handler Handler) {
	session, _ := store.Get(handler.Request, "user")
	session.Options = &sessions.Options{MaxAge: -1}
	session.Save(handler.Request, handler.ResponseWriter)
	renderHtml(handler, "login.html", map[string]interface{}{"logout": true})
}

// URL:view/{appid}/info
// 查看某个机器的具体信息
func infoHandler(handler Handler) {
	appid := mux.Vars(handler.Request)["appid"]
	if app, ok := master.app[appid]; ok {
		renderBaseHtml(handler, "base.html", "info.html", map[string]interface{}{"app": Status{app.typ, app.id, app.host, app.port, app.clientport, app.ready}})
	} else {
		renderBaseHtml(handler, "base.html", "error.html", map[string]interface{}{"error": "app not found"})
	}
}

//关闭系统
func shutdownHandler(handler Handler) {
	handler.ResponseWriter.Write([]byte("system is shutdown..."))
	go master.Stop()
}

func init() {
	store = sessions.NewCookieStore([]byte(CookieSecret))
}
