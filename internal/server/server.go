package server

import (
	"NodeArt/internal"
	"NodeArt/internal/auth"
	"NodeArt/internal/db"
	"NodeArt/internal/model/balance"
	spinModel "NodeArt/internal/model/spin"
	user2 "NodeArt/internal/model/user"
	"NodeArt/internal/spin"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"sync"
)

type webServer struct {
	host             string
	port             string
	userPersistor    user2.Persistor
	balancePersistor balance.Persistor
	spinPersistor    spinModel.Persistor
}

type Config struct {
	Host string
	Port string
	Conn *pgx.Conn
}

func userPersistor(c *pgx.Conn) user2.Persistor {
	storage := db.Storage{
		Conn: c,
	}
	return user2.NewPersistor(storage)
}

func balancePersistor(c *pgx.Conn) balance.Persistor {
	storage := db.BalanceStorage{
		Conn: c,
	}
	return balance.NewPersistor(storage)
}

func spinPersistor(c *pgx.Conn) spinModel.Persistor {
	storage := db.SpinStorage{
		Conn: c,
	}
	return spinModel.NewPersistor(storage)
}

func New(c Config) *webServer {
	return &webServer{
		host:             c.Host,
		port:             c.Port,
		userPersistor:    userPersistor(c.Conn),
		balancePersistor: balancePersistor(c.Conn),
		spinPersistor:    spinPersistor(c.Conn),
	}
}

func (w *webServer) Run() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/api/register", w.register)
	r.POST("/api/login", w.login)
	r.GET("/api/profile", auth.AuthenticateMiddleware, w.profile)

	r.POST("/api/wallet/deposit", auth.AuthenticateMiddleware, w.deposit)
	r.POST("/api/wallet/withdraw", auth.AuthenticateMiddleware, w.withdraw)

	r.POST("/api/slot/spin", auth.AuthenticateMiddleware, w.spin)
	r.POST("/api/slot/history", auth.AuthenticateMiddleware, w.history)

	r.Run(fmt.Sprintf("%s:%s", w.host, w.port))
}

func (w *webServer) register(c *gin.Context) {
	var signUp user2.User
	if err := c.ShouldBindJSON(&signUp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response": "wrong credentials",
		})
		return
	}

	err := w.userPersistor.Repo.InsertUser(signUp.Username, signUp.Password)
	if errors.Is(err, internal.UserAlreadyExists) {
		msg := fmt.Sprintf("input error: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"response": msg,
		})
		return
	}
	if err != nil {
		msg := fmt.Sprintf("can't save credentials: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"response": msg,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": "created",
	})
}

func (w *webServer) login(c *gin.Context) {
	var signIn user2.User
	if err := c.ShouldBindJSON(&signIn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response": "wrong credentials",
		})
		return
	}

	email, err := w.userPersistor.Repo.GetUser(signIn.Username, signIn.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response": "can't save credentials",
		})
		return
	}

	token, err := auth.CreateToken(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response": "token generation failed",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (w *webServer) profile(c *gin.Context) {
	user, ok := c.Get("current_user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "not authorized user",
		})
	}
	email := user.(string)

	newDeposit, err := w.balancePersistor.Repo.GetBalance(email)
	msg := fmt.Sprintf("can't get deposit: %s", err)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"response": msg,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deposit": newDeposit,
	})
}

func (w *webServer) deposit(c *gin.Context) {
	type Wallet struct {
		Deposit float64 `json:"deposit"`
	}
	wallet := Wallet{}
	if err := c.ShouldBindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response": "wrong deposit",
		})
		return
	}

	log.Printf("Wallet: %v", wallet)

	if wallet.Deposit < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "deposit bad value",
		})
		return
	}

	user, ok := c.Get("current_user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "not authorized user",
		})
	}
	email := user.(string)

	deposit, err := w.balancePersistor.Repo.GetBalance(email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "currently can't get deposit",
		})
		return
	}

	if deposit == 0 {
		_, err := w.balancePersistor.Repo.AddDeposit(email, wallet.Deposit)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "can't update wallet",
			})
			return
		}
	} else {
		err := w.balancePersistor.Repo.UpdateDeposit(email, wallet.Deposit)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "can't update wallet",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"deposit": "updated",
	})
}

func (w *webServer) withdraw(c *gin.Context) {
	type Wallet struct {
		Deposit float64 `json:"deposit"`
	}
	wallet := Wallet{}
	if err := c.ShouldBindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response": "wrong wallet",
		})
		return
	}

	log.Printf("Wallet: %v", wallet)

	if wallet.Deposit < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "deposit bad value",
		})
		return
	}

	user, ok := c.Get("current_user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "not authorized user",
		})
	}
	email := user.(string)

	d, err := w.balancePersistor.Repo.SubDeposit(email, wallet.Deposit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "cant sub wallet",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"deposit": d,
	})
}

func (w *webServer) spin(c *gin.Context) {
	type Bet struct {
		Amount float64 `json:"amount"`
	}
	bet := Bet{}
	if err := c.ShouldBindJSON(&bet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"response": "wrong bet amount",
		})
		return
	}

	log.Printf("Bet amount: %v", bet.Amount)

	if bet.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "bad bet amount",
		})
		return
	}
	user, ok := c.Get("current_user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "not authorized user",
		})
	}
	email := user.(string)

	mu := sync.Mutex{}
	mu.Lock()

	deposit, err := w.balancePersistor.Repo.GetBalance(email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "cant update wallet",
		})
		mu.Unlock()
		return
	}

	if bet.Amount > deposit {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "bet amount exceed deposit",
		})
		mu.Unlock()
		return
	}
	var newBalance float64
	sp := spin.Spin{}
	combo := 0
	slotResult, combination := sp.Spin()

	log.Printf("ADDSPIN: %d", combination)

	err = w.spinPersistor.Repo.AddSpin(email, combination)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "cant add spin info",
		})
		mu.Unlock()
		return
	}
	switch slotResult {
	case 0:
		newBalance = deposit - bet.Amount
	case 1:
		combo = 2
		newBalance = deposit + bet.Amount*2
	case 2:
		combo = 3
		newBalance = deposit + bet.Amount*10
	}

	log.Printf("COMBO: %d", combo)
	log.Printf("BET: %d", bet.Amount)
	log.Printf("DEPOSIT: %d", deposit)
	log.Printf("new balance: %d", newBalance)

	err = w.balancePersistor.Repo.UpdateWithNewValueDeposit(email, newBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "cant update deposit",
		})
		return
	}

	mu.Unlock()
	var result string
	if combo > 0 {
		result = "win"
	} else {
		result = "loss"
	}
	c.JSON(http.StatusOK, gin.H{
		"combo":      combo,
		"result":     result,
		"bet":        bet.Amount,
		"oldBalance": deposit,
		"newBalance": newBalance,
	})
}

func (w *webServer) history(c *gin.Context) {
	user, ok := c.Get("current_user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "not authorized user",
		})
	}
	email := user.(string)

	spins, err := w.spinPersistor.Repo.GetSpinHistory(email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "cant update wallet",
		})
		return
	}

	for idx, slot := range spins {
		fmt.Printf("%d: comb %d - date %s", idx, slot.Combination, slot.CreatedAt.Time.String())
	}
}
