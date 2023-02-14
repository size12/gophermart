package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/size12/gophermart/internal/config"
	"github.com/size12/gophermart/internal/entity"
	"golang.org/x/crypto/bcrypt"
)

type DBStorage struct {
	Cfg       config.Config
	DB        *sql.DB
	Queue     Queue
	StartTime time.Time
}

func NewDBStorage(cfg config.Config) (*DBStorage, error) {
	s := &DBStorage{Cfg: cfg, Queue: NewSliceQueue(), StartTime: time.Now()}

	DB, err := sql.Open("pgx", cfg.DataBaseURI)

	if err != nil {
		log.Fatalln("Failed open DB on startup: ", err)
		return s, err
	}

	err = MigrateUP(DB)
	if err != nil {
		log.Fatalln("Failed migrate DB: ", err)
		return s, err
	}

	s.DB = DB

	// каждые 10 секунд получаем необработанные ордера (которые были до запуска, добавляем их в очередь)
	go func() {
		orders, err := s.GetOrdersForUpdate(context.TODO())

		if err != nil {
			log.Println("Failed get orders for update")
			return
		}

		err = s.Queue.PushFrontOrders(orders...)
		if err != nil {
			log.Println("Failed push orders to queue")
			return
		}
		time.Sleep(10 * time.Second)
	}()

	return s, nil
}

func MigrateUP(DB *sql.DB) error {
	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		log.Printf("Failed create postgres instance: %v\n", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx", driver)
	if err != nil {
		log.Printf("Failed create migration instance: %v\n", err)
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("Failed migrate: ", err)
		return err
	}

	return nil
}

func (s *DBStorage) GetUser(ctx context.Context, search SearchType, value string) (entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.Cfg.AwaitTime)
	defer cancel()

	user := entity.User{}

	var row *sql.Row

	switch search {
	case SearchByID:
		query := `SELECT * FROM users WHERE id = $1`
		row = s.DB.QueryRowContext(ctx, query, value)
	case SearchByLogin:
		query := `SELECT * FROM users WHERE login = $1`
		row = s.DB.QueryRowContext(ctx, query, value)
	default:
		log.Fatalln("Failed search user by type")
		return user, errors.New("received wrong search type")
	}

	switch err := row.Scan(&user.ID, &user.Login, &user.Password, &user.Balance, &user.Withdrawn); err {
	case sql.ErrNoRows:
		return user, ErrNotFound
	case nil:
		return user, nil
	default:
		log.Println("Failed get user:", err)
		return user, err
	}
}

func (s *DBStorage) AddUser(ctx context.Context, user entity.User) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, s.Cfg.AwaitTime)
	defer cancel()

	_, err := s.GetUser(ctx, "login", user.Login)
	if err == nil {
		return 0, ErrLoginExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Login+user.Password), 0)
	if err != nil {
		log.Println("Failed generate hash from password:", err)
		return 0, err
	}

	_, err = s.DB.ExecContext(ctx, `INSERT INTO users (login, passw, balance, withdrawn) VALUES ($1, $2, $3, $4)`, user.Login, hash, 0, 0)

	if err != nil {
		log.Println("Failed added new user to DB:", err)
		return 0, err
	}

	err = s.DB.QueryRowContext(ctx, `SELECT id FROM users WHERE login = $1`, user.Login).Scan(&user.ID)

	if err != nil {
		log.Println("Failed get user ID:", err)
		return 0, err
	}

	return user.ID, nil
}

func (s *DBStorage) Withdraw(ctx context.Context, user entity.User, withdrawal entity.Withdraw) error {
	ctx, cancel := context.WithTimeout(ctx, s.Cfg.AwaitTime)
	defer cancel()

	result, err := s.DB.ExecContext(ctx, `UPDATE users SET balance = balance - $1, withdrawn = withdrawn + $1 WHERE id = $2 AND balance >= $1`, withdrawal.Sum, user.ID)

	if err != nil {
		log.Println("Failed withdraw:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Failed get affected rows")
		return err
	}

	if rowsAffected == 0 {
		return ErrNoMoney
	}

	_, err = s.DB.ExecContext(ctx, `INSERT INTO withdrawals (userid, num, amount, processed) VALUES ($1, $2, $3, $4)`, user.ID, withdrawal.Order, withdrawal.Sum, time.Now())

	if err != nil {
		log.Println("Failed insert withdrawal into withdrawals:", err)
		return err
	}

	return nil
}

func (s *DBStorage) WithdrawalHistory(ctx context.Context, user entity.User) ([]entity.Withdraw, error) {
	var withdrawals []entity.Withdraw

	ctx, cancel := context.WithTimeout(ctx, s.Cfg.AwaitTime)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, "SELECT num, amount, processed FROM withdrawals WHERE userid = $1 ORDER BY processed DESC ", user.ID)

	if err != nil {
		log.Println("Can't get withdrawals history from DB:", err)
		return withdrawals, err
	}

	for rows.Next() {
		withdrawal := entity.Withdraw{}
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.Time)
		if err != nil {
			log.Println("Error while scanning rows:", err)
			return withdrawals, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		log.Println("Rows error:", err)
		return withdrawals, err
	}

	return withdrawals, nil
}

func (s *DBStorage) AddOrder(ctx context.Context, order entity.Order) error {
	ctx, cancel := context.WithTimeout(ctx, s.Cfg.AwaitTime)
	defer cancel()

	row := s.DB.QueryRowContext(ctx, `SELECT userid FROM orders WHERE num = $1 LIMIT 1`, order.Number)

	orderDB := entity.Order{}

	err := row.Scan(&orderDB.UserID)

	if err == nil {
		if orderDB.UserID == order.UserID {
			return ErrAlreadyLoaded
		}
		return ErrLoadedByOtherUser
	}

	_, err = s.DB.ExecContext(ctx, `INSERT INTO orders (userid, num, stat, accrual, uploaded) VALUES ($1, $2, $3, $4, $5)`, order.UserID, order.Number, "NEW", 0, time.Now())

	if err != nil {
		log.Println("Failed insert new order into orders:", err)
		return err
	}

	err = s.Queue.PushBackOrders(order)
	if err != nil {
		log.Println("Failed push order to queue")
		return err
	}

	return nil
}

func (s *DBStorage) OrdersHistory(ctx context.Context, user entity.User) ([]entity.Order, error) {
	var orders []entity.Order

	ctx, cancel := context.WithTimeout(ctx, s.Cfg.AwaitTime)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, "SELECT num, stat, accrual, uploaded FROM orders WHERE userid = $1 ORDER BY uploaded DESC ", user.ID)

	if err != nil {
		log.Println("Can't get withdrawals history from DB:", err)
		return orders, err
	}

	for rows.Next() {
		order := entity.Order{}
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.EventTime)
		if err != nil {
			log.Println("Error while scanning rows:", err)
			return orders, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		log.Println("Rows error:", err)
		return orders, err
	}

	return orders, nil
}

func (s *DBStorage) GetOrderForUpdate() (entity.Order, error) {
	return s.Queue.GetOrder()
}

func (s *DBStorage) GetOrdersForUpdate(ctx context.Context) ([]entity.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, s.Cfg.AwaitTime)
	defer cancel()

	var orders []entity.Order

	rows, err := s.DB.QueryContext(ctx, `SELECT userid, num, stat FROM orders WHERE (stat = 'NEW' OR stat = 'PROCESSING') AND  uploaded < $1 ORDER BY uploaded ASC LIMIT 10`, s.StartTime)

	if err != nil {
		log.Println("can't get orders from DB for update:", err)
		return orders, err
	}

	for rows.Next() {
		order := entity.Order{}
		err = rows.Scan(&order.UserID, &order.Number, &order.Status)
		if err != nil {
			log.Println("Rows error:", err)
			return orders, err
		}
		orders = append(orders, order)
	}

	err = rows.Err()
	if errors.Is(err, sql.ErrNoRows) || err == nil {
		return orders, nil
	}

	return orders, err
}

func (s *DBStorage) UpdateOrders(ctx context.Context, orders ...entity.Order) error {
	ctx, cancel := context.WithTimeout(ctx, s.Cfg.AwaitTime)
	defer cancel()

	tx, err := s.DB.Begin()
	defer tx.Rollback()

	if err != nil {
		return err
	}

	stmtOrders, err := tx.PrepareContext(ctx, `UPDATE orders SET stat = $1, accrual = $2 WHERE num = $3`)
	if err != nil {
		return err
	}
	defer stmtOrders.Close()

	stmtUsers, err := tx.PrepareContext(ctx, `UPDATE users SET balance = balance + $1 WHERE id = $2`)
	if err != nil {
		return err
	}
	defer stmtUsers.Close()

	for _, order := range orders {
		if _, err := stmtOrders.ExecContext(ctx, order.Status, order.Accrual, order.Number); err != nil {
			return err
		}

		if order.Accrual > 0 {
			if _, err := stmtUsers.ExecContext(ctx, order.Accrual, order.UserID); err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) GetConfig() config.Config {
	return s.Cfg
}

func (s *DBStorage) PushFrontOrders(orders ...entity.Order) error {
	return s.Queue.PushFrontOrders(orders...)
}

func (s *DBStorage) PushBackOrders(orders ...entity.Order) error {
	return s.Queue.PushBackOrders(orders...)
}
