package user

type Repo interface {
	GetUser(name, password string) (string, error)
	InsertUser(name, password string) error
}

type Persistor struct {
	Repo Repo
}

func NewPersistor(repo Repo) Persistor {
	return Persistor{
		Repo: repo,
	}
}
