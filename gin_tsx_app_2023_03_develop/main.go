package main

//

// TODO imports ////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"os"
	"reflect"

	//_ "github.com/godror/godror"
	//_ "github.com/lib/pq"
	//_ "github.com/sijms/go-ora"
	//_ "github.com/mattn/go-sqlite3"
	_ "github.com/sijms/go-ora/v2"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TODO imports ////////////////////////////////////////////////////////////////////////////////////////////////////////

//

// TODO structs ////////////////////////////////////////////////////////////////////////////////////////////////////////

type StructConfig struct {
	// Gin
	DebugGin    bool
	HostGin     string
	PortGin     int
	PrintErr    bool
	LogFile     string
	CacheLow    int `json:"CacheLow"`
	CacheMiddle int `json:"CacheMiddle"`
	CacheHigh   int `json:"CacheHigh"`

	// CORS
	Cors cors.Config

	// Static
	StaticFolders []map[string]string

	// Templates
	Templates []string

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
	Kpd          float32   `json:"kpd"`
}

type TechTripWithKpd struct {
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
	Kpd          float32   `json:"kpd"`
}

type classLogs struct{}

type classGin struct{}

type classUtils struct{}

// TODO structs ////////////////////////////////////////////////////////////////////////////////////////////////////////

//

// TODO methods ////////////////////////////////////////////////////////////////////////////////////////////////////////

// todo classLogs

func (c classLogs) LogErrToConsole(err error) {
	classLogs{}.LogTxtToConsole(err.Error())
}

func (c classLogs) LogTxtToConsole(text string) {
	fmt.Println(fmt.Sprintf("error[%s]: %s\n", strings.Split(time.Now().String(), ".")[0], text))
}

func (c classLogs) LogToFile(text string) {
	message := fmt.Sprintf("error[%s]: %s\n", strings.Split(time.Now().String(), ".")[0], text)
	if true {
		fmt.Println(message)
	}
	file, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		c.LogErrToConsole(err)
		return
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			c.LogErrToConsole(err)
			return
		}
	}(file)
	_, err = file.WriteString(message)
	if err != nil {
		c.LogErrToConsole(err)
		return
	}
}

// todo classGin

func (c classGin) Middleware(route struct {
	Method       string
	RelativePath string
	Function     gin.HandlerFunc
	Auth         bool
}) gin.HandlerFunc {
	return func(context *gin.Context) {
		//log.Printf("[%s] %s\n", context.Request.Method, context.Request.RequestURI)

		// todo check subnet [16, 23...]
		switch strings.Split(context.Request.RemoteAddr, ".")[2] { // [172 30 16 225:58122]
		case
			"16",
			"23":
			// success
		default:
			classGin{}.ErrHandlerWithContext(errors.New("not have access from this subnet"), context)
			return
		}
		// todo check subnet [16, 23...]

		if route.Auth {
			// get config.headers.Authorization
			authorizationHeader := context.GetHeader("Authorization")
			if authorizationHeader == "" {
				classGin{}.ErrHandlerWithContext(errors.New("authorization failed"), context)
				return
			}

			// get JWT=$Qwerty!1234567
			tokenArr := strings.Split(authorizationHeader, "=")
			if len(tokenArr) < 2 {
				classGin{}.ErrHandlerWithContext(errors.New("authorization failed"), context)
				return
			}

			// check token
			status, err := CheckAccessToken(tokenArr[1])
			if err != nil {
				classGin{}.ErrHandlerWithContext(errors.New("authorization failed"), context)
				return
			}

			if !status {
				classGin{}.ErrHandlerWithContext(errors.New("authorization failed"), context)
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

		route.Function(context)
	}
}

func (c classGin) ErrHandlerWithContext(err error, context *gin.Context) {
	context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	classUtils{}.ErrorHandler(err)
}

// todo classUtils

func (c classUtils) ReadEnv(fileName string) StructConfig {
	jsonFile, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		classLogs{}.LogToFile(err.Error())
		return StructConfig{}
	}
	defer func(jsonFile *os.File) {
		err = jsonFile.Close()
		if err != nil {
			classLogs{}.LogToFile(err.Error())
			return
		}
	}(jsonFile)

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		classLogs{}.LogToFile(err.Error())
		return StructConfig{}
	}

	var conf StructConfig
	err = json.Unmarshal(byteValue, &conf)
	if err != nil {
		classLogs{}.LogToFile(err.Error())
		return StructConfig{}
	}

	return conf
}

func (c classUtils) Print(obj any) {
	fmt.Println(obj, reflect.TypeOf(obj), fmt.Sprintf("%+v", obj), fmt.Sprintf("%#v", obj))
	//fmt.Println(reflect.TypeOf(obj), fmt.Sprintf("%#v", obj))
}

func (c classUtils) ErrorNewHandler(detail string) {
	classUtils{}.ErrorHandler(errors.New(detail))
}

func (c classUtils) ErrorHandler(err error) {
	fmt.Printf("error: %s", err.Error())
}

//

// TODO methods ////////////////////////////////////////////////////////////////////////////////////////////////////////

//

// TODO global /////////////////////////////////////////////////////////////////////////////////////////////////////////

var VarConfig = classUtils{}.ReadEnv("env.json")

var VarRoutes = []struct {
	Method       string
	RelativePath string
	Function     gin.HandlerFunc
	Auth         bool
}{
	// SPA
	{"GET", "/", IndexHandler, false}, // Get
	//
	// dumptruck(s)
	{"GET", "/api/dumptrucks/eventstate/", DumptrucksEventStateHandler, false}, // Create   | curl -v -X POST 127.0.0.1:8080/users -d '{"username":"admin", "password": "admin"}'
	{"GET", "/api/trips/shift/", TehtripsHandler, false},                       // Create   | curl -v -X POST 127.0.0.1:8080/users -d '{"username":"admin", "password": "admin"}'
	//
	// users(s)
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
}

var TechTripsCache struct {
	Trips []TechTrip
	Time  time.Time
}

var DumptruckEventStateCache struct {
	Events []DumptruckEventState
	Time   time.Time
}

// TODO global /////////////////////////////////////////////////////////////////////////////////////////////////////////

//

// TODO actions ////////////////////////////////////////////////////////////////////////////////////////////////////////

func IndexHandler(context *gin.Context) {
	context.HTML(http.StatusOK, "index.html", gin.H{
		"title": "SPA",
	})
}

func RedirectEmptyHandler(context *gin.Context) {
	context.Redirect(http.StatusFound, "/")
}

func DumptrucksEventStateHandler(context *gin.Context) {
	// todo check cache
	if int(time.Now().Sub(DumptruckEventStateCache.Time).Seconds()) < VarConfig.CacheLow {
		context.JSON(http.StatusOK, map[string]map[string]any{"response": {"data": DumptruckEventStateCache.Events}})
		return
	}

	db, err := sql.Open("oracle", "oracle://DISPATCHER:disp@172.30.23.16:1521/PITENEW")
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			classGin{}.ErrHandlerWithContext(err, context)
			return
		}
	}(db)

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
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			classGin{}.ErrHandlerWithContext(err, context)
			return
		}
	}(rows)

	objs := make([]DumptruckEventState, 0)
	for rows.Next() {
		var obj DumptruckEventState
		err = rows.Scan(&obj.Vehid, &obj.Time, &obj.X, &obj.Y, &obj.Weight, &obj.Fuel, &obj.Speed)
		if err != nil {
			classGin{}.ErrHandlerWithContext(err, context)
			return
		}
		objs = append(objs, obj)
	}
	err = rows.Err()
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	//Print(objs)
	//for _, item := range objs {
	//	Print(item)
	//}

	DumptruckEventStateCache = struct {
		Events []DumptruckEventState
		Time   time.Time
	}{
		Events: objs,
		Time:   time.Now(),
	}

	context.JSON(http.StatusOK, map[string]map[string]any{"response": {"data": objs}})
}

func TehtripsHandler(context *gin.Context) {
	// todo check cache
	if int(time.Now().Sub(TechTripsCache.Time).Seconds()) < VarConfig.CacheMiddle {
		context.JSON(http.StatusOK, map[string]map[string]any{"response": {"data": TechTripsCache.Trips}})
		return
	}

	db, err := sql.Open("oracle", "oracle://DISPATCHER:disp@172.30.23.16:1521/PITENEW")
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			classGin{}.ErrHandlerWithContext(err, context)
			return
		}
	}(db)

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
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			classGin{}.ErrHandlerWithContext(err, context)
			return
		}
	}(rows)

	file, _ := excelize.OpenFile("data/table_norms.xlsx")
	matrix, _ := file.GetRows("Экскаваторы", excelize.Options{RawCellValue: true}) //rows, err := f.GetRows("Sheet1")
	dictNormShovs := make(map[string]map[string]float64)
	for _, row := range matrix[1:] {
		dictNormShov := make(map[string]float64)

		dictNormShov["Вскрыша скальная"], _ = strconv.ParseFloat(row[1], 64)
		dictNormShov["Вскрыша рыхлая"], _ = strconv.ParseFloat(row[2], 64)
		dictNormShov["Вскрыша транзитная"], _ = strconv.ParseFloat(row[3], 64)
		dictNormShov["Руда скальная"], _ = strconv.ParseFloat(row[4], 64)
		dictNormShov["ВКП скала"], _ = strconv.ParseFloat(row[5], 64)
		dictNormShov[""] = 0.001

		dictNormShovs[row[0]] = dictNormShov
	}

	file, _ = excelize.OpenFile("data/table_norms.xlsx")
	matrix, _ = file.GetRows("Самосвалы", excelize.Options{RawCellValue: true}) //rows, err := f.GetRows("Sheet1")
	dictNormDumptrucks := make(map[string]map[string]float64)
	for _, row := range matrix[1:] {
		dictNormDumptruck := make(map[string]float64)

		dictNormDumptruck["Health"], _ = strconv.ParseFloat(row[1], 64)
		dictNormDumptruck["Speed full"], _ = strconv.ParseFloat(row[2], 64)
		dictNormDumptruck["Speed empty"], _ = strconv.ParseFloat(row[3], 64)
		dictNormDumptruck["Unload time"], _ = strconv.ParseFloat(row[4], 64)
		dictNormDumptruck["Wait time"], _ = strconv.ParseFloat(row[5], 64)
		dictNormDumptruck[""] = 0.001

		dictNormDumptrucks[row[0]] = dictNormDumptruck
	}

	objs := make([]TechTrip, 0)
	for rows.Next() {
		var obj TechTrip
		err = rows.Scan(&obj.Tripcounter, &obj.Vehid, &obj.Shovid, &obj.Unloadid, &obj.Worktype, &obj.Timeload, &obj.Timeunload, &obj.Movetime, &obj.Weigth, &obj.Bucketcount, &obj.Avspeed, &obj.Length, &obj.Unloadlength, &obj.Loadheight, &obj.Unloadheight)
		if err != nil {
			classGin{}.ErrHandlerWithContext(err, context)
			return
		}
		obj.Kpd = ResByTrip(&obj, dictNormShovs, dictNormDumptrucks)
		objs = append(objs, obj)
	}
	err = rows.Err()
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	TechTripsCache = struct {
		Trips []TechTrip
		Time  time.Time
	}{
		Trips: objs,
		Time:  time.Now(),
	}

	context.JSON(http.StatusOK, map[string]map[string]any{"response": {"data": objs}})
}

func ResByTrip(trip *TechTrip, dictNormShovs map[string]map[string]float64,
	dictNormDumptrucks map[string]map[string]float64) float32 {

	//dictNormShovs := map[string]map[string]float64{
	//	"001": {
	//		"Вскрыша скальная":   2.5,
	//		"Вскрыша рыхлая":     2.5,
	//		"Вскрыша транзитная": 2.5,
	//		"Руда скальная":      2.5,
	//		"ВКП скала":          2.5,
	//	},
	//}

	//dictNormShovs := map[string]float64{
	//	"Вскрыша скальная":   2.5,
	//	"Вскрыша рыхлая":     2.5,
	//	"Вскрыша транзитная": 2.5,
	//	"Руда скальная":      2.5,
	//	"ВКП скала":          2.5,
	//}
	load := 0.0
	if dictNormShovs[trip.Shovid][trip.Worktype] > 0 {
		load = dictNormShovs[trip.Shovid][trip.Worktype]
	}

	//dictPath := map[string]map[string]float64{
	//	"128": {
	//		"Health":      100.0,
	//		"Speed full":  20.5,
	//		"Speed empty": 22.5,
	//		"Unload time": 0.5,
	//		"Wait time":   0.0,
	//	},
	//}

	//dictPath := map[string]float64{
	//	"Health":      100.0,
	//	"Speed full":  20.5,
	//	"Speed empty": 22.5,
	//	"Unload time": 0.5,
	//	"Wait time":   0.0,
	//}
	pathNorm := 60.0
	if dictNormDumptrucks[trip.Vehid]["Speed full"] > 0 {
		pathNorm = dictNormDumptrucks[trip.Vehid]["Speed full"]
	}

	path := float64(trip.Length) / pathNorm * 60
	unload := dictNormDumptrucks[trip.Vehid]["Unload time"]

	returnPathNorm := 60.0
	if dictNormDumptrucks[trip.Vehid]["Speed empty"] > 0 {
		pathNorm = dictNormDumptrucks[trip.Vehid]["Speed empty"]
	}
	returnPath := float64(trip.Length) / returnPathNorm * 60
	wait := dictNormDumptrucks[trip.Vehid]["Wait time"]

	resNorm := load + path + unload + returnPath + wait
	resFact := trip.Timeunload.Sub(trip.Timeload).Minutes()
	kpd := resNorm / resFact * 100
	//Print(kpd)
	//fmt.Printf(
	//	"Total kpd: %f | res: %f == norma: %f [load - %f, path - %f, unload - %f, returnPath - %f, wait - %f]",
	//	kpd, resFact, resNorm, load, path, unload, returnPath, wait)

	return float32(kpd)
}

func getFromExcel() {
	t := time.Now()

	file, _ := excelize.OpenFile("data/table_norms.xlsx")
	//matrix, _ := file.GetRows("Экскаваторы", excelize.Options{RawCellValue: true}) //rows, err := f.GetRows("Sheet1")
	matrix, _ := file.GetRows("Экскаваторы") //rows, err := f.GetRows("Sheet1")
	dictNormShovs := make(map[string]map[string]float64)
	for _, row := range matrix[1:] {
		dictNormShov := make(map[string]float64)

		dictNormShov["Вскрыша скальная"], _ = strconv.ParseFloat(row[1], 64)
		dictNormShov["Вскрыша рыхлая"], _ = strconv.ParseFloat(row[2], 64)
		dictNormShov["Вскрыша транзитная"], _ = strconv.ParseFloat(row[3], 64)
		dictNormShov["Руда скальная"], _ = strconv.ParseFloat(row[4], 64)
		dictNormShov["ВКП скала"], _ = strconv.ParseFloat(row[5], 64)
		dictNormShov[""] = 0.001

		dictNormShovs[row[0]] = dictNormShov
	}
	fmt.Println(dictNormShovs)

	fmt.Printf("ELAPSED TIME: %f", time.Since(t).Seconds())
}

// TODO actions ////////////////////////////////////////////////////////////////////////////////////////////////////////

//

// TODO main ///////////////////////////////////////////////////////////////////////////////////////////////////////////

func runGin() {
	// todo Configure framework
	if VarConfig.DebugGin {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// todo Create engine
	engine := gin.Default() // engine := gin.New()

	// todo Cross-Origin Resource Sharing (CORS)
	engine.Use(cors.New(VarConfig.Cors))
	//engine.Use(cors.New(cors.Config{
	//	AllowHeaders:     []string{"*"},
	//	AllowMethods:     []string{"*"},
	//	AllowOrigins:     []string{"*"},
	//	AllowCredentials: true,
	//	AllowFiles:       true,
	//	AllowWebSockets:  true,
	//	MaxAge:           12 * time.Hour,
	//}))

	// todo Configure static
	for _, staticFolder := range VarConfig.StaticFolders {
		engine.Static(staticFolder["relativePath"], staticFolder["root"]) // engine.Static("/static", "./frontend/build/static") // relativePath, root
	}

	// todo Configure html
	for _, template := range VarConfig.Templates {
		engine.LoadHTMLGlob(template) // engine.LoadHTMLGlob("./frontend/build/index.html") // pattern
	}

	// todo Binding routes
	for _, route := range VarRoutes {
		switch route.Method {
		case "GET": // Read
			engine.GET(route.RelativePath, classGin{}.Middleware(route))
		case "POST": // Create
			engine.POST(route.RelativePath, classGin{}.Middleware(route))
		case "PUT": // Update
			engine.PUT(route.RelativePath, classGin{}.Middleware(route))
		case "PATCH": // Update
			engine.PUT(route.RelativePath, classGin{}.Middleware(route))
		case "DELETE": // Delete
			engine.DELETE(route.RelativePath, classGin{}.Middleware(route))
		default:
			classLogs{}.LogToFile("unknown method")
		}
	}

	// todo Redirect 404
	engine.NoRoute(RedirectEmptyHandler)

	// todo Run engine
	if err := engine.Run(fmt.Sprintf("%s:%d", VarConfig.HostGin, VarConfig.PortGin)); err != nil {
		classLogs{}.LogToFile(err.Error())
	}
}

func main() {
	// todo Task list
	// tsx home add data
	// tsx add fields
	// tsx search, filter... fields
	// tsx home add data
	// add file field

	runGin()
	//getFromExcel()
	//conf := classUtils{}.ReadEnv("env.json")
	//classUtils{}.Print(conf)

	//TripsAnalyzer()
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

// TODO junk ///////////////////////////////////////////////////////////////////////////////////////////////////////////

func CreateDbPgConnection() (*sql.DB, error) {
	source := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		VarConfig.HostPgSQL, VarConfig.PortPgSQL, VarConfig.UserPgSQL, VarConfig.PasswordPgSQL, VarConfig.DatabasePgSQL, VarConfig.SslModePgSQL,
	)
	dbConnection, err := sql.Open(VarConfig.DriverNamePgSQL, source)
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
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), VarConfig.HashCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func CreateToken(user *User) (string, string, error) {
	// TODO NEED check token date is not expired
	a := user.Username
	b := user.DatetimeJoined.String()
	c := VarConfig.HashSalt

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

	if time.Now().Sub(datetimeCreated).Minutes() > VarConfig.TokenAccessLifetime.Minutes() {
		return false, errors.New("token lifetime expired")
	}

	return true, nil
}

func GetRequestBodyData(context *gin.Context) (map[string]any, error) {
	var data = make(map[string]any)
	err := context.ShouldBindJSON(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func RegisterUserHandler(context *gin.Context) {
	// get and check param
	username := context.PostForm("username")
	if username == "" {
		classGin{}.ErrHandlerWithContext(errors.New("username incorrect"), context)
		return
	}

	// get and check param
	password := context.PostForm("password")
	if password == "" {
		classGin{}.ErrHandlerWithContext(errors.New("password incorrect"), context)
		return
	}

	// hash password
	password, err := HashPassword(password)
	if password == "" {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	// write to db
	err = ExecuteInsertOrDeleteDb("insert into users (username, password) values ($1, $2);", username, password)
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
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
		classGin{}.ErrHandlerWithContext(errors.New("username incorrect"), context)
		return
	}

	// get and check password
	password := context.PostForm("password")
	if password == "" {
		classGin{}.ErrHandlerWithContext(errors.New("password incorrect"), context)
		return
	}

	// find users in db
	user := User{}
	err := ExecuteSelectOneDb([]any{&user.Id, &user.Username, &user.Password},
		"select id, username, password from users where username=$1;", username)
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}
	if user.Username == "" {
		classGin{}.ErrHandlerWithContext(errors.New("users not found"), context)
		return
	}

	// check hash password
	success := CheckPasswordHash(password, user.Password)
	if success != true {
		classGin{}.ErrHandlerWithContext(errors.New("password didn't match"), context)
		return
	}

	// generate tokens
	access, refresh, err := CreateToken(&user)
	if success != true {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	// write tokens to db
	err = ExecuteInsertOrDeleteDb("insert into tokens (username, access, refresh) values ($1, $2, $3);", user.Username, access, refresh)
	if err != nil {
		err = ExecuteInsertOrDeleteDb("update tokens set access=$1, refresh=$2 where username = $3;", access, refresh, user.Username)
		if err != nil {
			classGin{}.ErrHandlerWithContext(err, context)
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
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	context.JSON(http.StatusOK, map[string]map[string]any{"response": {"list": usernames}})
}

func CreateTasksDatabase() {
	err := ExecuteInsertOrDeleteDb("create table tasks (id serial not null unique, title varchar(255) not null unique);")
	if err != nil {
		classUtils{}.ErrorHandler(err)
		return
	}

	fmt.Println("database 'tasks' successfully added")
}

func CreateTaskHandler(context *gin.Context) {
	// get and check param
	title := context.PostForm("title")
	if title == "" {
		classGin{}.ErrHandlerWithContext(errors.New("title incorrect"), context)
		return
	}

	// insert to db
	err := ExecuteInsertOrDeleteDb("insert into tasks (title) values ($1);", title)
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	context.JSON(http.StatusCreated, map[string]string{"response": "successfully created"})
}

func ReadTaskHandler(context *gin.Context) {
	// get id from url
	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	// create object
	task := Task{}

	// select from db
	err = ExecuteSelectOneDb([]any{&task.Id, &task.Title}, "select id, title from tasks where id = $1", id)
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	context.JSON(http.StatusOK, map[string]any{"response": task})
}

func ReadTasksHandler(context *gin.Context) {
	// select from db
	rows, err := ExecuteRowsDb("select id, title from tasks order by id asc;")
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	// create objects
	tasks := make([]Task, 0)

	// fulling objects
	for rows.Next() {
		task := Task{}
		err = rows.Scan(&task.Id, &task.Title)
		if err != nil {
			classGin{}.ErrHandlerWithContext(err, context)
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
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	// get and check param
	title := context.PostForm("title")
	if title == "" {
		classGin{}.ErrHandlerWithContext(errors.New("title incorrect"), context)
		return
	}

	// update into db
	err = ExecuteInsertOrDeleteDb("update tasks set title=$1 where id = $2;", title, id)
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	context.JSON(http.StatusOK, map[string]string{"response": "successfully updated"})
}

func DeleteTaskHandler(context *gin.Context) {
	// get id from url
	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
		return
	}

	// delete from db
	err = ExecuteInsertOrDeleteDb("delete from tasks where id = $1;", id)
	if err != nil {
		classGin{}.ErrHandlerWithContext(err, context)
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

func CreateUsersDatabase() {
	query := "create table users (id serial not null unique, username varchar(255) not null unique, password varchar(255) not null, datetime_joined timestamp not null default now(), is_active bool default 'true');"
	err := ExecuteInsertOrDeleteDb(query)
	if err != nil {
		classUtils{}.ErrorHandler(err)
		return
	}

	fmt.Println("database 'users' successfully added")
}

func CreateTokensDatabase() {
	query := "create table tokens (id serial not null unique, username varchar(255) not null unique, access varchar(255) not null, refresh varchar(255) not null, datetime_created timestamp not null default now());"
	err := ExecuteInsertOrDeleteDb(query)
	if err != nil {
		classUtils{}.ErrorHandler(err)
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
		classUtils{}.ErrorHandler(err)
		fmt.Println(err)
		return nil
	}
	defer db.Close()

	//rows, err := db.Query("select sysdate from dual")
	rows, err := db.Query("SELECT VEHID, TIME, X, Y, WEIGHT, FUEL, SPEED FROM   EVENTSTATEARCHIVE WHERE  MESCOUNTER IN (SELECT MAX(MESCOUNTER)  FROM  EVENTSTATEARCHIVE WHERE  TIME BETWEEN ( SYSDATE - ( 1 / 24 / 60 * 500 ) ) AND SYSDATE GROUP  BY VEHID) ORDER  BY VEHID ")
	if err != nil {
		classUtils{}.ErrorHandler(err)
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
		//log.Println(obj)
		//fmt.Printf("%+v\n", columns)
		//for _, value := range columns {
		//	fmt.Printf("%+v\n", value)
		//	fmt.Println(value, reflect.TypeOf(value))
		//}
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

// TODO junk ///////////////////////////////////////////////////////////////////////////////////////////////////////////

//
