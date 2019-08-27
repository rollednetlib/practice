package main

import ("github.com/julienschmidt/httprouter"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
	"encoding/csv"
	"net/smtp"
	"net/http"
	"math/rand"
	"strings"
	"regexp"
	"time"
	"fmt"
	"log"
	"os"
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

/// Pool for redis connections
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
	publicRouter.ServeFiles("/static/*filepath", http.Dir("static"));
	publicRouter.GET("/", indexGet);
	publicRouter.GET("/signin", SignInGet);
	publicRouter.POST("/signin", SignInPost);
	publicRouter.GET("/signout", SignOutGet);

	publicRouter.GET("/reset", ResetPassGet);
	publicRouter.POST("/reset", ResetPassPost);

	publicRouter.GET("/reset/:onetime", OneTimeGet);
	publicRouter.POST("/reset/:onetime", OneTimePost);


	publicRouter.GET("/register", RegisterGet);

	publicRouter.GET("/profile", ProfileGet);
	http.ListenAndServe(":8000", publicRouter);
}


///// Web Page Routes /////
///Index Page Static
func indexGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "index.html");
}
///

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
		Name:	"email",
		Value:	email,
		Expires:time.Now().Add(900 * time.Second),
	});

	fmt.Println("Logging in ", email)
}
///

///Sign Out page
func SignOutGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.SetCookie(w, &http.Cookie{
		Name:	"email",
		MaxAge: -1,
		Expires: time.Now(),
	});
	http.SetCookie(w, &http.Cookie{
		Name:	"token",
		MaxAge:	-1,
		Expires: time.Now(),
	});
	http.Redirect(w, r, "/", 301);
}

/// Register
func RegisterGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "register.html")
}

func RegisterPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//	email := r.FormValue("username");
//	password := r.FormValue("password");
//	registerUser(email, password);
//
}


/// Reset Page
func ResetPassGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "reset.html");
}

func ResetPassPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm();
	http.ServeFile(w, r, "resetsent.html");
	email := r.FormValue("username");
	sms := r.FormValue("phone");
	host := ReadConf("host");
	if validEmail(email) {
		if checkDB("userDB:", email) {
			resetUrl := genUrl(email);
			fmt.Println(email);
			msg := "ONETIME RESET URL: https://"+ host + "/reset/" + resetUrl + "\n\n";
			if sms == "sms" {
				to := hgetDB("userDB:" + email, "phone");
				fmt.Println("Sending SMS! ", to);
				sendSMS(msg, to);
			} else {
				sendMail(msg, email)
			}
		}
	}
}
///


///Reset URL
func OneTimeGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
//	oneTime := ps.ByName("onetime");
	http.ServeFile(w, r, "onetime.html");
}

func OneTimePost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm();
	password := r.FormValue("password");
	hashword := hashPassword(password);
	oneTime := ps.ByName("onetime");
	fmt.Println("Resting !")
	fmt.Println("onetimeDB:" + oneTime)
	if checkDB("onetimeDB:", oneTime) {
		email := getDB("onetimeDB:" + oneTime);
		fmt.Println("EMAIL: ", email);
		hsetDB("userDB:" + email, "password", hashword);
		http.ServeFile(w, r, "onetimefinish.html");

	} else {
		fmt.Println("False");
	}

}


///Profile Pages
func ProfileGet (w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	email, _ := r.Cookie("email")
	token, _ := r.Cookie("token")
	fmt.Println(email.Value);
	fmt.Println(token.Value);
	if checkSession(email.Value, token.Value) {
		fmt.Println("TRUE");
	} else {

	}
}
///
///// //////



///// Utility Functions /////
/// Bcrypt password hashing
func hashPassword(password string) (string) {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14);
	return string(bytes);
}

func checkPassword(password string, hash string) bool {
//	fmt.Println(password, " pass III hash ", hash)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password));
	return err == nil;
}

///Regex Check email
func validEmail(email string) bool {
        re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
        if !(re.MatchString(email)) {
                return false;
        }
	return true;
}


///String Generator
func genToken(len int) string {
	rand.Seed(time.Now().UnixNano());
	bytes := make([]byte, len);
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25));
	}
	return string(bytes);
}

///SendMail
func sendMail(body string, to string) {
	from := ReadConf("smtpFrom");
	pass := ReadConf("smtpPass");
	smtpServer := ReadConf("smtpServer");
	smtpServerPort := ReadConf("smtpServerPort");
//	smtpServer := readConf("smtpServer");

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Enclave Registration\n\n" +
		body;

	err := smtp.SendMail(smtpServerPort, //SANATIZE
		smtp.PlainAuth("", from, pass, smtpServer), //SANATIZE
		from, []string{to}, []byte(msg)); //SANATIZE

	if err != nil {
		log.Printf("smtp error: %s", err);
		return
	}
	log.Print(" MAILING ");
}

///SendSMS
func sendSMS(msg string, to string) {

	//twilioNumber := "+17622043080";
	twilioNumber := ReadConf("twilioNumber");
	twilioUser := ReadConf("twilioUser");
	twilioToken := ReadConf("twilioToken");
	twilioAccount := ReadConf("twilioAccount");
	body := strings.NewReader(`Body=` + msg + `&From=` + twilioNumber + `&To=` + to)
	req, err := http.NewRequest("POST", "https://api.twilio.com/2010-04-01/Accounts/" + twilioAccount +"/Messages.json", body)
	if err != nil {
		// handle err
	}
	req.SetBasicAuth(twilioUser,twilioToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
}



///Config Reader
func ReadConf(item string) string {
	f, err := os.Open("enclave.conf");
	if err != nil {
		fmt.Printf("Could not open enclave.conf!\n");
	}
	defer f.Close();

	lines, _ := csv.NewReader(f).ReadAll();
	for _, each := range lines {
		if each[0] == item {
			return each[1];
		}
	}
	return "NONE";
}
///OneTimeUrl
func genUrl(email string) string {
	conn := pool.Get();
	defer conn.Close();
	resetUrl := genToken(64);
//	redis.String(conn.Do("GET", "rehsetDB:" + resetUrl)) != nil
	conn.Do("SETEX", "onetimeDB:" + resetUrl, 3600, email)
	return resetUrl;
}
///// /////




///// User Management /////
/// New Session
func newSession(email string) string {
	conn := pool.Get();
	defer conn.Close();
	sessionToken := genToken(128);
	conn.Do("SETEX", "sessionDB:" + email, 900, sessionToken);
	return sessionToken;
}

/// Refresh session cookies
func refreshSession(email string, sessionToken string) {
	conn := pool.Get();
	defer conn.Close();
	conn.Do("SETEX", "sessionDB:" + email, 900, sessionToken);
}

/// Check sessionToken
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

///Check user credentials
func checkCred(email string, password string) bool {
	passwordHash := hgetDB("userDB:" + email, "password");
	return checkPassword(password, passwordHash);
}

///Register a new User
func addUser(email string, password string) {
	conn := pool.Get();
	defer conn.Close();
	conn.Do("HSET", "userDB:" + email, "password", password);
}
///// /////



///// Redis DB management functions /////
func getDB(target string) string {
	conn := pool.Get();
	defer conn.Close();
	item, _ := redis.String(conn.Do("GET", target));
	return item;
}

/// Get a hash from db
func hgetDB(match string, target string) string {
	conn := pool.Get();
	defer conn.Close();
	item, _ := redis.String(conn.Do("HGET", match, target));
	return item;
}

/// Set an hash from DB
func hsetDB(match string, target string, value string) {
	conn := pool.Get();
	defer conn.Close();
	conn.Do("HSET", match, target, value);
}

///Check if item exists
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
///// /////
