package servers

import (
	"ToDoInfo/internal/log"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"ToDoInfo/internal/todo"
	"ToDoInfo/internal/todoparser"
)

type Server struct {
	indexTemplate *template.Template
	token         string
}

func New(token string) (*Server, error) {
	s := Server{token: token}
	var err error
	s.indexTemplate, err = template.ParseFiles("web/template/index.html")
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Server) Run() error {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	fs := http.FileServer(http.Dir("web/static/"))

	r.Route("/", func(r chi.Router) {
		r.Get("/", s.indexHandler())

		r.Handle("/static/*", http.StripPrefix("/static/", fs))
	})

	return http.ListenAndServe(":80", r)
}

func (s *Server) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskLists, err := todoparser.ParseTasks(s.token)
		if err != nil {
			log.Error(err)
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
