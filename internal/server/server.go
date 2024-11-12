package server

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"NodeArt/docs"
	"NodeArt/internal"
	"NodeArt/internal/auth"
	"NodeArt/internal/db"
	"NodeArt/internal/model/balance"
	spinModel "NodeArt/internal/model/spin"
	userModel "NodeArt/internal/model/user"
	"NodeArt/internal/spin"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/swaggo/files"       // swagger embed files
	"github.com/swaggo/gin-swagger" // gin-swagger middleware
)

type webServer struct {
	host             string
	port             string
	userPersistor    userModel.Persistor
	balancePersistor balance.Persistor
	spinPersistor    spinModel.Persistor
}

type Config struct {
	Host string
	Port string
	Conn *pgx.Conn
}

func userPersistor(c *pgx.Conn) userModel.Persistor {
	storage := db.Storage{
		Conn: c,
	}
	return userModel.NewPersistor(storage)
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

// Run godoc
// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apiKey JWT
// @in header
// @name token

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func (w *webServer) Run() {
	// programmatically set swagger info
	docs.SwaggerInfo.Title = "Swagger Slot API"
	docs.SwaggerInfo.Description = "This is a slot server."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	r := gin.Default()

	api := r.Group("/api")
	{
		r.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
		api.POST("/register", w.register)
		api.POST("/login", w.login)
		api.GET("/profile", auth.AuthenticateMiddleware, w.profile)

		walletGroup := api.Group("wallet")
		{
			walletGroup.POST("/deposit", auth.AuthenticateMiddleware, w.deposit)
			walletGroup.POST("/withdraw", auth.AuthenticateMiddleware, w.withdraw)
		}

		slotGroup := api.Group("slot")
		{
			slotGroup.POST("/spin", auth.AuthenticateMiddleware, w.spin)
			slotGroup.POST("/history", auth.AuthenticateMiddleware, w.history)
		}
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(fmt.Sprintf("%s:%s", w.host, w.port))
}

// Register godoc
// @Summary      Register an account
// @Description  register new account by email and password
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param payload body userModel.User true "username and password params"
// @Success      200  {object}  JSONResult "desc"
// @Failure      400  {object}  JSONResult "desc"
// @Failure      500  {object}  JSONResult "desc"
// @Router       /register [post]
func (w *webServer) register(c *gin.Context) {
	var signUp userModel.User
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

// Login godoc
// @Summary      Log in to an account
// @Description  login with email and password
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param payload body userModel.User true "username and password params"
// @Success      200  {object}   TokenResult "desc"
// @Failure      400  {object}   JSONResult "desc"
// @Failure      500  {object}   JSONResult "desc"
// @Router       /login [post]
func (w *webServer) login(c *gin.Context) {
	var signIn userModel.User
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

// Profile godoc
// @Summary      review details of a profile
// @Description  profile shows details
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Success      200  {object}   JSONResult "desc"
// @Failure      400  {object}   JSONResult "desc"
// @Failure      500  {object}   JSONResult "desc"
// @Router       /profile [get]
// @Security  Bearer
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
