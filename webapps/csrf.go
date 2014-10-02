package webapps

import (
	"html/template"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/srhnsn/go-utils/misc"
)

const csrfTokenLength = 12

func getCsrfHTML(r *http.Request) template.HTML {
	session := context.Get(r, "session").(*sessions.Session)
	csrfToken, ok := session.Values["csrf"].(string)

	if !ok {
		csrfToken = ""
	}

	return template.HTML(`<input name="csrf" type="hidden" value="` + csrfToken + `">`)
}

func isCsrfValid(r *http.Request, session *sessions.Session) bool {
	if r.Method == "GET" || r.Method == "HEAD" {
		return true
	}

	r.ParseForm()

	correctToken, ok := session.Values["csrf"].(string)
	submittedToken := r.PostForm.Get("csrf")

	return ok && correctToken == submittedToken && len(correctToken) == csrfTokenLength
}

func setCsrfToken(session *sessions.Session) {
	token, ok := session.Values["csrf"].(string)

	if !ok || len(token) < csrfTokenLength {
		session.Values["csrf"] = misc.GetRandomString(csrfTokenLength)
	}
}
