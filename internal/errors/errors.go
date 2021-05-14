package errors

const (
	RepoDBUnavailable        = temporaryError("Could not open database")
	RepoItemVersionConflict  = err("Item version does not match")
	RepoItemNotFound         = err("Item not found")
)

func New(s string) error {
	return err(s)
}

type err string

func (e err) Error() string {
	return string(e)
}

type temporaryError err

func (e temporaryError) Error() string {
	return string(e)
}

func (e temporaryError) IsTemporary() bool {
	return true
}

type temporary interface {
	IsTemporary() bool
}
