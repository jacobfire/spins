package balance

type Repo interface {
	AddDeposit(email string, amount float64) (float64, error)
	SubDeposit(email string, amount float64) (float64, error)
	GetBalance(email string) (float64, error)
	UpdateDeposit(email string, amount float64) error
	UpdateWithNewValueDeposit(email string, amount float64) error
}

type Persistor struct {
	Repo Repo
}

func NewPersistor(repo Repo) Persistor {
	return Persistor{
		Repo: repo,
	}
}
