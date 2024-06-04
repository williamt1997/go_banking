package initializers

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Account struct {
	AccountCode     int    `gorm:"column:account_code;unique;primaryKey;autoIncrement"`
	AccountName     string `gorm:"column:account_name;size:25"`
	AccountEmail    string `gorm:"column:account_email;size:50"`
	AccountPassword string `gorm:"column:account_password"`
	Card            []Card `gorm:"foreignKey:AccountCode"`
}

type Card struct {
	CardCode             int           `gorm:"column:card_code;unique;primaryKey;autoIncrement"`
	AccountCode          int           `gorm:"column:account_code"`
	CardBalance          float32       `gorm:"column:card_balance;"`
	TransactionSender    []Transaction `gorm:"foreignKey:SenderCode"`
	TransactionRecipient []Transaction `gorm:"foreignKey:RecipientCode"`
}

type Transaction struct {
	TransactionCode      int       `gorm:"column:transaction_code;unique;primaryKey;autoIncrement"`
	SenderCode           int       `gorm:"column:sender_code"`
	RecipientCode        int       `gorm:"column:recipient_code"`
	TransactionAmount    float32   `gorm:"column:transaction_amount;type:numeric(18,2)"`
	TransactionTimestamp time.Time `gorm:"column:transaction_timestamp;default:CURRENT_TIMESTAMP"`
}

var (
	PostgresDB *gorm.DB
)

func Create_database() {
	var DSN = "host=localhost user=postgres password=thorpe01685 dbname=gobanking_db port=5433"

	var err error
	PostgresDB, err = gorm.Open(postgres.Open(DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	} else {
		err = PostgresDB.AutoMigrate(&Account{}, &Card{}, &Transaction{})
		if err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}
	}
}
