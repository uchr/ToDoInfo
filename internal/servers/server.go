package servers

import (
	"context"
	"github.com/gorilla/sessions"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"ToDoInfo/internal/config"
	"ToDoInfo/internal/log"
	"ToDoInfo/internal/login"
	"ToDoInfo/internal/todo"
	"ToDoInfo/internal/todoparser"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	indexTemplate *template.Template
	cfg           config.Config
	store         *sessions.CookieStore
}

func New(cfg config.Config) (*Server, error) {
	s := Server{cfg: cfg}

	var err error
	s.indexTemplate, err = template.ParseFiles("web/template/index.html")
	if err != nil {
		return nil, err
	}

	s.store = sessions.NewCookieStore([]byte(cfg.SessionKey))

	return &s, nil
}

func (s *Server) Run() error {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(s.authMiddleware)

	fs := http.FileServer(http.Dir("web/static/"))

	r.Route("/", func(r chi.Router) {
		r.Get("/", s.indexHandler())

		r.Handle("/static/*", http.StripPrefix("/static/", fs))
	})

	return http.ListenAndServe(":80", r)
}

func (s *Server) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value("token").(string)
		taskLists, err := todoparser.ParseTasks(token)

		if err != nil {
			log.Error(err)
			return
		}

		listAges := taskLists.GetListAges()
		oldestTasks := taskLists.GetTopOldestTasks(5)
		rottenTasks := taskLists.GetRottenTasks(todo.TiredTaskRottenness)
		oldestTaskForList := taskLists.GetOldestTaskForList()

		pageData := GetPageData(listAges, oldestTasks, rottenTasks, oldestTaskForList)

		err = s.indexTemplate.Execute(w, pageData)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuthFailed := r.URL.Query().Get("isAuth")
		if isAuthFailed == "0" {
			// TODO: Show page with info about failed auth and button "Try again"
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		code := r.URL.Query().Get("code")
		if code != "" {
			log.Info("Request new token")
			token, expiredIn, err := login.Auth(s.cfg, code)
			if err != nil {
				log.Error(err)
				v := url.Values{}
				v.Add("isAuth", "0")
				http.Redirect(w, r, s.cfg.RedirectURI+"?"+v.Encode(), http.StatusMovedPermanently)
				return
			}
			session, err := s.store.Get(r, "auth-session")
			if err != nil {
				log.Error(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			session.Values["token"] = token
			session.Values["expiresAt"] = time.Now().Add(expiredIn).Format(time.RFC3339)
			err = session.Save(r, w)
			if err != nil {
				log.Error(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, s.cfg.RedirectURI, http.StatusMovedPermanently)
			return
		}

		session, err := s.store.Get(r, "auth-session")
		if err != nil {
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		isExpired := false
		if !session.IsNew {
			t, err := time.Parse(time.RFC3339, session.Values["expiresAt"].(string))
			if err != nil {
				log.Error(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			isExpired = time.Now().After(t)
		}

		// TODO: Show page with auth request the new session and redirect only for expired token
		if session.IsNew || isExpired {
			log.Info("Request auth code")
			http.Redirect(w, r, login.GetAuthRequest(s.cfg), http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), "token", session.Values["token"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
