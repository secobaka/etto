package task

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
)

func (p Priority) String() string {
	switch p {
	case PriorityHigh:
		return "high"
	case PriorityMedium:
		return "medium"
	case PriorityLow:
		return "low"
	default:
		return "low"
	}
}

func ParsePriority(s string) Priority {
	switch strings.ToLower(s) {
	case "high", "h":
		return PriorityHigh
	case "medium", "med", "m":
		return PriorityMedium
	default:
		return PriorityLow
	}
}

type Task struct {
	ID       int        `json:"id"`
	Title    string     `json:"title"`
	Due      *time.Time `json:"due,omitempty"`
	Priority Priority   `json:"priority"`
	Done     bool       `json:"done"`
}

type SortKey int

const (
	SortByDue SortKey = iota
	SortByPriority
)

type Store struct {
	path  string
	Tasks []Task
}

func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, ".etto")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	s := &Store{
		path: filepath.Join(dir, "tasks.json"),
	}

	if err := s.Load(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) Load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.Tasks = []Task{}
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &s.Tasks)
}

func (s *Store) Save() error {
	data, err := json.MarshalIndent(s.Tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) nextID() int {
	maxID := 0
	for _, t := range s.Tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	return maxID + 1
}

func (s *Store) Add(title string, due *time.Time, priority Priority) Task {
	t := Task{
		ID:       s.nextID(),
		Title:    title,
		Due:      due,
		Priority: priority,
		Done:     false,
	}
	s.Tasks = append(s.Tasks, t)
	return t
}

func (s *Store) Delete(id int) bool {
	for i, t := range s.Tasks {
		if t.ID == id {
			s.Tasks = append(s.Tasks[:i], s.Tasks[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) ToggleDone(id int) bool {
	for i, t := range s.Tasks {
		if t.ID == id {
			s.Tasks[i].Done = !t.Done
			return true
		}
	}
	return false
}

func (s *Store) Update(id int, title string, due *time.Time, priority Priority) bool {
	for i, t := range s.Tasks {
		if t.ID == id {
			s.Tasks[i].Title = title
			s.Tasks[i].Due = due
			s.Tasks[i].Priority = priority
			return true
		}
	}
	return false
}

func (s *Store) Sort(key SortKey) {
	sort.SliceStable(s.Tasks, func(i, j int) bool {
		// 未完了を先に
		if s.Tasks[i].Done != s.Tasks[j].Done {
			return !s.Tasks[i].Done
		}

		switch key {
		case SortByPriority:
			return s.Tasks[i].Priority > s.Tasks[j].Priority
		case SortByDue:
			if s.Tasks[i].Due == nil && s.Tasks[j].Due == nil {
				return false
			}
			if s.Tasks[i].Due == nil {
				return false
			}
			if s.Tasks[j].Due == nil {
				return true
			}
			return s.Tasks[i].Due.Before(*s.Tasks[j].Due)
		}
		return false
	})
}
