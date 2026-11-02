package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{
		mu:    sync.RWMutex{},
		clock: clock,
		seq:   0,
		tasks: make(map[string]Task),
	}
}
func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	validatedTitle := strings.TrimSpace(title)
	if len(validatedTitle) == 0 {
		return Task{}, ErrInvalidTitle
	}
	r.seq++

	currentId := strconv.Itoa(int(r.seq))
	newTask := Task{
		ID:        currentId,
		Title:     validatedTitle,
		UpdatedAt: r.clock.Now(),
		Done:      false,
	}

	r.tasks[currentId] = newTask

	return newTask, nil
}

func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]

	if ok {
		return task, ok
	}

	return Task{}, ok
}
func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resultTask := make([]Task, 0, 10)
	for _, task := range r.tasks {
		if task.Done == done {
			resultTask = append(resultTask, task)
		}
	}

	return resultTask
}
func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]

	if !ok {
		return Task{}, ErrNotFound
	}

	if task.Done != done {
		task.UpdatedAt = r.clock.Now()
	}

	task.Done = done
	r.tasks[id] = task

	return task, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &httpHandler{repo: repo}
}
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	url := r.URL
	paths := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")

	if len(paths) == 0 || paths[0] != "tasks" || len(paths) > 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch method {
	case http.MethodGet:
		query := url.Query()
		if len(paths) > 1 {
			if len(paths[1]) == 0 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			h.handleGet(w, r, paths[1])
		} else if !query.Has("done") || (query.Has("done") && len(query["done"]) == 1 &&
			(query.Get("done") == "false" || query.Get("done") == "true")) {
			h.handleList(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case http.MethodPost:
		h.handleCreate(w, r)
	case http.MethodPatch:
		if len(paths) > 1 {
			if len(paths[1]) == 0 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			h.handlePatch(w, r, paths[1])
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

type JsonCreateReq struct {
	Title *string `json:"title"`
}

func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var data JsonCreateReq
	err := decodeStrictJSON(r.Body, &data)
	defer r.Body.Close()

	if err != nil || data.Title == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	task, err := h.repo.Create(*data.Title)

	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusCreated, task)
}
func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, task)
}
func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	valueDone := query.Get("done")
	done := false

	switch {
	case len(valueDone) == 0:
		done = false
	case valueDone == "true":
		done = true
	case valueDone == "false":
		done = false
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	doneList := h.repo.List(done)

	slices.SortFunc(doneList, func(e1 Task, e2 Task) int {
		updateComp := e2.UpdatedAt.Compare(e1.UpdatedAt)

		if updateComp != 0 {
			return updateComp
		}

		return cmp.Compare(e1.ID, e2.ID)
	})

	writeJSON(w, http.StatusOK, doneList)
}

type JsonUpdateReq struct {
	Done *bool `json:"done"`
}

func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var data JsonUpdateReq
	err := decodeStrictJSON(r.Body, &data)
	defer r.Body.Close()

	if err != nil || data.Done == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *data.Done)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func decodeStrictJSON(r io.Reader, v any) error {

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	err := dec.Decode(v)
	if err != nil {
		return err
	}

	var anyTrashData any
	err = dec.Decode(anyTrashData)
	if !errors.Is(err, io.EOF) {
		return errors.New("trash data in json detected")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, err = w.Write(data)
	if err != nil {
		return
	}
}
