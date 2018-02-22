// Reading and writing files are basic tasks needed for
// many Go programs. First we'll look at some examples of
// reading files.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/secsy/goftp"
)

// Reading files requires checking most calls for errors.
// This helper will streamline our error checks below.
func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Configuration struct {
	Server       string `json:"Server"`
	User         string `json:"User"`
	Password     string `json:"Password"`
	Serverpath   string `json:"Serverpath"`
	Hexpath      string `json:"Hexpath"`
	Hexbuildpath string `json:"Hexbuildpath"`
	Binpath      string `json:"Binpath"`
}

var configuration Configuration

func main() {

	dat1, _ := ioutil.ReadFile("conf.json")
	decoder := json.NewDecoder(bytes.NewBufferString(string(dat1)))
	fmt.Println(string(dat1))
	err := decoder.Decode(&configuration)
	//fmt.Println(configuration)

	// Perhaps the most basic file reading task is
	// slurping a file's entire contents into memory.
	dat, err := ioutil.ReadFile(configuration.Hexbuildpath)
	check(err)
	//fmt.Print(string(dat))

	dataStr := string(dat)
	//fmt.Print(dataStr)
	dataStrArray := strings.Split(dataStr, "\r\n")
	before := 0
	for i, str := range dataStrArray {
		if strings.Contains(str, ":106000") {
			before = i
		}
	}
	after := 0
	for i := len(dataStrArray) - 30; i < len(dataStrArray); i++ {
		if strings.Contains(dataStrArray[i], ":04aff00000fe00005f") {
			after = i
		}
	}
	slice := dataStrArray[before : after-1]
	outputData := strings.Join(slice[:], "\r\n")

	d1 := []byte(outputData)
	err = ioutil.WriteFile(configuration.Hexpath, d1, 0644)
	check(err)

	execCmd("hex2bin", configuration.Hexpath)

	//eddig

	config := goftp.Config{
		User:               configuration.User,
		Password:           configuration.Password,
		ConnectionsPerHost: 1,
		Timeout:            10 * time.Second,
		IPv6Lookup:         false,
		ActiveTransfers:    false,
		Logger:             os.Stderr,
	}

	client, err := goftp.DialConfig(config, configuration.Server)
	if err != nil {
		fmt.Println("Error connecting to FTP")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	uploadFile, err := os.Open(configuration.Binpath)
	if err != nil {
		fmt.Println("Error opening file for FTP")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = client.Store(configuration.Serverpath, uploadFile)
	if err != nil {
		fmt.Println("Error FTP'ing File")
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func execCmd(path string, args ...string) {
	fmt.Printf("Running: %q %q\n", path, strings.Join(args, " "))
	cmd := exec.Command(path, args...)
	bs, err := cmd.CombinedOutput()
	fmt.Printf("Output: %s", bs)
	fmt.Printf("Error: %v\n\n", err)
}
