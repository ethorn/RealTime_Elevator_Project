package cabBackup

import (
	"fmt"
	"os"
	"encoding/json"
)

var file *os.File
var err error

func Init(id string) {
	file, err = os.OpenFile("cabId"+id+".json", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println("Error when trying to open the cab backup file: ", err)
	}
}

func ReadOrdersFromBackupFile(id string) [4][3]bool {
	genericCabList := [4][3]bool{{false, false, false}, {false, false, false}, {false, false, false}, {false, false, false}} //TODO should be replaced with configvalues imo

	buf := make([]byte, 81)

	var n int
	n, err = file.ReadAt(buf, 0)
	if err != nil {
		fmt.Println("Cab file reading error: ", err) //OBS er ikke farlig at det kommer EOF-melding, brukte 6 t av livet mitt på å oppdage dette
	}
	//fmt.Println("This is what n looks like: ", n) //Nyttig for debugging purposes	

	err := json.Unmarshal(buf[:n], &genericCabList)
	if err != nil {
		fmt.Println("Writing to JSON on Unmarshal failed: ", err)
	}

	//Nyttig for debugging purposes
	fmt.Println("This is what genericCabList looks like after restoration:", genericCabList,"\n")// This is what buf looks like:", buf[:n], "\n")
	return genericCabList
}


func WriteOrdersToBackupFile(requests [4][3]bool) {
	cabOrdersJSON, err := json.Marshal(requests)
	if err != nil {
		fmt.Println("Cab file writing JSON error: ", err)
	}

	n, err := file.WriteAt(cabOrdersJSON, 0)
	if err != nil {
		fmt.Println("Cab file writing write error: ", err)
	}
	// fmt.Println("the value of n is: ", n) // Nyttig for debugging purposes

	file.Truncate(int64(n))
	if err != nil {
		fmt.Println("Cab file writing truncate error: ", err)
	}
}