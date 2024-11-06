package spin

import nodeart_spin "NodeArt/internal/db/spin_repo/spin"

type Repo interface {
	AddSpin(email string, combination string) error
	GetSpinHistory(email string) ([]nodeart_spin.Spin, error)
}

type Persistor struct {
	Repo Repo
}

func NewPersistor(repo Repo) Persistor {
	return Persistor{
		Repo: repo,
	}
}
