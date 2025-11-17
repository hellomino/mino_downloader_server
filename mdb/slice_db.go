package mdb

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"minodl/log"
)

const (
	PartOnlyOne = 1
	EMPTY       = ""
	SpiltTag    = "_"
	DateFormat  = "200601"
)

var (
	migrated    = new(sync.Map)
	migrateLock sync.Mutex
	mainDB      *gorm.DB
	mainOne     sync.Once
)

type SP struct {
	Value    interface{}
	ByID     int64  // 关键ID
	ByDate   string // 年月日
	Part     int64  // 多个分表
	Absolute string // 自定义绝对分表
}

func isValidDate(date string) bool {
	inputTime, err := time.Parse("200601", date)
	if err != nil {
		log.Error("error spart date：%s", date)
		return false
	}
	now := time.Now()
	currentYear, currentMonth := now.Year(), now.Month()
	inputYear, inputMonth := inputTime.Year(), inputTime.Month()
	// 计算总月份
	currentTotalMonths := currentYear*12 + int(currentMonth)
	inputTotalMonths := inputYear*12 + int(inputMonth)
	diff := currentTotalMonths - inputTotalMonths
	if inputYear < 2024 {
		return false
	} else if diff == -12 {
		//一年后
		return false
	} else {
		return true
	}
}

// Slice 获取自动分表
func Slice(odb *gorm.DB, sp SP) *gorm.DB {
	if sp.Value == nil {
		return nil
	}
	origin := reflect.TypeOf(sp.Value).String()
	var tail string
	if sp.Absolute == EMPTY {
		subNames := make([]string, 0, 10)
		if sp.ByID > 0 {
			subNames = append(subNames, strconv.FormatInt(sp.ByID, 10))
		}
		if sp.Part > PartOnlyOne {
			subNames = append(subNames, fmt.Sprintf("%d", sp.ByID%sp.Part))
		}
		if sp.ByDate != EMPTY {
			subNames = append(subNames, sp.ByDate)
			// 判断是否是有效日期
			if !isValidDate(sp.ByDate) {
				return nil
			}
		}
		if len(subNames) > 0 {
			tail = SpiltTag + strings.Join(subNames, SpiltTag)
		}
	} else {
		tail = SpiltTag + sp.Absolute
	}
	tableKey := origin + tail
	tableName, ok := migrated.Load(tableKey)
	// 创建全新会话
	if !ok {
		SetMainDB()
		migrateLock.Lock()
		defer migrateLock.Unlock()
		// 防止表并发迁移
		if tableName, ok = migrated.Load(tableKey); !ok {
			if tableName = autoMigrateTable(tail, sp.Value); tableName != nil {
				migrated.Store(tableKey, tableName)
			} else {
				// 迁移出错返回，原始DB
				return odb
			}
		}
	}
	log.Debug("slice table name:%v, key:%s", tableName, tableKey)
	return newSession(odb).Table(tableName.(string))
}

// autoMigrateTable 自动迁移表, 在使用之前完成
func autoMigrateTable(tail string, ett interface{}) interface{} {
	var err error
	var tdb *gorm.DB
	tdb = newSession(mainDB)
	if tdb.Statement.Schema, err = schema.Parse(ett, migrated, tdb.NamingStrategy); err == nil && tdb.Statement.Table == "" {
		if tables := strings.Split(tdb.Statement.Schema.Table, "."); len(tables) == 2 {
			tdb.Statement.TableExpr = &clause.Expr{SQL: tdb.Statement.Quote(tdb.Statement.Schema.Table)}
			tdb.Statement.Table = tables[1]
		}
		target := fmt.Sprintf("%s%s", tdb.Statement.Schema.Table, tail)
		log.Info("main db migrate slice table  %s", target)
		if err = tdb.Table(target).Migrator().AutoMigrate(ett); err != nil {
			log.Error("migrator error %v", err)
		}
		return target
	}
	return nil
}

func SetMainDB() {
	mainOne.Do(func() {
		migrateLock.Lock()
		defer migrateLock.Unlock()
		mainDB = Mysql
	})
}

func newSession(odb *gorm.DB) *gorm.DB {
	return odb.Session(&gorm.Session{
		NewDB:       true,
		SkipHooks:   true,
		Initialized: true,
	})
}
