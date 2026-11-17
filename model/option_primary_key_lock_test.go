package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestMigrateOptionPrimaryKeyDoesNotStallOnOneConnection(t *testing.T) {
	db, err := gorm.Open(mysql.Open("newapi:newapi@tcp(127.0.0.1:3306)/newapi_locktest?parseTime=true&charset=utf8mb4"), &gorm.Config{})
	if err != nil {
		t.Skip(err.Error())
	}
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.Exec("DROP TABLE IF EXISTS options").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS options_pk_tmp").Error)
	require.NoError(t, db.Exec("CREATE TABLE options (`key` varchar(191), value longtext)").Error)
	require.NoError(t, db.Exec("INSERT INTO options (`key`, value) VALUES ('site_name', 'old'), ('site_name', 'new')").Error)

	done := make(chan error, 1)
	go func() {
		done <- migrateOptionPrimaryKey(db)
	}()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(8 * time.Second):
		t.Fatal("options primary key migration blocked on a one-connection pool")
	}

	var value string
	require.NoError(t, db.Raw("SELECT value FROM options WHERE `key` = ?", "site_name").Scan(&value).Error)
	assert.Equal(t, "new", value)
}
