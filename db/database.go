package db;

const 	DATA_LOC	string			= 	"~/.dddata"
const 	SWITCH_TABLE	string		= 	"switches"
const 	SERVER_TABLE	string		= 	"servers"
const	TEST_TABLE		string		= 	"testnets"
const	NODES_TABLE		string		= 	"nodes"

func dbInit(){
	db := getDB()
	defer db.Close()

	switchTable := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s,%s);",
		SWITCH_TABLE,
		"id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT",
		"addr TEXT NOT NULL",
		"iface TEXT NOT NULL",
		"brand INTEGER NOT NULL")

	serverTable := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s,%s, %s,%s,%s, %s);",
		SERVER_TABLE,
		
		"id INTERGER NOT NULL PRIMARY KEY AUTOINCREMENT",
		"addr TEXT NOT NULL",
		"iaddr_ip TEXT NOT NULL",

		"iaddr_gateway TEXT NOT NULL",
		"iaddr_subnet INTERGER NOT NULL",
		"nodes INTEGER NOT NULL DEFAULT 0",

		"max INTEGER NOT NULL",
		"iface TEXT NOT NULL",
		"switch INTEGER NOT NULL",
		"name TEXT")

	testTable := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s,%s);",
		TEST_TABLE,
		"id INTERGER NOT NULL PRIMARY KEY AUTOINCREMENT",
		"blockchain TEXT NOT NULL",
		"nodes INTERGER NOT NULL",
		"image TEXT NOT NULL")

	nodesTable := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s);",
		NODES_TABLE,

		"id INTERGER NOT NULL PRIMARY KEY AUTOINCREMENT",
		"test_net INTERGER NOT NULL",
		"server INTERGER NOT NULL",

		"local_id INTERGER NOT NULL",
		"ip TEXT NOT NULL")
	db.Exec(switchTable)
	db.Exec(serverTable)
	db.Exec(testTable)
	db.Exec(nodesTable)

}

func getDB() *sql.DB {
	if _, err := os.Stat(DATA_LOC); os.IsNotExist(err) {
	  	dbInit()
	}
	db, err := sql.Open("sqlite3", DATA_LOC)
	checkFatal(err)
	return db;
}

