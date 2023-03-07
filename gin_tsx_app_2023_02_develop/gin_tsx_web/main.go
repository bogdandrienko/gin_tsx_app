package main

//

// TODO imports ////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"reflect"

	//_ "github.com/godror/godror"
	//_ "github.com/lib/pq"
	//_ "github.com/sijms/go-ora"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/sijms/go-ora/v2"
	"golang.org/x/crypto/bcrypt"

	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TODO imports ////////////////////////////////////////////////////////////////////////////////////////////////////////

//

// TODO structs ////////////////////////////////////////////////////////////////////////////////////////////////////////

type ConfigS struct {
	// Gin
	DebugGin bool
	HostGin  string
	PortGin  int

	// Postgresql
	DriverNamePgSQL string
	HostPgSQL       string
	PortPgSQL       int
	UserPgSQL       string
	PasswordPgSQL   string
	DatabasePgSQL   string
	SslModePgSQL    string

	// Additional
	HashCost             int
	HashSalt             string
	TokenAccessLifetime  time.Duration
	TokenRefreshLifetime time.Duration
}

type RouteS struct {
	Method       string
	RelativePath string
	Function     gin.HandlerFunc
	Auth         bool
}

type User struct {
	// main
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	IsActive bool   `json:"is_active"`

	// roles
	IsModerator bool `json:"is_moderator"`
	IsAdmin     bool `json:"is_admin"`
	IsSuperuser bool `json:"is_superuser"`

	// additional
	Name              string    `json:"name"`
	Surname           string    `json:"surname"`
	Patronymic        string    `json:"patronymic"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone"`
	Avatar            string    `json:"avatar"`
	DatetimeJoined    time.Time `json:"datetime_joined"`
	DatetimeLastLogin time.Time `json:"datetime_last_login"`
}

type Token struct {
	// main
	Id       int    `json:"id"`
	Username string `json:"username"`
	Access   string `json:"access"`
	Refresh  string `json:"refresh"`

	// additional
	DatetimeCreated time.Time `json:"datetime_created"`
}

type Task struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

type DumptruckEventState struct {
	Vehid  string  `json:"vehid"`
	Time   string  `json:"time"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Weight float32 `json:"weight"`
	Fuel   float32 `json:"fuel"`
	Speed  float32 `json:"speed"`
}

type TechTrip struct {
	Tripcounter  int32     `json:"tripcounter"`
	Vehid        string    `json:"vehid"`
	Shovid       string    `json:"shovid"`
	Unloadid     string    `json:"unloadid"`
	Worktype     string    `json:"worktype"`
	Timeload     time.Time `json:"timeload"`
	Timeunload   time.Time `json:"timeunload"`
	Movetime     time.Time `json:"movetime"`
	Weigth       int16     `json:"weigth"`
	Bucketcount  int16     `json:"bucketcount"`
	Avspeed      float32   `json:"avspeed"`
	Length       float32   `json:"length"`
	Unloadlength float32   `json:"unloadlength"`
	Loadheight   int16     `json:"loadheight"`
	Unloadheight int16     `json:"unloadheight"`
}

type Resp struct {
	Response []DumptruckEventState `json:"response"`
}

type Trade struct {
	Time   time.Time
	Symbol string
	Price  float64
	IsBuy  bool
}

type UserNew struct {
	Id               int32
	Username         string
	Password         string
	Registrationtime time.Time
	Salary           float32
	Active           bool
}

// TODO structs ////////////////////////////////////////////////////////////////////////////////////////////////////////

//

// TODO methods ////////////////////////////////////////////////////////////////////////////////////////////////////////

func (d DumptruckEventState) Print() {
	fmt.Println(d, reflect.TypeOf(d), fmt.Sprintf("%+v", d))
}

//

// TODO methods ////////////////////////////////////////////////////////////////////////////////////////////////////////

//

// TODO global /////////////////////////////////////////////////////////////////////////////////////////////////////////

var ConfigV = ConfigConstructor()

// TODO global /////////////////////////////////////////////////////////////////////////////////////////////////////////

//

// TODO actions ////////////////////////////////////////////////////////////////////////////////////////////////////////

func Print(obj any) {
	fmt.Println(obj, reflect.TypeOf(obj), fmt.Sprintf("%+v", obj), fmt.Sprintf("%#v", obj))
}

func NewApp(routes []RouteS) error {
	// todo Configure engine
	if ConfigV.DebugGin {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// todo Create engine
	//engine := gin.New()
	engine := gin.Default()

	// todo TEMP
	engine.Static("/static", "./frontend/build/static") //
	engine.LoadHTMLGlob("./frontend/build/index.html")
	engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main website",
		})
	})
	// todo TEMP

	engine.Use(cors.New(cors.Config{
		AllowHeaders: []string{"*"},
		AllowMethods: []string{"*"},
		//AllowAllOrigins: true,
		AllowOrigins: []string{"*"},
		//AllowOriginFunc: func(origin string) bool {
		//	return origin == "http://127.0.0.1:3000"
		//},
		//AllowMethods:     []string{"PUT", "PATCH"},
		//AllowHeaders:     []string{"Origin"},
		//ExposeHeaders:    []string{"Content-Length"},
		//AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))

	// todo Binding routes
	for _, route := range routes {
		switch route.Method {
		case "POST": // Create
			engine.POST(route.RelativePath, Middleware(route.Function, route))
		case "GET": // Read
			engine.GET(route.RelativePath, Middleware(route.Function, route))
		case "PUT": // Update
			engine.PUT(route.RelativePath, Middleware(route.Function, route))
		case "DELETE": // Delete
			engine.DELETE(route.RelativePath, Middleware(route.Function, route))
		default:
			ErrorHandler(errors.New("unknown method"))
		}
	}

	// todo Run engine
	return engine.Run(fmt.Sprintf("%s:%d", ConfigV.HostGin, ConfigV.PortGin))
}

func Middleware(next gin.HandlerFunc, route RouteS) gin.HandlerFunc {

	time.Sleep(time.Millisecond * 1500)

	return func(context *gin.Context) {
		//log.Printf("[%s] %s\n", context.Request.Method, context.Request.RequestURI)

		if route.Auth {
			// get config.headers.Authorization
			authorizationHeader := context.GetHeader("Authorization")
			if authorizationHeader == "" {
				ErrHandlerWithContext(errors.New("authorization failed"), context)
				return
			}

			// get JWT=$Qwerty!1234567
			tokenArr := strings.Split(authorizationHeader, "=")
			if len(tokenArr) < 2 {
				ErrHandlerWithContext(errors.New("authorization failed"), context)
				return
			}

			// check token
			status, err := CheckAccessToken(tokenArr[1])
			if err != nil {
				ErrHandlerWithContext(errors.New("authorization failed"), context)
				return
			}

			if !status {
				ErrHandlerWithContext(errors.New("authorization failed"), context)
				return
			}

			//userID := context.Request.Header.Get("x-id")
			//if userID == "" {
			//	log.Printf("[%s] %s - error: userID is not provided\n", context.Request.Method, context.Request.RequestURI)
			//	ErrHandlerWithContext(errors.New("users id is not provided"), context)
			//	return
			//}
		}

		//
		//ctx := r.Context()
		//ctx = context.WithValue(ctx, "id", userID)
		//
		//r = r.WithContext(ctx)

		next(context)
	}
}

func ErrorHandler(err error) {
	fmt.Printf("error: %s", err.Error())
}

func ErrHandlerWithContext(err error, context *gin.Context) {
	context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	ErrorHandler(err)
}

func CreateDbPgConnection() (*sql.DB, error) {
	source := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		ConfigV.HostPgSQL, ConfigV.PortPgSQL, ConfigV.UserPgSQL, ConfigV.PasswordPgSQL, ConfigV.DatabasePgSQL, ConfigV.SslModePgSQL,
	)
	dbConnection, err := sql.Open(ConfigV.DriverNamePgSQL, source)
	if err != nil {
		return nil, err
	}

	return dbConnection, nil
}

func ExecuteSelectOneDb(object []any, query string, args ...any) error {
	dbConnection, err := CreateDbPgConnection()
	if err != nil {
		return err
	}
	defer func(dbConnection *sql.DB) {
		err = dbConnection.Close()
		if err != nil {
			return
		}
	}(dbConnection)

	rows, err := dbConnection.Query(query, args...)
	if err != nil {
		return err
	}

	rows.Next()
	err = rows.Scan(object...)
	if err != nil {
		return err
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

func ExecuteSelectManyDb(objects *[]string, query string, args ...any) error {
	dbConnection, err := CreateDbPgConnection()
	if err != nil {
		return err
	}
	defer func(dbConnection *sql.DB) {
		err = dbConnection.Close()
		if err != nil {
			return
		}
	}(dbConnection)

	rows, err := dbConnection.Query(query, args...)
	if err != nil {
		return err
	}

	usernames := make([]string, 0)
	for rows.Next() {
		var obj string
		err = rows.Scan(&obj)
		if err != nil {
			return err
		}

		usernames = append(usernames, obj)
	}
	fmt.Println(usernames)
	*objects = usernames

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

func ExecuteRowsDb(query string, args ...any) (*sql.Rows, error) {
	dbConnection, err := CreateDbPgConnection()
	if err != nil {
		return nil, err
	}
	defer func(dbConnection *sql.DB) {
		err = dbConnection.Close()
		if err != nil {
			return
		}
	}(dbConnection)

	rows, err := dbConnection.Query(query, args...)
	if err != nil {
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func ExecuteInsertOrDeleteDb(query string, args ...any) error {
	dbConnection, err := CreateDbPgConnection()
	if err != nil {
		return err
	}
	defer func(dbConnection *sql.DB) {
		err = dbConnection.Close()
		if err != nil {
			return
		}
	}(dbConnection)

	dbTransaction, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	defer func(dbTransaction *sql.Tx) {
		_ = dbTransaction.Rollback()
	}(dbTransaction)

	_, err = dbTransaction.Exec(query, args...)
	if err != nil {
		return err
	}

	err = dbTransaction.Commit()
	if err != nil {
		return err
	}

	return nil
}

func GetRequestBodyData(context *gin.Context) (map[string]any, error) {
	var data = make(map[string]any)
	err := context.ShouldBindJSON(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ConfigConstructor() *ConfigS {
	config_ := &ConfigS{
		// Gin
		HostGin:  "0.0.0.0",
		PortGin:  8083,
		DebugGin: true,

		// Postgresql
		DriverNamePgSQL: "postgres",
		HostPgSQL:       "127.0.0.1",
		PortPgSQL:       5432,
		UserPgSQL:       "pgs_usr",
		PasswordPgSQL:   "12345Qwerty!",
		DatabasePgSQL:   "pgs_db",
		SslModePgSQL:    "disable",

		// Additional
		HashCost:             14,
		HashSalt:             "Qwerty!12345",
		TokenAccessLifetime:  time.Minute * 10,
		TokenRefreshLifetime: time.Minute * 60 * 24,
	}
	ReadConfig(config_, "../.env")
	return config_
}

func ReadConfig(config *ConfigS, filename string) *ConfigS {
	// TODO NEED read and parse from .ENV file
	//return config{DebugGin: c.DebugGin, HostGin: c.HostGin, PORT: c.PORT}
	_ = filename
	return config
}

func HashPassword(passwords ...string) (string, error) {
	password := ""
	for _, pwd := range passwords {
		password += pwd
	}
	if len(password) < 4 {
		return "", errors.New("password very simple")
	} else if len(password) > 41 {
		password = password[:40]
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), ConfigV.HashCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func CreateToken(user *User) (string, string, error) {
	// TODO NEED check token date is not expired
	a := user.Username
	b := user.DatetimeJoined.String()
	c := ConfigV.HashSalt

	access, err := HashPassword(a, b, c)
	if err != nil {
		return "", "", err
	}

	refresh, err := HashPassword(c, b, a)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func CheckAccessToken(accessToken string) (bool, error) {
	// find users in db
	var username string
	var datetimeCreated time.Time
	err := ExecuteSelectOneDb([]any{&username, &datetimeCreated},
		"select username, datetime_created from tokens where access=$1;", accessToken)
	if err != nil {
		return false, err
	}
	if username == "" {
		return false, errors.New("token not found")
	}

	if time.Now().Sub(datetimeCreated).Minutes() > ConfigV.TokenAccessLifetime.Minutes() {
		return false, errors.New("token lifetime expired")
	}

	return true, nil
}

func CreateUsersDatabase() {
	query := "create table users (id serial not null unique, username varchar(255) not null unique, password varchar(255) not null, datetime_joined timestamp not null default now(), is_active bool default 'true');"
	err := ExecuteInsertOrDeleteDb(query)
	if err != nil {
		ErrorHandler(err)
		return
	}

	fmt.Println("database 'users' successfully added")
}

func CreateTokensDatabase() {
	query := "create table tokens (id serial not null unique, username varchar(255) not null unique, access varchar(255) not null, refresh varchar(255) not null, datetime_created timestamp not null default now());"
	err := ExecuteInsertOrDeleteDb(query)
	if err != nil {
		ErrorHandler(err)
		return
	}

	fmt.Println("database 'tokens' successfully added")
}

func CheckCaptchaHandler(context *gin.Context) {
	// get data from request
	//_, err := GetRequestBodyData(context)
	//if err != nil {
	//	ErrHandlerWithContext(err, context)
	//	return
	//}

	context.JSON(http.StatusOK, map[string]string{"response": "success"})
}

func CheckOracle() {
	db, err := sql.Open("oracle", "oracle://DISPATCHER:disp@172.30.23.16:1521/PITENEW")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	rows, err := db.Query("select sysdate from dual")
	if err != nil {
		fmt.Println("Error running query: ")
		fmt.Println(err)
		return
	}
	defer rows.Close()

	var thedate string
	for rows.Next() {

		rows.Scan(&thedate)
	}
	fmt.Printf("The date is: %s\n", thedate)
}

func Trips() []any { //
	db, err := sql.Open("oracle", "oracle://DISPATCHER:disp@172.30.23.16:1521/PITENEW")
	if err != nil {
		ErrorHandler(err)
		fmt.Println(err)
		return nil
	}
	defer db.Close()

	//rows, err := db.Query("select sysdate from dual")
	rows, err := db.Query("SELECT VEHID, TIME, X, Y, WEIGHT, FUEL, SPEED FROM   EVENTSTATEARCHIVE WHERE  MESCOUNTER IN (SELECT MAX(MESCOUNTER)  FROM  EVENTSTATEARCHIVE WHERE  TIME BETWEEN ( SYSDATE - ( 1 / 24 / 60 * 500 ) ) AND SYSDATE GROUP  BY VEHID) ORDER  BY VEHID ")
	if err != nil {
		ErrorHandler(err)
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	var obj DumptruckEventState

	for rows.Next() {
		s := reflect.ValueOf(&obj).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}
		err := rows.Scan(columns...)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(obj)
		fmt.Printf("%+v\n", columns)
		for _, value := range columns {
			fmt.Printf("%+v\n", value)
			fmt.Println(value, reflect.TypeOf(value))
		}
		return columns
	}
	return nil

	//for rows.Next() {
	//	s := reflect.ValueOf(&obj).Elem()
	//	numCols := s.NumField()
	//	columns := make([]interface{}, numCols)
	//	for i := 0; i < numCols; i++ {
	//		field := s.Field(i)
	//		columns[i] = field.Addr().Interface()
	//	}
	//	err := rows.Scan(columns...)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Println(obj)
	//}

	//vehstats := make([]any, 0)
	//for rows.Next() {
	//	var obj any
	//	//fmt.Println(rows.Columns())
	//	//fmt.Println(rows.ColumnTypes())
	//	err = rows.Scan(&obj)
	//	if err != nil {
	//		ErrorHandler(err)
	//		fmt.Println(err)
	//		return
	//	}
	//	vehstats = append(vehstats, obj)
	//}
	//fmt.Println(vehstats)
}

func DumptrucksEventStateHandler(context *gin.Context) {
	db, err := sql.Open("oracle", "oracle://DISPATCHER:disp@172.30.23.16:1521/PITENEW")
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}
	defer db.Close()

	rows, err := db.Query(
		`
SELECT vehid,
       TIME,
       x,
       y,
       weight,
       fuel,
       speed
FROM   eventstatearchive
WHERE  mescounter IN (SELECT max(mescounter)
                      FROM   eventstatearchive
                      WHERE  TIME BETWEEN ( SYSDATE - ( 1 / 24 / 60 * 500 ) ) AND SYSDATE
                      GROUP  BY vehid)
ORDER  BY vehid 
`)
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}
	defer rows.Close()

	objs := make([]DumptruckEventState, 0)
	for rows.Next() {
		var obj DumptruckEventState
		err = rows.Scan(&obj.Vehid, &obj.Time, &obj.X, &obj.Y, &obj.Weight, &obj.Fuel, &obj.Speed)
		if err != nil {
			ErrHandlerWithContext(err, context)
			return
		}
		objs = append(objs, obj)
	}
	err = rows.Err()
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	Print(objs)
	for _, item := range objs {
		Print(item)
	}

	context.JSON(http.StatusOK, map[string]map[string][]DumptruckEventState{"response": {"data": objs}})
}

func TehtripsHandler(context *gin.Context) {
	db, err := sql.Open("oracle", "oracle://DISPATCHER:disp@172.30.23.16:1521/PITENEW")
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}
	defer db.Close()

	rows, err := db.Query(
		`
SELECT tripcounter,
       vehid,
       shovid,
       unloadid,
       worktype,
       timeload,
       timeunload,
       movetime,
       weight,
       nvl(bucketcount, -1) bucketcount,
       avspeed,
       length,
       unloadlength,
       loadheight,
       unloadheight
FROM   vehtrips vt
WHERE  vt.timeunload BETWEEN getpredefinedtimefrom('за указанную смену', getcurshiftnum(0, SYSDATE), getcurshiftdate(0, SYSDATE)) AND getpredefinedtimeto('за указанную смену', getcurshiftnum(0, SYSDATE), getcurshiftdate(0, SYSDATE))
ORDER  BY timeload DESC 
`)
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}
	defer rows.Close()

	objs := make([]TechTrip, 0)
	for rows.Next() {
		var obj TechTrip
		err = rows.Scan(&obj.Tripcounter, &obj.Vehid, &obj.Shovid, &obj.Unloadid, &obj.Worktype, &obj.Timeload, &obj.Timeunload, &obj.Movetime, &obj.Weigth, &obj.Bucketcount, &obj.Avspeed, &obj.Length, &obj.Unloadlength, &obj.Loadheight, &obj.Unloadheight)
		if err != nil {
			ErrHandlerWithContext(err, context)
			return
		}
		objs = append(objs, obj)
	}
	err = rows.Err()
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	Print(objs)
	for _, item := range objs {
		Print(item)
	}

	context.JSON(http.StatusOK, map[string]map[string][]TechTrip{"response": {"data": objs}})
}

func RegisterUserHandler(context *gin.Context) {
	// get and check param
	username := context.PostForm("username")
	if username == "" {
		ErrHandlerWithContext(errors.New("username incorrect"), context)
		return
	}

	// get and check param
	password := context.PostForm("password")
	if password == "" {
		ErrHandlerWithContext(errors.New("password incorrect"), context)
		return
	}

	// hash password
	password, err := HashPassword(password)
	if password == "" {
		ErrHandlerWithContext(err, context)
		return
	}

	// write to db
	err = ExecuteInsertOrDeleteDb("insert into users (username, password) values ($1, $2);", username, password)
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	time.Sleep(time.Millisecond * 2000)

	// return
	context.JSON(http.StatusCreated, map[string]string{"response": "success"})
}

func LoginUserHandler(context *gin.Context) {
	// get and check username
	username := context.PostForm("username")
	if username == "" {
		ErrHandlerWithContext(errors.New("username incorrect"), context)
		return
	}

	// get and check password
	password := context.PostForm("password")
	if password == "" {
		ErrHandlerWithContext(errors.New("password incorrect"), context)
		return
	}

	// find users in db
	user := User{}
	err := ExecuteSelectOneDb([]any{&user.Id, &user.Username, &user.Password},
		"select id, username, password from users where username=$1;", username)
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}
	if user.Username == "" {
		ErrHandlerWithContext(errors.New("users not found"), context)
		return
	}

	// check hash password
	success := CheckPasswordHash(password, user.Password)
	if success != true {
		ErrHandlerWithContext(errors.New("password didn't match"), context)
		return
	}

	// generate tokens
	access, refresh, err := CreateToken(&user)
	if success != true {
		ErrHandlerWithContext(err, context)
		return
	}

	// write tokens to db
	err = ExecuteInsertOrDeleteDb("insert into tokens (username, access, refresh) values ($1, $2, $3);", user.Username, access, refresh)
	if err != nil {
		err = ExecuteInsertOrDeleteDb("update tokens set access=$1, refresh=$2 where username = $3;", access, refresh, user.Username)
		if err != nil {
			ErrHandlerWithContext(err, context)
			return
		}
	}

	// return
	context.JSON(http.StatusOK, map[string]map[string]string{"response": {"access": access, "refresh": refresh}})
}

func GetAllUsersHandler(context *gin.Context) {
	usernames := make([]string, 0)
	err := ExecuteSelectManyDb(&usernames, "select username from users;")
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	context.JSON(http.StatusOK, map[string]map[string]any{"response": {"list": usernames}})
}

func CreateTasksDatabase() {
	err := ExecuteInsertOrDeleteDb("create table tasks (id serial not null unique, title varchar(255) not null unique);")
	if err != nil {
		ErrorHandler(err)
		return
	}

	fmt.Println("database 'tasks' successfully added")
}

func CreateTaskHandler(context *gin.Context) {
	// get and check param
	title := context.PostForm("title")
	if title == "" {
		ErrHandlerWithContext(errors.New("title incorrect"), context)
		return
	}

	// insert to db
	err := ExecuteInsertOrDeleteDb("insert into tasks (title) values ($1);", title)
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	context.JSON(http.StatusCreated, map[string]string{"response": "successfully created"})
}

func ReadTaskHandler(context *gin.Context) {
	// get id from url
	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	// create object
	task := Task{}

	// select from db
	err = ExecuteSelectOneDb([]any{&task.Id, &task.Title}, "select id, title from tasks where id = $1", id)
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	context.JSON(http.StatusOK, map[string]any{"response": task})
}

func ReadTasksHandler(context *gin.Context) {
	// select from db
	rows, err := ExecuteRowsDb("select id, title from tasks order by id asc;")
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	// create objects
	tasks := make([]Task, 0)

	// fulling objects
	for rows.Next() {
		task := Task{}
		err = rows.Scan(&task.Id, &task.Title)
		if err != nil {
			ErrHandlerWithContext(err, context)
			return
		}
		tasks = append(tasks, task)
	}

	context.JSON(http.StatusOK, map[string]map[string]any{"response": {"list": tasks}})
}

func UpdateTaskHandler(context *gin.Context) {
	// get id from url
	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	// get and check param
	title := context.PostForm("title")
	if title == "" {
		ErrHandlerWithContext(errors.New("title incorrect"), context)
		return
	}

	// update into db
	err = ExecuteInsertOrDeleteDb("update tasks set title=$1 where id = $2;", title, id)
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	context.JSON(http.StatusOK, map[string]string{"response": "successfully updated"})
}

func DeleteTaskHandler(context *gin.Context) {
	// get id from url
	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	// delete from db
	err = ExecuteInsertOrDeleteDb("delete from tasks where id = $1;", id)
	if err != nil {
		ErrHandlerWithContext(err, context)
		return
	}

	context.JSON(http.StatusOK, map[string]string{"response": "successfully deleted"})
}

func CreateSqlite3() {
	db, err := sql.Open("sqlite3", "./sqlite3.db")
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal(err)
		return
	}
	defer db.Close()

	// todo Users [CREATE]
	//	schemaSQL := `
	//--DROP TABLE users;
	//CREATE TABLE IF NOT EXISTS users (
	// id INTEGER PRIMARY KEY AUTOINCREMENT,
	// username VARCHAR(256) unique NOT NULL DEFAULT '',
	// password VARCHAR(256) NOT NULL DEFAULT '',
	// registrationtime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	// salary FLOAT NOT NULL DEFAULT '0.0',
	// active BOOLEAN NOT NULL DEFAULT 'false'
	//);
	//CREATE INDEX IF NOT EXISTS username_idx ON users(username);
	//rows, err := db.Exec(schemaSQL)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	log.Fatal(err)
	//	return
	//}
	//Print(rows)
	// todo Users [CREATE]

	// todo Users [INSERT]
	//`
	//	schemaSQL := `
	//INSERT INTO users(username, password, registrationtime, salary, active) values('Bogdan1','Qwerty!12345',CURRENT_TIMESTAMP, '259.02','false')
	//`
	//rows, err := db.Exec(schemaSQL)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	log.Fatal(err)
	//	return
	//}
	//Print(rows)
	// todo Users [INSERT]

	// todo Users [SELECT MANY]
	//rows, err := db.Query(`select id, username, password, registrationtime, salary, active from users`)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	log.Fatal(err)
	//	return
	//}
	//defer rows.Close()
	//
	//users := make([]UserNew, 0)
	//for rows.Next() {
	//	var user UserNew
	//	err = rows.Scan(&user.Id, &user.Username, &user.Password, &user.Registrationtime, &user.Salary, &user.Active)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		log.Fatal(err)
	//		return
	//	}
	//	users = append(users, user)
	//}
	//
	//err = rows.Err()
	//if err != nil {
	//	fmt.Println(err.Error())
	//	log.Fatal(err)
	//	return
	//}
	//
	//Print(users)
	//for _, item := range users {
	//	Print(item)
	//}
	// todo Users [SELECT MANY]
}

// TODO actions ////////////////////////////////////////////////////////////////////////////////////////////////////////

//

// TODO main ///////////////////////////////////////////////////////////////////////////////////////////////////////////

func run() {
	err := NewApp([]RouteS{
		// users(s)
		{"GET", "/api/dumptrucks/eventstate/", DumptrucksEventStateHandler, false}, // Create   | curl -v -X POST 127.0.0.1:8080/users -d '{"username":"admin", "password": "admin"}'
		{"GET", "/api/trips/shift/", TehtripsHandler, false},                       // Create   | curl -v -X POST 127.0.0.1:8080/users -d '{"username":"admin", "password": "admin"}'
		//{"GET", "/api/users/check_captcha", CheckCaptchaHandler, false}, // Create   | curl -v -X POST 127.0.0.1:8080/users -d '{"username":"admin", "password": "admin"}'
		//{"POST", "/api/users/register", RegisterUserHandler, false},     // Create   | curl -v -X POST 127.0.0.1:8080/users -d '{"username":"admin", "password": "admin"}'
		//{"POST", "/api/users/login", LoginUserHandler, false},           // Create   | curl -v -X POST 127.0.0.1:8080/users/login -d '{"username":"admin", "password": "admin"}'
		//{"GET", "/api/users/", GetAllUsersHandler, false},               // Create   | curl -v -X POST 127.0.0.1:8080/users/login -d '{"username":"admin", "password": "admin"}'
		//
		//// tasks(s)
		//{"POST", "/api/tasks", CreateTaskHandler, true},       // Create   | curl -v -X POST 127.0.0.1:8080/tasks -d '{"title":"Amon Ra","author":"V.Pelevin"}'
		//{"GET", "/api/tasks/:id", ReadTaskHandler, true},      // Read     | curl -v -X GET 127.0.0.1:8080/tasks/1
		//{"GET", "/api/tasks", ReadTasksHandler, true},         // Read all | curl -v -X GET 127.0.0.1:8080/tasks
		//{"PUT", "/api/tasks/:id", UpdateTaskHandler, true},    // Update   | curl -v -X PUT 127.0.0.1:8080/tasks/1 -d '{"title":"War and peace","author":"N.Tolstoy"}'
		//{"DELETE", "/api/tasks/:id", DeleteTaskHandler, true}, // Delete   | curl -v -H 'x-id:1' -X DELETE 127.0.0.1:8080/tasks/1
	})
	if err != nil {
		ErrorHandler(err)
		log.Fatal(err)
		return
	}
}

func main() {
	// tsx home add data
	// tsx add fields
	// tsx search, filter... fields
	// tsx home add data
	// add file field
	// add redis-cache

	run()
	//CreateSqlite3()

	//CheckOracle()

	//CreateUsersDatabase()
	//CreateTokensDatabase()
	//tasks.CreateTasksDatabase()
}

// TODO main ///////////////////////////////////////////////////////////////////////////////////////////////////////////

//

// TODO extra //////////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO extra //////////////////////////////////////////////////////////////////////////////////////////////////////////

//
