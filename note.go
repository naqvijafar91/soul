package soul

import (
	"errors"
	"strings"

	"fyne.io/fyne/v2/data/binding"
)

// Version gives version
type Version uint

// Note is a notes struct
type Note struct {
	ID      string
	Text    binding.String
	Version Version
}

// NoteRepository is a repository of notes
type NoteRepository interface {
	GetAll() ([]Note, error)
	Create(note *Note) error
	Update(note *Note) error
}

type NoteService struct {
	Repo  NoteRepository
	Notes []Note
}

func (ns *NoteService) LoadAll() error {
	notes, err := ns.Repo.GetAll()
	if err != nil {
		return err
	}

	ns.Notes = notes

	return nil
}

func (ns *NoteService) Create() (*Note, error) {
	note := Note{Text: binding.NewString()}
	err := ns.Repo.Create(&note)
	if err != nil {
		return nil, err
	}

	ns.Notes = append(ns.Notes, note)

	return &note, nil
}

func (ns *NoteService) Update(note *Note) error {
	return ns.Repo.Update(note)
}

// NewNoteService creates a new NoteService
func NewNoteService(repo NoteRepository) *NoteService {
	return &NoteService{Repo: repo}
}

func (note *Note) Title() binding.String {
	return newTitleString(note.Text)
}

type titleString struct {
	binding.String
}

func (t *titleString) Get() (string, error) {
	content, err := t.String.Get()
	if err != nil {
		return "Error", err
	}

	if content == "" {
		return "Untitled", nil
	}

	return strings.SplitN(content, "\n", 2)[0], nil
}

func (t *titleString) Set(string) error {
	return errors.New("cannot set content from title")
}

func newTitleString(in binding.String) binding.String {
	return &titleString{in}
}

func NewBindingFromString(str string) binding.String {
	val := binding.NewString()
	val.Set(str)

	return val
}
