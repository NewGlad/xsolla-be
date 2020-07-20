package newsapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/NewGlad/xsolla-be/internal/app/model"
	"github.com/NewGlad/xsolla-be/internal/app/store"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

var (
	errIncorectEmailOrPassword = errors.New("Invalid email or password")
	sessionName                = "NewsApiSession"
)

const (
	ctxKeyUser ctxKey = iota
)

type ctxKey uint8

// APIServer ...
type APIServer struct {
	config       *APIConfig
	logger       *logrus.Logger
	router       *mux.Router
	store        *store.Store
	sessionStore sessions.Store
}

// New ...
func New(config *APIConfig, sessionStore sessions.Store) *APIServer {
	return &APIServer{
		config:       config,
		logger:       logrus.New(),
		router:       mux.NewRouter(),
		sessionStore: sessionStore,
	}
}

// Start ...
func (server *APIServer) Start() error {
	if err := server.configureLogger(); err != nil {
		return err
	}
	if err := server.configureStore(); err != nil {
		return err
	}
	server.configureRouter()
	server.logger.Infof(
		"Start server on addr '%v' with logging level '%v'",
		server.config.BindAddr, server.config.LogLevel,
	)
	return http.ListenAndServe(server.config.BindAddr, server.router)
}

func (server *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(server.config.LogLevel)
	if err != nil {
		return err
	}
	server.logger.SetLevel(level)
	return nil
}

func (server *APIServer) configureStore() error {
	store := store.New(server.config.StoreConfig)
	if err := store.Open(); err != nil {
		return err
	}
	server.store = store
	return nil
}

func (server *APIServer) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := server.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
		})
		startTime := time.Now()
		rw := &ResponseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)
		logger.Infof("Handle request %s %s in %v, return code %v",
			r.Method,
			r.RequestURI,
			time.Now().Sub(startTime),
			rw.code,
		)
	})
}

func (server *APIServer) configureRouter() {
	server.router.Use(server.logRequest)
	// Auth handlers
	server.router.HandleFunc("/signup", server.handleCreateUser()).Methods("POST")
	server.router.HandleFunc("/signin", server.handleCreateSession()).Methods("POST")
	// News handlers
	newsRouter := server.router.PathPrefix("/news").Subrouter()
	newsRouter.Use(server.autehtificateUser)
	newsRouter.HandleFunc("", server.handleCreateNews()).Methods("POST")
	newsRouter.HandleFunc("/{newsId:[0-9]+}", server.handleGetNewsByID()).Methods("GET")
	newsRouter.HandleFunc("/top", server.handleGetTopNews()).Methods("GET")
	// Like handlers
	newsRouter.HandleFunc("/{newsId:[0-9]+}/like", server.handleAddLike()).Methods("POST")
	newsRouter.HandleFunc("/{newsId:[0-9]+}/dislike", server.handleRemoveLike()).Methods("POST")
}

func (server *APIServer) handleGetTopNews() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topNews, err := server.store.News.GetTop(server.config.TopNewsLimit)
		if err != nil {
			server.error(w, r, http.StatusBadRequest, err)
			return
		}
		server.respond(w, r, http.StatusOK, topNews)

	}
}

func (server *APIServer) handleRemoveLike() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		newsID, err := strconv.Atoi(vars["newsId"])
		if err != nil {
			server.error(w, r, http.StatusBadRequest, errors.New("Bad given news ID"))
			return
		}
		user := r.Context().Value(ctxKeyUser).(*model.User)
		if user == nil {
			server.error(w, r, http.StatusBadRequest,
				errors.New("Can't load request context"))
			return
		}
		if err := server.store.News.RemoveLike(newsID, user.ID); err != nil {
			server.error(w, r, http.StatusBadRequest, err)
			return
		}
		server.respond(w, r, http.StatusOK, nil)

	}
}

func (server *APIServer) handleAddLike() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		newsID, err := strconv.Atoi(vars["newsId"])
		if err != nil {
			server.error(w, r, http.StatusBadRequest, errors.New("Bad given news ID"))
			return
		}
		user := r.Context().Value(ctxKeyUser).(*model.User)
		if user == nil {
			server.error(w, r, http.StatusBadRequest,
				errors.New("Can't load request context"))
			return
		}
		if err := server.store.News.AddLike(newsID, user.ID); err != nil {
			server.error(w, r, http.StatusBadRequest, err)
			return
		}
		server.respond(w, r, http.StatusOK, nil)

	}
}
func (server *APIServer) handleGetNewsByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["newsId"])
		if err != nil {
			server.error(w, r, http.StatusBadRequest, errors.New("Bad given news ID"))
			return
		}

		news, err := server.store.News.FindByID(id)
		if err != nil {
			server.error(w, r, http.StatusNotFound, err)
			return
		}
		server.respond(w, r, http.StatusOK, news)
	}
}

func (server *APIServer) handleCreateNews() http.HandlerFunc {
	type request struct {
		Content string `json:"content"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			server.error(w, r, http.StatusBadRequest, err)
			return
		}

		user := r.Context().Value(ctxKeyUser).(*model.User)
		if user == nil {
			server.error(w, r, http.StatusBadRequest,
				errors.New("Can't load request context"))
			return
		}
		news := model.News{AuthorID: user.ID, Content: req.Content}
		if err := server.store.News.Create(&news); err != nil {
			server.error(w, r, http.StatusInternalServerError, err)
			return
		}
		server.respond(w, r, http.StatusCreated, news)
	}
}

func (server *APIServer) autehtificateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := server.sessionStore.Get(r, sessionName)
		if err != nil {
			server.error(w, r, http.StatusInternalServerError, err)
			return
		}
		id, ok := session.Values["user_id"]
		if !ok {
			server.error(w, r, http.StatusUnauthorized, errors.New("Not autehtificated"))
			return
		}
		user, err := server.store.User.FindByID(id.(int))
		if err != nil {
			server.error(w, r, http.StatusUnauthorized, errors.New("Not autehtificated"))
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, user)))
	})
}

func (server *APIServer) handleCreateSession() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			server.error(w, r, http.StatusBadRequest, err)
			return
		}
		user, err := server.store.User.FindByUsername(req.Username)

		if err != nil || !user.CheckPassword(req.Password) {
			server.error(w, r, http.StatusUnauthorized, errIncorectEmailOrPassword)
			return
		}
		session, err := server.sessionStore.Get(r, sessionName)
		if err != nil {
			server.error(w, r, http.StatusInternalServerError, err)
			return
		}
		session.Values["user_id"] = user.ID
		if err := server.sessionStore.Save(r, w, session); err != nil {
			server.error(w, r, http.StatusInternalServerError, err)
		}
		server.respond(w, r, http.StatusOK, nil)
	}
}

func (server *APIServer) handleCreateUser() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		request := &request{}
		if err := json.NewDecoder(r.Body).Decode(request); err != nil {
			server.error(w, r, http.StatusBadRequest, err)
			return
		}
		user := &model.User{
			Username: request.Username,
			Password: request.Password,
		}
		if err := server.store.User.Create(user); err != nil {
			server.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		server.respond(w, r, http.StatusCreated, user)
	}

}

func (server *APIServer) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	server.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (server *APIServer) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
