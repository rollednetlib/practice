package main

import ("fmt"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
	"strings"
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
	userDB := "userDB:"
	email := "testmail@mail.org"
	password := "TestPass"
	if checkDB("userDB:",email) {
		fmt.Println("FOUND!");
	} else {
		fmt.Println("NOT FOUND!\nAdding to DB.");
		addUser(email, "OneTimePassasdasd");
	}
	setDB(userDB + email, "passwordHash","$2a$14$mKidP0mTs8P7Y.NC7BLofO9SkTIBc9Y8GhKNviVyj1rotU7uftxaK");
	if checkCred(email, password) {
		fmt.Println("PASSED!");
	} else {
		fmt.Println("FAILED");
	}

}

// Bcrypt password hashing
func hashPassword(password string) (string) {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14);
	return string(bytes);
}

func checkPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password));
	return err == nil;
}
//

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


//Redis DB management functions
func getDB(match string, target string) string {
	conn := pool.Get();
	defer conn.Close();
	item, _ := redis.String(conn.Do("HGET", match, target));
	return item;
}

func setDB(match string, target string, value string) {
	conn := pool.Get();
	defer conn.Close();
	conn.Do("HSET", match, target, value);
}

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

//	fmt.Println(keys)

	// check if we need to stop...
	if iter == 0  {
		break
		}
	}
	for _, element := range keys {
	//	fmt.Println(element);
		if strings.Contains(element, target) {
			return true;
		}
	}
	return false;
}
