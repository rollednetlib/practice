package main

import (
    "fmt"
    "strings"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "net/smtp"
    "html/template"
    "encoding/csv"
    "log"
    "os"
    "bufio"
    "math/rand"
    "time"
    "crypto/sha512"
    "encoding/base64"
    "regexp"
)

func genToken(len int) string {
	rand.Seed(time.Now().UnixNano());
	bytes := make([]byte, len);
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25));
	}
	return string(bytes);
}

func genTokenFile(len int, file string) string {
	sessionString := genToken(len);
	f, err := os.OpenFile(file , os.O_CREATE, 0660);
	if err != nil {
		panic(err);
	}
	defer f.Close();

	scanner := bufio.NewScanner(f);
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), sessionString) {
			sessionString := genToken(len);
			fmt.Println(sessionString);
		}
	}
	jar, err := os.OpenFile(file, os.O_RDWR|os.O_APPEND, 0660);
	jar.WriteString(sessionString + "\n");
	defer jar.Close();
	return sessionString;
}

func hashPassword(password string) string {
	hash := sha512.Sum512([]byte(password));
	return base64.StdEncoding.EncodeToString(hash[0:]);
}

func sendMail(body string, to string) {
	from := readConf("smtpFrom");
	pass := readConf("smtpPass");
	smtpServer := readConf("smtpServer");

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Enclave Registration\n\n" +
		body;

	err := smtp.SendMail("smtp.gmail.com:587", //SANATIZE
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"), //SANATIZE
		from, []string{to}, []byte(msg)); //SANATIZE

	if err != nil {
		log.Printf("smtp error: %s", err);
		return
	}
	log.Print(" MAILING ");
}

func registerUser(email string) {
	host := readConf("host");

	if (checkFile("userRegisterJar", email, "none") != 0) {
		//Check if user has already undergone registration
		fmt.Println("REG FAILED, ALREADY REGISTERED\n");
	} else {
		confirmationToken := genToken(64);
		confString := confirmationToken + "," + email;
		checkFile("DBpendingReg", confString, "append");
	registrationMsg := "To complete registration and begin working on projects, click the link or login using the temporary credentials\n\n" + "https://" + host + "/registrationConfirmation/?token=" + confirmationToken;
		//		
		sendMail(registrationMsg, email);
		userID := genTokenFile(128, "DBreg");
		fmt.Printf("UserID: %s\n\n", userID);
	}
}

func resetUser(email string) {
	resetToken := genToken(32)
	resetMsg := "To reset your enclave account, click the link\n" + "https://127.0.0.1:8000/resetUser?token" + resetToken + "\n";
	sendMail(resetMsg, email);
}

func checkFile(file string, target string, flag string) int {
	f, err :=  os.OpenFile(file, os.O_CREATE, 0660);
	if err != nil {
		panic(err);
	}
	defer f.Close();
	scanner :=  bufio.NewScanner(f);
	fmt.Println("SCANNING\n\n");
	for scanner.Scan() {
		fmt.Println("SCAN LINE\n");
		if strings.Contains(scanner.Text(), target) {
			fmt.Println("DETECTED\n");
			if flag == "delete" {

			}
			return 0;
		} else {
			fmt.Println("NOT DETECTED\n");
			if flag == "append" {
				jar, _ := os.OpenFile(file, os.O_RDWR|os.O_APPEND, 0660);
				jar.WriteString(target + "\n");
			}
			return 1;
		}
	}
	return 0;
}


func IndexGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "home.html");
}

func ResetGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token := ps.ByName("confirmationToken");
    if (checkFile("tokenJar", token, "none") != 0 ){
	http.ServeFile(w, r, "reset.html");
    } else {
	    //Prompt user for new password and remove token from list
	    t, _ := template.ParseFiles("reset.tmpl");
	    t.Execute(w, nil);
    }
}

//func ResetPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

//}

//func Landing(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//	data, _ := ioutil.ReadFile("home.html");
//	fmt.Fprintf(w, data);
	//	w.Write([]byte("home.html"))
	//fs := http.File("landing.html");
	//http.Handle("");
//}

func RegisterGet( w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, _ := template.ParseFiles("register.tmpl");
	t.Execute(w, nil);
}

func RegisterPost( w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm();
	email := strings.Join(r.Form["email"],"")
	if ( !strings.Contains(email, "@" ) ) {
		fmt.Fprintf(w, "Failed to register, use a mail.mil address\n", email);
	}  else if ( strings.Contains( email, "@mail.mil") ) {
		fmt.Fprintf(w, "A registration email has been sent! It will be cancelled if not completed with 24 hours\n");
		registerUser(email);
//		sendMail("Register TEMP\n\n24 Hours to register\n", email);
//		registerEmail(email);
	}  else if ( strings.Contains( email, "@fmofs.com") ) {
		fmt.Fprintf(w, "A registration email has been sent! It will be cancelled if not completed with 24 hours\n");
		registerUser(email);
//		sendMail("Register TEMP\n\n24 Hours to register\n", email);
	} else {
		fmt.Fprintf(w, "Your mailing address is not valid!\nOnly mail.mil addreses are approved!\n");
	}
	//	email.Trim("]")
}

//func ResetGet( w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//	http.ServeFile(w, r, "reset.html");
//}


func ResetPost (w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm();
	email := strings.Join(r.Form["email"], "");
	status := validEmail(email)
	if status == "registered" {
		http.ServeFile(w, r, "resetSent.html");
		resetUser(email);
	}
	if status == "unregistered" {
		http.ServeFile(w, r, "resetFailedUnReg.html");
	}
	if status == "pendingregistration" {
		http.ServeFile(w, r, "resetFailedPendingReg.html");
	}
	if status == "invalid" {
		http.ServeFile(w, r, "invalidEmail.html");
	}

}

func validEmail(email string) string {
	validdom := readConf("validDom");
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !(re.MatchString(email)) {
		return "invalid"
	}
	if (!strings.Contains( email, validdom)) {
		return "invalid"
	}
	if checkFile("DBpendingReg", email , "none") == 0 {
		return "pendingregistration"
	}
	if checkFile("DBreg", email, "none") == 0 {
		return "registered"
	}
	if checkFile("FBreg", email,"none") != 0 {
		return "unregisted"
	}
	return "invalid"
}

func SigninPost (w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm();
	username := strings.Join(r.Form["username"], "");
	password := strings.Join(r.Form["password"], "");
	fmt.Printf(username, password);

}

func SigninGet (w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile
}

func readConf(item string) string {
	f, err := os.Open("enclave.conf");
	if err != nil {
		fmt.Printf("Could not open enclave.conf\n");
		os.Exit(1);
	}
	defer f.Close();

	lines, _ := csv.NewReader(f).ReadAll();
	for _, each := range lines {
		if each[0] == item {
			return each[1];
		}
	}

//	if adminBind == publicBind {
//		log.Fatal("admin and public ports can not be on the same page!");
//		os.Exit(1);
//	}
	return "NONE";
}

func main() {
	publicBind := readConf("publicBind");


	publicRouter := httprouter.New();

	publicRouter.GET("/", IndexGet);
	publicRouter.GET("/reset", ResetGet);
	publicRouter.POST("/resetp", ResetPost);
//      router.POST("/test", testFunc);
//    router.GET("/reset", Reset);

//    router.GET("/", indexGet);

	publicRouter.GET("/signin", SigninGet);
	publicRouter.POST("/signinp", SigninPost);

//	publicRouter.GET("/console", UserConsoleGet);
//    router.POST("/consolep", UserConsolePost);

//    router.GET("/profile", UserProfileGet);
//    router.POST("/profilep", UserProfilePost);

	publicRouter.GET("/register", RegisterGet);
	publicRouter.POST("/registerp", RegisterPost);

//	adminRouter := httprouter.New();
//	adminRouter.GET("/admin", admin);

	log.Fatal(http.ListenAndServe(publicBind, publicRouter))
//	log.Fatal(http.ListenAndServe(adminBind, adminRouter))
}

