// 定制 mysql

package cmysql

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"

	"micro-loan/common/tools"
	"micro-loan/common/types"
	"time"
)

func init() {
	if !tools.IsProductEnv() {
		orm.Debug = true
	}

	dbType := beego.AppConfig.String("db_type")
	dbCharset := beego.AppConfig.String("db_charset")

	orm.RegisterDriver(dbType, orm.DRMySQL)

	// 注册`admin`
	dbHost := beego.AppConfig.String("db_admin_host")
	dbPort := beego.AppConfig.String("db_admin_port")
	dbName := beego.AppConfig.String("db_admin_name")
	dbUser := beego.AppConfig.String("db_admin_user")
	dbPwd := beego.AppConfig.String("db_admin_pwd")

	//fmt.Printf("dbHost: %s, dbPort: %s, dbName: %s, dbPwd:%s\n", dbHost, dbPort, dbName, dbPwd)
	fmt.Printf("types.OrmDataBaseAdmin: %s\n", types.OrmDataBaseAdmin)
	orm.RegisterDataBase(types.OrmDataBaseAdmin, dbType, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	db, _ := orm.GetDB(types.OrmDataBaseAdmin)
	db.SetConnMaxLifetime(time.Hour)

	// 注册`api`
	dbHost = beego.AppConfig.String("db_api_host")
	dbPort = beego.AppConfig.String("db_api_port")
	dbName = beego.AppConfig.String("db_api_name")
	dbUser = beego.AppConfig.String("db_api_user")
	dbPwd = beego.AppConfig.String("db_api_pwd")

	fmt.Printf("types.OrmDataBaseApi: %s\n", types.OrmDataBaseApi)
	orm.RegisterDataBase(types.OrmDataBaseApi, dbType, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	db, _ = orm.GetDB(types.OrmDataBaseApi)
	db.SetConnMaxLifetime(time.Hour)

	// 注册`adminSlave`
	dbHost = beego.AppConfig.String("db_admin_slave_host")
	dbPort = beego.AppConfig.String("db_admin_slave_port")
	dbName = beego.AppConfig.String("db_admin_slave_name")
	dbUser = beego.AppConfig.String("db_admin_slave_user")
	dbPwd = beego.AppConfig.String("db_admin_slave_pwd")

	fmt.Printf("types.OrmDataBaseAdminSlave: %s\n", types.OrmDataBaseAdminSlave)
	//logs.Debug(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	orm.RegisterDataBase(types.OrmDataBaseAdminSlave, dbType, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	db, _ = orm.GetDB(types.OrmDataBaseAdminSlave)
	db.SetConnMaxLifetime(time.Hour)

	// 注册`apiSlave`
	dbHost = beego.AppConfig.String("db_api_slave_host")
	dbPort = beego.AppConfig.String("db_api_slave_port")
	dbName = beego.AppConfig.String("db_api_slave_name")
	dbUser = beego.AppConfig.String("db_api_slave_user")
	dbPwd = beego.AppConfig.String("db_api_slave_pwd")

	fmt.Printf("types.OrmDataBaseApiSlave: %s\n", types.OrmDataBaseApiSlave)
	//logs.Debug(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	orm.RegisterDataBase(types.OrmDataBaseApiSlave, dbType, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	db, _ = orm.GetDB(types.OrmDataBaseApiSlave)
	db.SetConnMaxLifetime(time.Hour)

	//// 注册`riskMonitor`
	dbHost = beego.AppConfig.String("db_risk_monitor_host")
	dbPort = beego.AppConfig.String("db_risk_monitor_port")
	dbName = beego.AppConfig.String("db_risk_monitor_name")
	dbUser = beego.AppConfig.String("db_risk_monitor_user")
	dbPwd = beego.AppConfig.String("db_risk_monitor_pwd")

	fmt.Printf("types.OrmDataBaseRiskMonitor: %s\n", types.OrmDataBaseRiskMonitor)
	////logs.Debug(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	orm.RegisterDataBase(types.OrmDataBaseRiskMonitor, dbType, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	db, _ = orm.GetDB(types.OrmDataBaseRiskMonitor)
	db.SetConnMaxLifetime(time.Hour)

	//// 注册`riskMonitorSlave`
	//dbHost = beego.AppConfig.String("db_risk_monitor_slave_host")
	//dbPort = beego.AppConfig.String("db_risk_monitor_slave_port")
	//dbName = beego.AppConfig.String("db_risk_monitor_slave_name")
	//dbUser = beego.AppConfig.String("db_risk_monitor_slave_user")
	//dbPwd = beego.AppConfig.String("db_risk_monitor_slave_pwd")

	//fmt.Printf("types.OrmDataBaseRiskMonitorSlave: %s\n", types.OrmDataBaseRiskMonitorSlave)
	////logs.Debug(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	//orm.RegisterDataBase(types.OrmDataBaseRiskMonitorSlave, dbType, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	//db, _ = orm.GetDB(types.OrmDataBaseRiskMonitorSlave)
	//db.SetConnMaxLifetime(time.Hour)

	//// 注册`riskMonitor`
	dbHost = beego.AppConfig.String("db_message_host")
	dbPort = beego.AppConfig.String("db_message_port")
	dbName = beego.AppConfig.String("db_message_name")
	dbUser = beego.AppConfig.String("db_message_user")
	dbPwd = beego.AppConfig.String("db_message_pwd")

	fmt.Printf("types.OrmDataBaseMessage: %s\n", types.OrmDataBaseMessage)
	////logs.Debug(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	orm.RegisterDataBase(types.OrmDataBaseMessage, dbType, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbUser, dbPwd, dbHost, dbPort, dbName, dbCharset))
	db, _ = orm.GetDB(types.OrmDataBaseMessage)
	db.SetConnMaxLifetime(time.Hour)
}
