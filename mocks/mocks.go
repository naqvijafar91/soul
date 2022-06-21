package mocks

import (
	"soul"

	"fyne.io/fyne/v2/data/binding"
	"github.com/stretchr/testify/mock"
)

func NewNoteRepository() *NoteRepository {
	repo := new(NoteRepository)

	var initStrFunc = func(str string) binding.String {
		val := binding.NewString()
		val.Set(str)

		return val
	}

	repo.On("GetAll").Return([]soul.Note{
		{
			ID:      "1",
			Text:    initStrFunc("first string"),
			Version: 1,
		},
		{
			ID:      "2",
			Text:    initStrFunc("second string"),
			Version: 1,
		},
		{
			ID:      "3",
			Text:    initStrFunc("third string"),
			Version: 1,
		},
		{
			ID:      "4",
			Text:    initStrFunc("fourth string"),
			Version: 1,
		},
	}, nil)

	repo.On("Create", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		note := args[0].(*soul.Note)
		note.ID = "new Note"
	})

	return repo
}
