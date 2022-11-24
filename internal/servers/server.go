package servers

import (
	"context"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"

	"github.com/uchr/ToDoInfo/internal/config"
	"github.com/uchr/ToDoInfo/internal/log"
	"github.com/uchr/ToDoInfo/internal/login"
	"github.com/uchr/ToDoInfo/internal/todo"
	"github.com/uchr/ToDoInfo/internal/todometrics"
)

type TaskProvider interface {
	GetTasks(ctx context.Context, token string) ([]todo.TaskList, error)
}

type Server struct {
	indexTemplate *template.Template
	authTemplate  *template.Template
	errorTemplate *template.Template

	cfg          config.Config
	store        *sessions.CookieStore
	taskProvider TaskProvider
}

func New(cfg config.Config, taskProvider TaskProvider) (*Server, error) {
	s := Server{
		cfg:          cfg,
		taskProvider: taskProvider,
	}

	var err error
	s.indexTemplate, err = template.ParseFiles("web/template/index.html")
	if err != nil {
		return nil, err
	}

	s.authTemplate, err = template.ParseFiles("web/template/auth.html")
	if err != nil {
		return nil, err
	}

	s.errorTemplate, err = template.ParseFiles("web/template/error.html")
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
	r.Use(log.LoggerMiddleware())
	r.Use(middleware.Recoverer)
	r.Use(s.authMiddleware)

	fs := http.FileServer(http.Dir("web/static/"))

	r.Route("/", func(r chi.Router) {
		r.Get("/", s.indexHandler())
		r.Get("/auth", s.authHandler())
		r.Get("/token", s.tokenHandler())
		r.Get("/login", s.loginHandler())
		r.Get("/logout", s.logoutHandler())

		r.Handle("/static/*", http.StripPrefix("/static/", fs))
	})

	return http.ListenAndServe(s.cfg.HostURI, r)
}

func (s *Server) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value("token").(string)
		taskLists, err := s.taskProvider.GetTasks(r.Context(), token)

		if err != nil {
			err = s.errorTemplate.Execute(w, NewErrorPageData(s.cfg.RedirectURI, http.StatusInternalServerError))

			log.Error(err)
			return
		}

		metrics := todometrics.New(taskLists)
		pageData := NewMainPageData(s.cfg.RedirectURI, metrics)

		err = s.indexTemplate.Execute(w, pageData)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

func (s *Server) authHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isAuthFailed := r.URL.Query().Get("isAuth")
		err := s.authTemplate.Execute(w,
			struct {
				RedirectURI  string
				IsAuthFailed bool
			}{
				RedirectURI:  s.cfg.RedirectURI,
				IsAuthFailed: isAuthFailed == "0",
			},
		)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

func (s *Server) tokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		log.Info("Request new token")
		token, expiredIn, err := login.Auth(r.Context(), s.cfg, code)
		if err != nil {
			log.Error(err)
			v := url.Values{}
			v.Add("isAuth", "0")
			http.Redirect(w, r, s.cfg.RedirectURI+"auth"+"?"+v.Encode(), http.StatusPermanentRedirect)
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

		http.Redirect(w, r, s.cfg.RedirectURI, http.StatusPermanentRedirect)
	}
}

func (s *Server) loginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, login.GetAuthRequest(s.cfg), http.StatusPermanentRedirect)
	}
}

func (s *Server) logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := s.store.Get(r, "auth-session")
		if err != nil {
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !session.IsNew {
			session.Options.MaxAge = -1
			err = session.Save(r, w)
			if err != nil {
				log.Error(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, s.cfg.RedirectURI, http.StatusTemporaryRedirect)
	}
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.String(), "/static") {
			next.ServeHTTP(w, r)
			return
		}

		session, err := s.store.Get(r, "auth-session")
		if err != nil {
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		authURLs := []string{"/auth", "/token", "/login"}
		for _, authURL := range authURLs {
			if strings.HasPrefix(r.URL.String(), authURL) {
				if !session.IsNew {
					http.Redirect(w, r, s.cfg.RedirectURI, http.StatusPermanentRedirect)
				}

				next.ServeHTTP(w, r)
				return
			}
		}

		if session.IsNew {
			http.Redirect(w, r, s.cfg.RedirectURI+"auth", http.StatusPermanentRedirect)
			return
		}

		t, err := time.Parse(time.RFC3339, session.Values["expiresAt"].(string))
		if err != nil {
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		isExpired := time.Now().After(t)

		if isExpired {
			log.Info("Azure token is expired")
			http.Redirect(w, r, s.cfg.RedirectURI+"login", http.StatusTemporaryRedirect)
			return
		}

		ctx := context.WithValue(r.Context(), "token", session.Values["token"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
