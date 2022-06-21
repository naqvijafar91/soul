module soul

go 1.15

replace github.com/jonhadfield/gosn-v2 => ./../gosn-v2

require (
	fyne.io/fyne/v2 v2.2.1
	github.com/boltdb/bolt v1.3.1
	github.com/brianvoe/gofakeit/v6 v6.16.0
	github.com/google/uuid v1.3.0
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.2
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
)
