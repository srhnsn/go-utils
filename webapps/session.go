package webapps

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/srhnsn/go-utils/i18n"
	"github.com/srhnsn/go-utils/log"
)

type SessionConfig struct {
	Routes http.Handler
	Secret string
}

var sessionConfig SessionConfig
var cookieOptions = sessions.Options{
	Path:     "/",
	MaxAge:   0,
	HttpOnly: true,
}
var store *sessions.CookieStore
var UserSessionHandler func(r *http.Request, session *sessions.Session)

func InitSessions(config SessionConfig) {
	sessionConfig = config
	store = sessions.NewCookieStore([]byte(config.Secret))
}

func FlashMessage(r *http.Request, w http.ResponseWriter) {
	SendResponse("flash", r, w)
}

var RequestHandlerFunc http.HandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	defer context.Clear(r)

	session := getSession(r)
	setSessionOptions(session)
	setCsrfToken(session)
	context.Set(r, "session", session)

	data := getDefaultTemplateData(r)
	context.Set(r, "data", data)

	setTranslateFuncs(r, data)
	SetCspHeader(w)

	w.Header().Set("Cache-Control", "no-cache, must-revalidate")

	if !isCsrfValid(r, session) {
		session.Save(r, w)
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		log.Warning.Println("CSRF error")
		return
	}

	if UserSessionHandler != nil {
		UserSessionHandler(r, session)
	}

	sessionConfig.Routes.ServeHTTP(w, r)
})

func SendResponse(templateName string, r *http.Request, w http.ResponseWriter) {
	data := context.Get(r, "data").(TemplateData)
	session := context.Get(r, "session").(*sessions.Session)
	tpl := GetTemplate(templateName, r)

	session.Save(r, w)
	sanitizeData(data)

	err := tpl.Execute(w, data)

	if err != nil {
		log.Error.Panicf(`Execution of template "%s" failed: %s`, templateName, err)
	}
}

func GetTemplateData(r *http.Request) TemplateData {
	return context.Get(r, "data").(TemplateData)
}

func GetFutureT(r *http.Request) i18n.FutureTranslateFunc {
	data := GetTemplateData(r)
	return data["FutureT"].(i18n.FutureTranslateFunc)
}

func GetStringT(r *http.Request) i18n.DataWrappedTranslateFunc {
	data := GetTemplateData(r)
	return data["StringT"].(i18n.DataWrappedTranslateFunc)
}

func GetUnsafeT(r *http.Request) i18n.UnescapedTranslateFunc {
	data := GetTemplateData(r)
	return data["UnsafeT"].(i18n.UnescapedTranslateFunc)
}

func getSession(r *http.Request) *sessions.Session {
	session, _ := store.Get(r, "session")
	return session
}

func setTranslateFuncs(r *http.Request, data TemplateData) {
	T := i18n.GetTranslationFunc(r)
	dataWrappedT := i18n.GetDataWrappedTranslateFunc(T, data)
	unescapedT := i18n.GetUnescapedTranslatFunc(dataWrappedT)
	futureT := i18n.GetFutureTranslateFunc(unescapedT)

	data["FutureT"] = futureT
	data["StringT"] = dataWrappedT
	data["UnsafeT"] = unescapedT
}

func setSessionOptions(session *sessions.Session) {
	session.Options = &cookieOptions
}
