package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	mysqlrepo "referral-system/internal/repository/mysql"
	"referral-system/internal/service"
)

type App struct {
	DB                  *sql.DB
	UserRepo            *mysqlrepo.UserRepository
	ReferralRuleRepo    *mysqlrepo.ReferralRuleRepository
	ReferralRelationRepo *mysqlrepo.ReferralRelationRepository
	ReferralEventRepo   *mysqlrepo.ReferralEventRepository
	CreditAccountRepo   *mysqlrepo.CreditAccountRepository
	CreditLedgerRepo    *mysqlrepo.CreditLedgerRepository
	ReferralStatsRepo   *mysqlrepo.ReferralStatsDailyRepository
	ReferralService     *service.ReferralService
}

func New(ctx context.Context, cfg Config) (*App, error) {
	if cfg.MySQL.DSN == "" {
		return nil, fmt.Errorf("mysql dsn is required")
	}

	db, err := sql.Open("mysql", cfg.MySQL.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(50)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	userRepo := mysqlrepo.NewUserRepository(db)
	referralRuleRepo := mysqlrepo.NewReferralRuleRepository(db)
	referralRelationRepo := mysqlrepo.NewReferralRelationRepository(db)
	referralEventRepo := mysqlrepo.NewReferralEventRepository(db)
	creditAccountRepo := mysqlrepo.NewCreditAccountRepository(db)
	creditLedgerRepo := mysqlrepo.NewCreditLedgerRepository(db)
	referralStatsRepo := mysqlrepo.NewReferralStatsDailyRepository(db)
	txMgr := service.NewSQLTxManager(db)

	referralSvc := service.NewReferralService(service.ReferralServiceDeps{
		TxMgr:                txMgr,
		UserRepo:             userRepo,
		ReferralRuleRepo:     referralRuleRepo,
		ReferralRelationRepo:  referralRelationRepo,
		ReferralEventRepo:     referralEventRepo,
		CreditAccountRepo:     creditAccountRepo,
		CreditLedgerRepo:      creditLedgerRepo,
		StatsRepo:             referralStatsRepo,
	})

	return &App{
		DB:                  db,
		UserRepo:            userRepo,
		ReferralRuleRepo:    referralRuleRepo,
		ReferralRelationRepo: referralRelationRepo,
		ReferralEventRepo:   referralEventRepo,
		CreditAccountRepo:   creditAccountRepo,
		CreditLedgerRepo:    creditLedgerRepo,
		ReferralStatsRepo:   referralStatsRepo,
		ReferralService:     referralSvc,
	}, nil
}

func (a *App) Close() error {
	if a == nil || a.DB == nil {
		return nil
	}
	return a.DB.Close()
}

func (a *App) Handler() http.Handler {
	return NewHTTPHandler(a.ReferralService)
}
