package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/erikdubbelboer/gspt"

	"micro-loan/common/models"
	"micro-loan/common/tools"
)

type workArgsT struct {
	Model     string
	Database  string
	Table     string
	Chunk     bool
	Input     string
	Output    string
	SkipField string
	Help      bool
}

const programName = "export-tool"

var workArgs workArgsT

func init() {
	flag.StringVar(&workArgs.Model, "model", "schema", "set export model, support:schema,data")
	flag.StringVar(&workArgs.Database, "database", "", "database")
	flag.StringVar(&workArgs.Table, "table", "", "databases tables")
	flag.BoolVar(&workArgs.Chunk, "chunk", false, "export all data use chunk")
	flag.StringVar(&workArgs.Input, "input", "", "export query sql filename")
	flag.StringVar(&workArgs.Output, "output", "", "output file")
	flag.StringVar(&workArgs.SkipField, "skip-field", "", "set skip field when create INSERT sql")
	flag.BoolVar(&workArgs.Help, "h", false, "show usage and exit")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stdout, programName+`
Usage:
  ./%s -h
  ./%s --database=[api|admin] --table=t1,t2...|all [--output=./output]
  ./%s --model=data --database=[api|admin] --table=tb --chunk=true|false --input=./input.sql [--skip-field=f1,f2...] [--output=./output.sql]
`, programName, programName, programName)

	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	flag.Parse()

	if workArgs.Help {
		flag.Usage()
	}

	if len(workArgs.Database) == 0 {
		flag.Usage()
	}

	if workArgs.Model != "schema" && workArgs.Model != "data" {
		fmt.Fprintf(os.Stdout, "no support model: %s\n", workArgs.Model)
		os.Exit(10)
	}

	if workArgs.Database != "api" && workArgs.Database != "admin" {
		fmt.Fprintf(os.Stdout, "no support db: %s\n", workArgs.Database)
		os.Exit(11)
	}

	if workArgs.Model == "schema" && len(workArgs.Table) == 0 {
		fmt.Fprintf(os.Stdout, "export schema, but no table assign.\n")
		os.Exit(12)
	}

	if workArgs.Model == "data" {
		if workArgs.Chunk == false && len(workArgs.Input) == 0 {
			fmt.Fprintf(os.Stdout, "export data, but no sql file assign.\n")
			os.Exit(13)
		}
	}

	if len(workArgs.Table) <= 0 {
		fmt.Fprintf(os.Stdout, "please assign table name.\n")
		os.Exit(14)
	}

	gspt.SetProcTitle(programName)

	doWork(workArgs)
}

func doWork(workArgs workArgsT) {
	var output = os.Stdout
	if len(workArgs.Output) > 0 {
		f, err := os.Create(workArgs.Output)
		if err != nil {
			logs.Error("[doWork] can open file: %s, err: %s", workArgs.Output, err.Error())
			os.Exit(20)
		}
		defer f.Close()

		output = f
	}

	timeNow := time.Now()
	comment := fmt.Sprintf("/* export %s by %s at: %d-%02d-%02d %02d:%02d:%02d */\n\n", workArgs.Model, programName,
		timeNow.Year(), int(timeNow.Month()), timeNow.Day(),
		timeNow.Hour(), timeNow.Minute(), timeNow.Second())
	output.WriteString(comment)

	o := orm.NewOrm()
	api := models.Order{}
	admin := models.Admin{}
	if workArgs.Database == "admin" {
		o.Using(admin.UsingSlave())
	} else {
		o.Using(api.UsingSlave())
	}

	if workArgs.Model == "schema" {
		doWorkExportSchem(workArgs, o, output)
	} else {
		doWorkExportData(workArgs, o, output)
	}
}

func doWorkExportSchem(workArgs workArgsT, o orm.Ormer, output *os.File) {
	logs.Informational("[doWorkExportSchem] start work")

	var tables []string

	if workArgs.Table == "all" {
		sql := "SHOW TABLES"
		logs.Debug("[doWorkExportSchem] sql: %s", sql)

		var dbResult []orm.Params
		num, errSql := o.Raw(sql).Values(&dbResult)
		if errSql != nil || num <= 0 {
			logs.Error("[doWorkExportSchem] sql has wrong, sql:", sql, ", num:", num, ", errSql:", errSql)
			return
		}

		for _, dbResultKV := range dbResult {
			//logs.Debug("[doWorkExportSchem] dbResultKV:", dbResultKV)
			for _, tblName := range dbResultKV {
				tables = append(tables, tblName.(string))
			}
		}
	} else {
		tables = strings.Split(workArgs.Table, ",")
	}
	//logs.Debug("[doWorkExportSchem] tables: %#v\n", tables)

	for _, tbl := range tables {
		sql := fmt.Sprintf("SHOW CREATE TABLE `%s`", tbl)
		logs.Debug("[doWorkExportSchem] sql: %s", sql)

		var dbResult []orm.Params
		num, errSql := o.Raw(sql).Values(&dbResult)
		if errSql != nil || num <= 0 {
			logs.Error("[doWorkExportSchem] sql has wrong, sql:", sql, ", num:", num, ", errSql:", errSql)
			continue
		}

		addIf := fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", tbl)
		output.WriteString(addIf)

		dbResultKV := dbResult[0]
		//fmt.Printf("dbResultKV->CT: %s\n", dbResultKV["Create Table"])
		createSQL := dbResultKV["Create Table"].(string) + ";\n"
		re := regexp.MustCompile(`AUTO_INCREMENT=(\d+) `)
		createSQL = re.ReplaceAllString(createSQL, "")

		output.WriteString(createSQL)
		output.WriteString("\n")
	}

	logs.Informational("[doWorkExportSchem] jobs have done.")
}

func doWorkExportData(workArgs workArgsT, o orm.Ormer, output *os.File) {
	logs.Informational("[doWorkExportData] start work")

	if workArgs.Chunk {
		logs.Informational("[doWorkExportData] use chunk")
		const chunkSize int64 = 1000

		var total int64
		totalSQL := fmt.Sprintf(`SELECT COUNT(*) AS total FROM %s`, workArgs.Table)
		r := o.Raw(totalSQL)
		r.QueryRow(&total)

		var pageTotal int64 = int64(math.Ceil(float64(total) / float64(chunkSize)))
		logs.Debug("[doWorkExportData] pageTotal: %d", pageTotal)

		for i := int64(0); i < pageTotal; i++ {
			offset := i * chunkSize
			sql := fmt.Sprintf(`SELECT * FROM %s LIMIT %d OFFSET %d`, workArgs.Table, chunkSize, offset)
			logs.Debug("[doWorkExportData] sql: %s", sql)
			output.WriteString(fmt.Sprintf("/** chunk: %d */\n", i))
			doWorkExportDataUseChunk(workArgs, o, output, sql)
		}
	} else {
		sqlBytes, err := ioutil.ReadFile(workArgs.Input)
		if err != nil {
			fmt.Fprintf(os.Stdout, "cat not read sql file:%s, err: %s\n", workArgs.Input, err.Error())
			os.Exit(30)
		}

		sql := string(sqlBytes)
		doWorkExportDataUseChunk(workArgs, o, output, sql)
	}

	logs.Informational("[doWorkExportData] jobs have done.")
}

func doWorkExportDataUseChunk(workArgs workArgsT, o orm.Ormer, output *os.File, sql string) {
	logs.Informational("[doWorkExportDataUseChunk] chunk jobs start.")
	logs.Debug("sql:", sql)

	var dbResult []orm.Params
	num, errSql := o.Raw(sql).Values(&dbResult)
	if errSql != nil {
		log := fmt.Sprintf("[doWorkExportDataUseChunk] sql has wrong, sql: %s, num: %d, errSql: %s", sql, num, errSql.Error())
		fmt.Fprintf(os.Stdout, log)
		logs.Error(log)
		os.Exit(31)
	}

	var fieldBox []string
	var skipFieldBox = make(map[string]bool)
	expSkipField := strings.Split(workArgs.SkipField, ",")
	if len(expSkipField) > 0 {
		for _, field := range expSkipField {
			skipFieldBox[field] = true
		}
	}

	for i := int64(0); i < num; i++ {
		dbResultKV := dbResult[i]
		//fmt.Println(dbResultKV)
		if i == 0 {
			for k, _ := range dbResultKV {
				if skipFieldBox[k] {
					continue
				}
				fieldBox = append(fieldBox, k)
			}

			initSql := fmt.Sprintf("INSERT INTO `%s` (`%s`) VALUES\n", workArgs.Table, strings.Join(fieldBox, "`, `"))
			output.WriteString(initSql)
		}

		//fmt.Println("fieldBox:", fieldBox)
		//fmt.Println("skipFieldBox:", skipFieldBox)

		var values []string
		for _, field := range fieldBox {
			if dbResultKV[field] != nil {
				values = append(values, fmt.Sprintf(`'%s'`, tools.AddSlashes(dbResultKV[field].(string))))
			} else {
				values = append(values, `''`)
			}
		}
		vSql := fmt.Sprintf("(%s)", strings.Join(values, ", "))
		if i < num-1 {
			vSql += ",\n"
		}
		output.WriteString(vSql)
	}

	output.WriteString(";\n\n")

	logs.Informational("[doWorkExportDataUseChunk] chunk jobs have done.")
}
