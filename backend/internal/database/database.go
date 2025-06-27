package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
	driverName string
}

func Initialize(databaseURL string) (*DB, error) {
	// 解析数据库URL
	driver, dsn := parseDatabaseURL(databaseURL)

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	wrappedDB := &DB{
		DB:         db,
		driverName: driver,
	}

	// 创建表
	if err := wrappedDB.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return wrappedDB, nil
}

func parseDatabaseURL(databaseURL string) (string, string) {
	if strings.HasPrefix(databaseURL, "mysql://") {
		// 转换 mysql://user:pass@host:port/db 为 user:pass@tcp(host:port)/db
		dsn := strings.TrimPrefix(databaseURL, "mysql://")
		// 查找 @ 符号的位置
		atIndex := strings.LastIndex(dsn, "@")
		if atIndex == -1 {
			return "mysql", dsn
		}

		userPass := dsn[:atIndex]
		hostDb := dsn[atIndex+1:]

		// 查找 / 符号的位置来分离主机和数据库
		slashIndex := strings.Index(hostDb, "/")
		if slashIndex == -1 {
			return "mysql", fmt.Sprintf("%s@tcp(%s)/", userPass, hostDb)
		}

		host := hostDb[:slashIndex]
		db := hostDb[slashIndex+1:]

		return "mysql", fmt.Sprintf("%s@tcp(%s)/%s", userPass, host, db)
	}
	if strings.HasPrefix(databaseURL, "sqlite3://") {
		return "sqlite3", strings.TrimPrefix(databaseURL, "sqlite3://")
	}
	// 默认使用 SQLite
	return "sqlite3", "./llm.db"
}

func (db *DB) createTables() error {
	// 使用存储的驱动名称
	driver := db.driverName

	var userTable, apiCallTable, tokenRechargeTable string

	if driver == "mysql" {
		// MySQL 建表语句
		userTable = `
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			tokens INT DEFAULT 1000,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
		`

		apiCallTable = `
		CREATE TABLE IF NOT EXISTS api_calls (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			endpoint VARCHAR(100) NOT NULL,
			tokens_consumed INT NOT NULL,
			request_data TEXT,
			response_data TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
		`

		tokenRechargeTable = `
		CREATE TABLE IF NOT EXISTS token_recharges (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			amount INT NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
		`
	} else {
		// SQLite 建表语句
		userTable = `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			tokens INTEGER DEFAULT 1000,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`

		apiCallTable = `
		CREATE TABLE IF NOT EXISTS api_calls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			endpoint VARCHAR(100) NOT NULL,
			tokens_consumed INTEGER NOT NULL,
			request_data TEXT,
			response_data TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		`

		tokenRechargeTable = `
		CREATE TABLE IF NOT EXISTS token_recharges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			amount INTEGER NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		`
	}

	tables := []string{userTable, apiCallTable, tokenRechargeTable}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return err
		}
	}

	return nil
}
