package main

import ("github.com/julienschmidt/httprouter"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"math/rand"
	"strings"
	"time"
	"fmt"
)

var pool * redis.Pool

type userEntry struct {
	status string;
	email string;
	phone string;
	password string;
	userID string;
	userSeed string;
}

func init(){
	pool = &redis.Pool {
		MaxIdle: 40,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "localhost:6379");
			return conn, err;
		},
	}
}

func main() {
	publicRouter := httprouter.New();
	publicRouter.GET("/signin", SignInGet);
	publicRouter.POST("/signin", SignInPost);
	http.ListenAndServe(":8000", publicRouter);
}


///// Web Page Routes /////
///Sign in Pages
func SignInGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "signin.html");
}

func SignInPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm();
	email := r.FormValue("username")
	password := r.FormValue("password")
	if !checkCred(email, password) {
		fmt.Fprintf(w, "FAILED!");
	}
	token := newSession(email);
	tokenCookie := http.Cookie{
		Name:	"token",
		Value:	token,
		Expires:time.Now().Add(900 * time.Second),
	};
	http.SetCookie(w, &tokenCookie);
	http.SetCookie(w, &http.Cookie{
		Name:	"id",
		Value:	email,
		Expires:time.Now().Add(900 * time.Second),
	});

	fmt.Println("Logging in ", email)
}

///

///// Utility Functions /////
// Bcrypt password hashing
func hashPassword(password string) (string) {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14);
	return string(bytes);
}

func checkPassword(password string, hash string) bool {
//	fmt.Println(password, " pass III hash ", hash)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password));
	return err == nil;
}

//String Generator
func genToken(len int) string {
	rand.Seed(time.Now().UnixNano());
	bytes := make([]byte, len);
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25));
	}
	return string(bytes);
}
// // //


///// User Management /////
// New Session
func newSession(email string) string {
	conn := pool.Get();
	defer conn.Close();
	sessionToken := genToken(128);
	conn.Do("SETEX", "sessionDB:" + email, 900, sessionToken);
	return sessionToken;
}

func refreshSession(email string, sessionToken string) {
	conn := pool.Get();
	defer conn.Close();
	conn.Do("SETEX", "sessionDB:" + email, 900, sessionToken);
}

func checkSession(email string, sessionToken string) bool {
	conn := pool.Get();
	defer conn.Close();
	token, err := redis.String(conn.Do("GET", "sessionDB:" + email));
	if err != nil {
		return false;
	}
	if sessionToken == token {
		return true;
	}
	return false;
}

//Check user credentials
func checkCred(email string, password string) bool {
	passwordHash := getDB("userDB:" + email, "passwordHash");
	return checkPassword(password, passwordHash);
}

//Register a new User
func addUser(email string, password string) {
	conn := pool.Get();
	defer conn.Close();
	conn.Do("HSET", "userDB:" + email, "password", password);
}


/// Redis DB management functions ///
// Get an item from db
func getDB(match string, target string) string {
	conn := pool.Get();
	defer conn.Close();
	item, _ := redis.String(conn.Do("HGET", match, target));
	return item;
}

// Set an item from DB
func setDB(match string, target string, value string) {
	conn := pool.Get();
	defer conn.Close();
	conn.Do("HSET", match, target, value);
}

//Check if item exists
func checkDB(match string, target string) bool {
	conn := pool.Get();
	defer conn.Close();
	iter := 0

	// this will store the keys of each iteration
	var keys []string
	for {

	// we scan with our iter offset, starting at 0
	if arr, err := redis.MultiBulk(conn.Do("SCAN", iter, "match", match + "*")); err != nil {
		panic(err)
	} else {

	        // now we get the iter and the keys from the multi-bulk reply
		iter, _ = redis.Int(arr[0], nil)
		keys, _ = redis.Strings(arr[1], nil)
	}

	if iter == 0  {
		break
		}
	}
	for _, element := range keys {
		if strings.Contains(element, target) {
			return true;
		}
	}
	return false;
}
