package main

import ("os"
	"fmt"
	"encoding/csv"
)

func readConf(item string) string {
	f, err := os.Open("enclave.conf");
	if err != nil {
		fmt.Printf("Failed to open enclave.conf\n");
		os.Exit(1)
	}
	defer f.Close()
	lines, _ := csv.NewReader(f).ReadAll();
	for _, each := range lines {
		if each[0] == item {
			return each[1];
		}
	}
	return "NONE";
}

func main() {
	adminBind := readConf("adminBind");
	fmt.Printf("Admin bind: %s\n", adminBind);
}
