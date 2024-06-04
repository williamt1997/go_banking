package main

import (
	"encoding/json"
	"fmt"
	"go_banking/cmd/bank_app/initializers"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type CreateAccount struct {
	AccountName     string `json:"account_name" gorm:"not null"`
	AccountEmail    string `json:"account_email" gorm:"not null"`
	AccountPassword string `json:"account_password" gorm:"not null"`
}

type LoginAccount struct {
	AccountEmail    string `json:"account_email" gorm:"not null"`
	AccountPassword string `json:"account_password" gorm:"not null"`
}

type CardCreate struct {
	CardBalance string `json:"card_balance" gorm:"not null"`
}

type GetCard struct {
	Card_code    int     `json:"card_code"`
	Account_code int     `json:"account_code"`
	Card_balance float32 `json:"card_balance"`
}

type GetCardCode struct {
	Card_code_id int `json:"card_code_id" gorm:"not null"`
}

type MakeTransaction struct {
	Recipient_code     string `json:"recipient_code" gorm:"not null"`
	Transaction_amount string `json:"transaction_amount" gorm:"not null"`
}

var Store = cookie.NewStore([]byte("b67daea11cddfb826c4c7a1ec3a8ec824da526223653c0e748f3cbb4a683a71b"))

func init() {
	initializers.Get_envs()
	initializers.Create_database()
	var result initializers.Account
	initializers.PostgresDB.First(&result)
	fmt.Println(result)
}
func main() {
	r := gin.Default()
	r.Use(sessions.Sessions("mysession", Store))
	r.LoadHTMLGlob("./templates/*")
	r.Static("/Styles", "./Styles/")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.tmpl", map[string]string{"title": "home_page"})
	})

	r.GET("/HomeRedir", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/")
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(200, "register.tmpl", map[string]string{"title": "register_page"})
	})

	r.GET("/RegisterRedir", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/register")
	})

	r.POST("/register", func(c *gin.Context) {
		var newAccount CreateAccount
		c.ShouldBindJSON(&newAccount)

		initializers.Create_database()
		var query initializers.Account

		result := initializers.PostgresDB.Where("account_email = ?", newAccount.AccountEmail).Find(&query)
		if result.RowsAffected < 1 {
			hash, err := bcrypt.GenerateFromPassword([]byte(newAccount.AccountPassword), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(500, gin.H{"message": "Error Creating Hashed Password", "is_ok": "no"})
			} else {
				accountCreate := initializers.Account{AccountName: newAccount.AccountName, AccountEmail: newAccount.AccountEmail, AccountPassword: string(hash)}
				err := initializers.PostgresDB.Create(&accountCreate).Error
				if err != nil {
					c.JSON(500, gin.H{"message": "Error Inserting New Account Into Database", "is_ok": "no"})
				} else {
					c.JSON(201, gin.H{"message": "Account Created: Please Log In", "is_ok": "yes"})
				}
			}
		} else {
			c.JSON(409, gin.H{"message": "Email Already Exists", "is_ok": "no"})
		}
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.tmpl", map[string]string{"title": "login_page"})
	})

	r.GET("/LoginRedir", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	r.POST("/login", func(c *gin.Context) {
		var loginAccount LoginAccount
		c.ShouldBindJSON(&loginAccount)

		initializers.Create_database()
		var query initializers.Account

		err := initializers.PostgresDB.Where("account_email = ?", loginAccount.AccountEmail).First(&query).Error
		if err != nil {
			c.JSON(401, gin.H{"message": "incorrect email or password", "is_ok": "no"})
		} else {
			if err := bcrypt.CompareHashAndPassword([]byte(query.AccountPassword), []byte(loginAccount.AccountPassword)); err != nil {
				c.JSON(401, gin.H{"message": "incorrect email or password", "is_ok": "no"})
			} else {
				println("Correct")
				session := sessions.Default(c)
				session.Set("account_code", query.AccountCode)
				session.Set("account_name", query.AccountName)
				session.Save()
				c.JSON(202, gin.H{"message": "Logged In Successfully", "is_ok": "yes"})
			}
		}
	})

	r.GET("/account", func(c *gin.Context) {
		authentication(c)
		c.HTML(200, "account.tmpl", map[string]string{"title": "register_page"})
	})

	r.GET("/AccountRedir", func(c *gin.Context) {
		authentication(c)
		c.Redirect(http.StatusFound, "/account")
	})

	r.GET("/AccountName", func(c *gin.Context) {
		authentication(c)
		session := sessions.Default(c)
		var my_seshname = session.Get("account_name")

		c.JSON(202, gin.H{"message": my_seshname})
	})

	r.GET("/createcard", func(c *gin.Context) {
		authentication(c)
		c.HTML(200, "createcard.tmpl", map[string]string{"title": "register_page"})
	})

	r.GET("/CreateCardRedir", func(c *gin.Context) {
		authentication(c)
		c.Redirect(http.StatusFound, "/createcard")
	})

	r.POST("/createcard", func(c *gin.Context) {
		authentication(c)
		var cardCreate CardCreate
		c.ShouldBindJSON(&cardCreate)
		cardBalanc, err := strconv.ParseFloat(cardCreate.CardBalance, 32)
		if err != nil {
			c.JSON(500, gin.H{"message": "Error Converting To Float", "is_ok": "no"})
		} else {
			card_bal := float32(cardBalanc)

			session := sessions.Default(c)
			var my_seshcode = session.Get("account_code").(int)

			createCard := initializers.Card{AccountCode: my_seshcode, CardBalance: card_bal}

			err := initializers.PostgresDB.Create(&createCard).Error
			if err != nil {
				c.JSON(500, gin.H{"message": "Error Inserting New Card Into Database", "is_ok": "no"})
			} else {
				c.JSON(201, gin.H{"message": "Card Created", "is_ok": "yes"})
			}
		}
	})

	r.GET("/getcard", func(c *gin.Context) {
		authentication(c)

		session := sessions.Default(c)
		var my_seshcode = session.Get("account_code").(int)

		var getCard []GetCard

		err := initializers.PostgresDB.Raw("SELECT * FROM cards WHERE account_code = ?", my_seshcode).Scan(&getCard).Error
		if err != nil {
			c.JSON(500, gin.H{"message": "Error Querying Database"})
		} else {
			test, err := json.Marshal(getCard)
			if err != nil {
				fmt.Println("no")
			} else {
				c.JSON(200, gin.H{"message": string(test)})
			}
		}
	})

	r.PUT("/GetCardID", func(c *gin.Context) {
		authentication(c)

		var getCardcode GetCardCode
		c.ShouldBindJSON(&getCardcode)

		session := sessions.Default(c)
		session.Set("card_code", getCardcode.Card_code_id)
		err := session.Save()
		if err != nil {
			c.JSON(406, gin.H{"message": "Error Adding Card Code To Session", "is_ok": "no"})
		} else {
			c.JSON(200, gin.H{"message": "Session Appended With Card Code", "is_ok": "yes"})
		}
	})

	r.GET("/transaction", func(c *gin.Context) {
		authentication(c)
		c.HTML(200, "transaction.tmpl", map[string]string{"title": "register_page"})
	})

	r.GET("/MakeTransactionRedir", func(c *gin.Context) {
		authentication(c)
		c.Redirect(http.StatusFound, "/transaction")
	})

	r.GET("/maketransaction", func(c *gin.Context) {
		authentication(c)
		c.HTML(200, "maketransaction.tmpl", map[string]string{"title": "register_page"})
	})

	r.POST("/create_transaction", func(c *gin.Context) {
		authentication(c)
		session := sessions.Default(c)
		var my_seshcode = session.Get("card_code").(int)

		var makeTransaction MakeTransaction
		c.ShouldBindJSON(&makeTransaction)

		recipient_code, err := strconv.Atoi(makeTransaction.Recipient_code)
		if err != nil {
			c.JSON(500, gin.H{"message": "Error Converting Recipient_code To Int", "is_ok": "no"})
		} else {
			initializers.Create_database()
			var queryone initializers.Card
			var querytwo initializers.Card

			if my_seshcode == recipient_code {
				c.JSON(418, gin.H{"message": "You Cannot Transfer Finances To The Same Card", "is_ok": "no"})
			} else {
				transamount, err := strconv.ParseFloat(makeTransaction.Transaction_amount, 32)
				if err != nil {
					c.JSON(500, gin.H{"message": "Error Converting Transaction To Float", "is_ok": "no"})
				} else {
					err := initializers.PostgresDB.Where("card_code = ?", my_seshcode).Find(&queryone).Error
					if err != nil {
						c.JSON(500, gin.H{"message": "Error Querying Database", "is_ok": "no"})
					} else {
						sender_amount := float32(transamount)
						if queryone.CardBalance < sender_amount {
							c.JSON(418, gin.H{"message": "Your Transaction Amount Cannot Be More Than Your Cards Balance", "is_ok": "no"})
						} else {
							result := initializers.PostgresDB.Where("card_code = ?", makeTransaction.Recipient_code).Find(&querytwo)
							if result.RowsAffected < 1 {
								c.JSON(404, gin.H{"message": "Recipient Card Code Not Found", "is_ok": "no"})
							} else {
								var sender_new_amount = queryone.CardBalance - sender_amount
								var recipient_new_amount = querytwo.CardBalance + sender_amount

								fmt.Println(queryone.CardBalance, " - ", sender_amount, "=", sender_new_amount)
								fmt.Println(querytwo.CardBalance, " + ", sender_amount, "=", recipient_new_amount)

								var addamount []initializers.Card

								err := initializers.PostgresDB.Raw("UPDATE cards SET card_balance = ? WHERE card_code = ?", sender_new_amount, my_seshcode).Scan(&addamount).Error
								if err != nil {
									c.JSON(500, gin.H{"message": "Error updating Sender Balance To The Cards Table", "is_ok": "no"})
								} else {
									err := initializers.PostgresDB.Raw("UPDATE cards SET card_balance = ? WHERE card_code = ?", recipient_new_amount, makeTransaction.Recipient_code).Scan(&addamount).Error
									if err != nil {
										c.JSON(500, gin.H{"message": "Error updating Recipient Balance To The Cards Table", "is_ok": "no"})
									} else {
										transactionCreate := initializers.Transaction{SenderCode: my_seshcode, RecipientCode: recipient_code, TransactionAmount: sender_amount}

										err := initializers.PostgresDB.Create(&transactionCreate).Error
										if err != nil {
											c.JSON(500, gin.H{"message": "Error Inserting New Transaction Into Database", "is_ok": "no"})
										} else {
											c.JSON(201, gin.H{"message": "Transaction Made", "is_ok": "yes"})
										}
									}
								}
							}
						}
					}
				}
			}
		}
	})

	r.GET("/record", func(c *gin.Context) {
		authentication(c)
		c.HTML(200, "record.tmpl", map[string]string{"title": "register_page"})
	})

	r.GET("/RecordsRedir", func(c *gin.Context) {
		authentication(c)
		c.Redirect(http.StatusFound, "/record")
	})

	r.GET("/viewrecord", func(c *gin.Context) {
		authentication(c)
		c.HTML(200, "viewrecord.tmpl", map[string]string{"title": "register_page"})
	})

	r.GET("/transaction_history", func(c *gin.Context) {
		authentication(c)

		session := sessions.Default(c)
		var my_seshcode = session.Get("card_code").(int)

		var transactionGetSender []initializers.Transaction
		var transactionGetRecipient []initializers.Transaction

		err := initializers.PostgresDB.Raw("SELECT * FROM transactions WHERE sender_code = ?", my_seshcode).Scan(&transactionGetSender).Error
		if err != nil {
			c.JSON(500, gin.H{"message": "Error Querying Database"})
		} else {
			out_json, err := json.Marshal(transactionGetSender)
			if err != nil {
				c.JSON(500, gin.H{"message": "Error Converting Queried Data To JSON"})
			} else {
				err := initializers.PostgresDB.Raw("SELECT * FROM transactions WHERE recipient_code = ?", my_seshcode).Scan(&transactionGetRecipient).Error
				if err != nil {
					c.JSON(500, gin.H{"message": "Error Querying Database"})
				} else {
					in_json, err := json.Marshal(transactionGetRecipient)
					if err != nil {
						c.JSON(500, gin.H{"message": "Error Converting Queried Data To JSON"})
					} else {
						fmt.Println(string(out_json))
						c.JSON(200, gin.H{"out": string(out_json), "in": string(in_json)})
					}
				}
			}
		}
	})

	r.POST("/logout", func(c *gin.Context) {
		authentication(c)
		session := sessions.Default(c)
		session.Clear()
		session.Save()

		c.Redirect(http.StatusFound, "/")
	})

	log.Fatal(r.Run())

}

func authentication(c *gin.Context) {
	session := sessions.Default(c)
	var test = session.Get("account_code")
	if test == nil {
		c.Redirect(http.StatusFound, "/")
	}
}
