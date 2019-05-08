package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// App provides references to the router and the database.
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize is responsible for creating a database connection and wiring up the routes
func (a *App) Initialize(user, password, instanceName, databaseName string) {
	var err error
	a.DB, err = configureCloudSQL(cloudSQLConfig{
		Username: user,
		Password: password,
		// The connection name of the Cloud SQL v2 instance, i.e.,
		// "project:region:instance-id"
		// Cloud SQL v1 instances are not supported.
		Instance: instanceName,
		Database: databaseName,
	})
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

// Run starts the application
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	u := user{ID: id}
	if err := u.getUser(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	users, err := getUsers(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var u user
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := u.createUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, u)
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var u user
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	u.ID = id

	if err := u.updateUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid User ID")
		return
	}

	u := user{ID: id}
	if err := u.deleteUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/users", a.getUsers).Methods("GET")
	a.Router.HandleFunc("/user", a.createUser).Methods("POST")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.getUser).Methods("GET")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.updateUser).Methods("PUT")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.deleteUser).Methods("DELETE")
}

type cloudSQLConfig struct {
	Username, Password, Instance, Database string
}

func configureCloudSQL(config cloudSQLConfig) (*sql.DB, error) {
	// Use one of the environment variables automatically added to the running
	// containers to detect whether the app runs on Google Cloud Run or locally.
	// See https://cloud.google.com/run/docs/reference/container-contract#env-vars
	// for the list of environment variables exposed by Google Cloud Run.
	if os.Getenv("K_SERVICE") != "" {
		// Running in production on Google Cloud Run.
		return newMySQLDB(MySQLConfig{
			Username:     config.Username,
			Password:     config.Password,
			UnixSocket:   "/cloudsql/" + config.Instance,
			DatabaseName: config.Database,
		})
	}

	// Running locally for development.
	return newMySQLDB(MySQLConfig{
		Username:     config.Username,
		Password:     config.Password,
		Host:         "localhost",
		Port:         3306,
		DatabaseName: config.Database,
	})
}

// MySQLConfig stores MySQL connection details.
type MySQLConfig struct {
	// Optional.
	Username, Password string

	// Host of the MySQL instance.
	//
	// If set, UnixSocket should be unset.
	Host string

	// Port of the MySQL instance.
	//
	// If set, UnixSocket should be unset.
	Port int

	// UnixSocket is the filepath to a unix socket.
	//
	// If set, Host and Port should be unset.
	UnixSocket string

	// Database name.
	DatabaseName string
}

// connectionString returns a connection string suitable for sql.Open.
func (c MySQLConfig) connectionString() string {
	var cred string
	// [username[:password]@]
	if c.Username != "" {
		cred = c.Username
		if c.Password != "" {
			cred = cred + ":" + c.Password
		}
		cred = cred + "@"
	}

	if c.UnixSocket != "" {
		return fmt.Sprintf("%sunix(%s)/%s", cred, c.UnixSocket, c.DatabaseName)
	}
	return fmt.Sprintf("%stcp([%s]:%d)/%s", cred, c.Host, c.Port, c.DatabaseName)
}

// newMySQLDB creates a new BookDatabase backed by a given MySQL server.
func newMySQLDB(config MySQLConfig) (*sql.DB, error) {
	// Check database and table exists. If not, create it.
	// TODO
	// if err := config.ensureTableExists(); err != nil {
	// 	return nil, err
	// }

	conn, err := sql.Open("mysql", config.connectionString())
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("mysql: could not establish a good connection: %v", err)
	}

	db := conn

	// Prepared statements. The actual SQL queries are in the code near the
	// relevant method (e.g. addBook).
	// TODO
	// if db.list, err = conn.Prepare(listStatement); err != nil {
	// 	return nil, fmt.Errorf("mysql: prepare list: %v", err)
	// }
	// if db.listBy, err = conn.Prepare(listByStatement); err != nil {
	// 	return nil, fmt.Errorf("mysql: prepare listBy: %v", err)
	// }
	// if db.get, err = conn.Prepare(getStatement); err != nil {
	// 	return nil, fmt.Errorf("mysql: prepare get: %v", err)
	// }
	// if db.insert, err = conn.Prepare(insertStatement); err != nil {
	// 	return nil, fmt.Errorf("mysql: prepare insert: %v", err)
	// }
	// if db.update, err = conn.Prepare(updateStatement); err != nil {
	// 	return nil, fmt.Errorf("mysql: prepare update: %v", err)
	// }
	// if db.delete, err = conn.Prepare(deleteStatement); err != nil {
	// 	return nil, fmt.Errorf("mysql: prepare delete: %v", err)
	// }

	return db, nil
}
