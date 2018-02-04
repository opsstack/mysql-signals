package main

// TODO
// Later - option to not to delta processing in tool, just return counts

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/ogier/pflag"
	"log"
	"os"
	"strconv"
	"time"
)

// constants
const (
	version   = "0.1"
	copyright = "Copyright 2018 by OpsStack"
)

// Global vars
var (
	argServerServer   string
	argServerPort     string
	argServerUser     string
	argServerPassword string
	argStatsMetric    string
	argStatusFileName string
	argCredFileName   string
	flagVerbose       bool
	flagVeryVerbose   bool
	flagHelp          bool
)

// init is called automatically at start
func init() {

	// Setup arguments, must do before calling Parse()
	flag.StringVarP(&argServerServer, "server", "S", "127.0.0.1", "Server Host")
	flag.StringVarP(&argServerPort, "port", "P", "3306", "Server Port")
	flag.StringVarP(&argServerUser, "user", "u", "", "User")
	flag.StringVarP(&argServerPassword, "password", "p", "", "Password")
	flag.StringVarP(&argStatsMetric, "metric", "m", "c", "Metric Type")
	flag.StringVarP(&argStatusFileName, "statusfile", "f", "", "Status File")
	flag.StringVarP(&argCredFileName, "credfile", "c", "", "Credential File")
	flag.BoolVarP(&flagVerbose, "verbose", "v", false, "Verbose Output")
	flag.BoolVarP(&flagVeryVerbose, "very-verbose", "w", false, "Very Verbose Output")
	flag.BoolVarP(&flagHelp, "help", "h", false, "Help")

	flag.Parse() // Process argurments
}

func main() {

	var (
		err          error
		infoFileName string
		lastRunInfo  [4]int
		lastRunCount int
		lastRunTime  int
		deltaTime    int
		nowTime      int
		deltaCount   int
		rows         *sql.Rows
		query        string
		query2       string
		result       string
		rateCount    int
		errorCount   int
		latency      float64
		queryRate    float64
		errorRate    float64
	)

	startTime := time.Now()

	if flagVerbose {
		fmt.Println("")
		fmt.Printf("MySQLMetrics Version %s - %s\n", version, copyright)
		fmt.Printf("Starting at: %s \n", startTime.Format(time.UnixDate))
		if argServerPassword == "" {
			fmt.Printf("Arguments: %s \n\n", os.Args[1:]) // Skip program name
		} else {
			fmt.Printf("Arguments: Can't show \n\n") // Now show if PW here
		}
	}

	// Check our command-line arguments
	argsCheck(version, copyright)

	// Get our last run counters from status file
	infoFileName = argStatusFileName // Need more safety checks before open argument filename?
	lastRunTime, lastRunCount, lastRunInfo, err = getLastRunInfo(infoFileName)

	dbHost := argServerServer
	dbPort := argServerPort
	dbUser := argServerUser
	dbPassword := argServerPassword
	dbDatabase := ""
	dbCharSet := "utf8" //Not sure if need, but set and use so it's clear

	// Prompt for or read in user and password if not on command line
	if argCredFileName == "" {
		// No cred file so prompt
		if dbUser == "" {
			fmt.Print("Enter DB user: ")
			var input string
			fmt.Scanln(&input)
			dbUser = input
			fmt.Println(dbUser)
		}
		if dbPassword == "" {
			fmt.Print("Enter password: ")
			var input string
			fmt.Scanln(&input)
			dbPassword = input
			fmt.Println(dbPassword)
		}
	} else { // Given file name, so read it; user one first line, password on second line
		credFileName := argCredFileName
		f, err := os.Open(credFileName)

		if err == nil {
			scanner := bufio.NewScanner(f)
			scanner.Scan()
			dbUser = scanner.Text()
			scanner.Scan()
			dbPassword = scanner.Text()
		} else { //  Have error
			log.Fatalln("Error reading credential file.")
			os.Exit(1)
		}
		f.Close()
	}

	// Connect to the database and run queries

	// Format: user:password@tcp(localhost:5555)/dbname?charset=utf8
	dataSourceName := dbUser + ":" + dbPassword + "@" + "tcp(" + dbHost + ":" + dbPort +
		")/" + dbDatabase + "?charset=" + dbCharSet
	db, err := sql.Open("mysql", dataSourceName)
	checkErr(err)

	// Get MySQL Info
	query = "SELECT @@VERSION;"
	rows, err = db.Query(query)
	checkErr(err)
	rows.Next()
	err = rows.Scan(&result)
	checkErr(err)
	if flagVerbose {
		fmt.Printf("DB Version: %s\n", result)
	}

	// Get version info, as 5.7 behaves differently
	baseVersion, err := strconv.ParseFloat(result[:3], 64)
	checkErr(err)
	// Do any version issues here if needed
	switch baseVersion {
	}

	switch argStatsMetric {
	case "r":
		var tableName string
		if baseVersion < 5.7 { // Need to query performance schema in 5.7+
			tableName = "INFORMATION_SCHEMA.GLOBAL_STATUS"
		} else {
			tableName = "performance_schema.global_status"
		}
		query = "SELECT sum(variable_value) " + "FROM " + tableName +
			" WHERE variable_name IN ('com_select', 'com_u pdate', 'com_delete', 'com_insert', 'qcache_hits') ;"

	case "e":
		query = "SELECT sum(sum_errors) " +
			"FROM performance_schema.events_statements_summary_by_user_by_event_name " +
			"WHERE event_name IN ('statement/sql/select', 'statement/sql/insert', 'statement/sql/update', 'statement/sql/delete');"

	case "l":
		query = "SELECT (avg_timer_wait)/1e9 AS avg_latency_ms " +
			"FROM performance_schema.events_statements_summary_global_by_event_name " +
			"WHERE event_name = 'statement/sql/select';"
		query2 = "TRUNCATE TABLE performance_schema.events_statements_summary_global_by_event_name ;"
	}

	rows, err = db.Query(query)
	checkErr(err)
	rows.Next()
	err = rows.Scan(&result)
	checkErr(err)
	if flagVerbose {
		fmt.Printf("DB Result: %s\n", result)
	}

	// See if we have a 2nd query for truncating; note we discard results
	if query2 != "" {
		_, err = db.Query(query2)
		checkErr(err)
	}

	nowTime = int(time.Now().Unix())

	switch argStatsMetric {
	case "r":
		rateCount, err = strconv.Atoi(result)
		checkErr(err)
		deltaCount = rateCount - lastRunCount
		deltaTime = nowTime - lastRunTime
		queryRate = float64(deltaCount / deltaTime)
		lastRunCount = rateCount // Update last count
	case "e":
		errorCount, err = strconv.Atoi(result)
		checkErr(err)
		deltaCount = errorCount - lastRunCount
		deltaTime = nowTime - lastRunTime
		errorRate = float64(deltaCount / deltaTime)
		lastRunCount = errorCount // Update last count
	case "l":
		latency, err = strconv.ParseFloat(result, 64)
		checkErr(err)
	}
	lastRunTime = nowTime // Update last time

	db.Close()

	err = saveLastRunInfo(infoFileName, lastRunTime, lastRunCount, lastRunInfo)
	checkErr(err)

	endTime := time.Now()
	duration := int(endTime.Sub(startTime).Nanoseconds() / 1000000)

	// Output
	if flagVerbose {
		fmt.Printf("Run duration: %d ms\n", duration)
	}

	switch argStatsMetric {
	case "r":
		if flagVerbose {
			fmt.Printf("Query rate: ")
		}
		fmt.Printf("%f", queryRate)
		if flagVerbose {
			fmt.Printf("/sec\n")
		} else {
			fmt.Printf("\n")
		}

	case "e":
		if flagVerbose {
			fmt.Print("Error rate: ")
		}
		fmt.Printf("%f\n", errorRate)

	case "l":
		if flagVerbose {
			fmt.Printf("Avg Response Time: ")
		}
		fmt.Printf("%f", latency)
		if flagVerbose {
			fmt.Printf(" ms\n")
		} else {
			fmt.Printf("\n")
		}
	default:
		panic("Invdalid Stats Metric in main")
	} // Switch on argStatsMetric

	runTime := time.Now().Sub(startTime) / time.Millisecond
	if flagVerbose {
		fmt.Printf("\n")
		fmt.Printf("Total exec time (ms): %d \n\n", runTime)
	}

	// Exit normally
	os.Exit(0)
} // Main

// Process arguments
func argsCheck(version string, copyright string) {

	if flagHelp {
		fmt.Printf("GoldenWebReader Version %s - %s\n\n", version, copyright)
		fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Require metric type
	if argStatsMetric != "r" && argStatsMetric != "e" && argStatsMetric != "l" {
		log.Fatalln("Stats Metric not valid - should be r, e, l")
		os.Exit(1)
	}

	// Can have cred file OR user/password
	if argCredFileName != "" && (argServerUser != "" || argServerPassword != "") {
		log.Fatalln("Cannot supply BOTH a credential file and a user or password.")
		os.Exit(1)
	}
}

// Error checking for various things
func checkErr(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}
