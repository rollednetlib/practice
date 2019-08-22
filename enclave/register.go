package main

import (
    "fmt"
    "strings"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "net/smtp"
    "html/template"
    "log"
    "os"
    "bufio"
    "math/rand"
    "time"
)

func genSessionString(len int) string {
	rand.Seed(time.Now().UnixNano());
	bytes := make([]byte, len);
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25));
	}
	return string(bytes);
}

func genSessionID(len int) string {
	sessionString := genSessionString(len);
	f, err := os.OpenFile("sessionjar", os.O_CREATE, 0660);
	if err != nil {
		panic(err);
	}
	defer f.Close();

	scanner := bufio.NewScanner(f);
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), sessionString) {
			sessionString := genSessionString(len);
			fmt.Println(sessionString);
		}
	}
	jar, err := os.OpenFile("sessionjar", os.O_RDWR|os.O_APPEND, 0660);
	jar.WriteString(sessionString + "\n");
	defer jar.Close();
	return sessionString;
}


func sendMail(body string, to string) {
//	from := "bot@example.com"	
//	pass := "Access Token"	

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Aquinas Password\n\n" +
		body

//	err := smtp.SendMail("smtp.example.com:587", 
//		smtp.PlainAuth("", from, pass, "smtp.example.com"), 
//		from, []string{to}, []byte(msg)) 

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
	log.Print(" MAILING ")
}

func registerUser(email string) {
	userID := genSessionID(128);
	confirmationToken := genSessionID(64);
	if (checkFile("userRegisterJar", email) != 0) {
		//Check if user has already undergone registration
		fmt.Println("REG FAILED, ALREADY REGISTERED\n");
	} else {
		//
		
		sendMail(registrationMsg, email);
	}
}

func checkFile(file string, target string) int {
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
			fmt.Println("EMAIL DETECTED\n");
			return 0;
		} else {
			fmt.Println("NOT DETECTED\n");
			jar, _ := os.OpenFile(file, os.O_RDWR|os.O_APPEND, 0660);
			jar.WriteString(target + "\n");
			return 1;
		}
	}
	return 0;
}


func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprint(w, "Welcome!\n");
}

func registerConfirmation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    fmt.Fprintf(w, "Checking token, %s!\n", ps.ByName("confirmationToken"));
}


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
		fmt.Fprintf(w, "Your mailing address is not approved!\n");
	}
	//	email.Trim("]")
}

func main() {
    router := httprouter.New()
 //   router.ServeFiles("/", http.Dir("/"));
    router.GET("/confirmation/:confirmationToken", registrationConfirmation);
  //  router.GET("/reset", Reset);
  //  router.GET("/signin", Signin);
    router.GET("/register", RegisterGet);
    router.POST("/registerp", RegisterPost);

    log.Fatal(http.ListenAndServe(":8000", router))
}

