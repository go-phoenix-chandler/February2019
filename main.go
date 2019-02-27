package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var (
	name    string
	version string
	gitSHA  string
)

const usage = `version: %s - git: %s
Usage: %s [-h] [-v]
Options:
  -h            this help
  -v            show version and exit
 
Examples: 
  %[3]s -t localhost:8888       run the server 
`

func main() {
	var vers bool

	flag.Usage = func() {
		w := os.Stderr
		for _, arg := range os.Args {
			if arg == "-h" {
				w = os.Stdout
				break
			}
		}
		fmt.Fprintf(w, usage, version, gitSHA, name)
	}

	flag.BoolVar(&vers, "v", false, "")
	flag.Parse()

	if vers {
		fmt.Fprintf(os.Stdout, "version: %s\n", version)
		return
	}

	router := mux.NewRouter()
	data := newStore()
	data.addTodo(todo{ID: uint32(1), Tasks: []task{task{Name: "task 1"}}})
	c := newController(data)
	ba := &BasicAuth{
		Username: "go",
		Password: "go",
	}
	router.HandleFunc(todoURL, ba.Authenticate(c.retrieveTodosHandler())).Methods(http.MethodGet)
	router.HandleFunc(todoURL, ba.Authenticate(c.addTodoHandler())).Methods(http.MethodPost)
	if err := http.ListenAndServe(":8888", router); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

type controller struct {
	store *store
}

func newController(s *store) *controller {
	return &controller{
		store: s,
	}
}

const (
	baseAPI = "/api/v1"
	todoURL = baseAPI + "/todo"
)

type todosRequest struct {
	Data struct {
		Todos []todo `json:"todos"`
	} `json:"data"`
}
type request struct{}

type task struct {
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
	Completed   bool      `json:"completed"`
}

type todo struct {
	ID       uint32 `json:"id"`
	Tasks    []task `json:"tasks"`
	Category string `json:"category"`
}

func (c *controller) retrieveTodosHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		todos := c.store.todos()
		b, err := json.Marshal(todos)
		if err != nil {
			log.Println(err)
			http.Error(w, "please try again later", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}

func (c *controller) addTodoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t todo
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			log.Println(err)
			http.Error(w, "please try again later", http.StatusInternalServerError)
			return
		}
		c.store.addTodo(t)
		w.WriteHeader(http.StatusOK)
	}
}

type store struct {
	mu   *sync.Mutex
	data map[uint32]todo
}

func newStore() *store {
	return &store{
		mu:   &sync.Mutex{},
		data: make(map[uint32]todo),
	}
}

func (s *store) todos() []todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	todos := make([]todo, len(s.data))
	for _, v := range s.data {
		todos = append(todos, v)
	}
	return todos
}

func (s *store) addTodo(t todo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[t.ID]; !ok {
		s.data[t.ID] = t
		return
	}
	s.data[t.ID] = t
	return
}

type Authenticator interface {
	Authenticate(handler http.HandlerFunc) http.HandlerFunc
}

// BasicAuth contains a username and password combination
type BasicAuth struct {
	Username string
	Password string
}

func (b *BasicAuth) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := b.authenticate(r); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (b *BasicAuth) authenticate(r *http.Request) error {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return errors.New("invalid authorization header")
	}
	if user != b.Username || pass != b.Password {
		return errors.New("invalid credentials")
	}
	return nil
}
