package main

import (
	"fmt"
	"io/ioutil"
	"time"
)

func main() {
	//dir := os.Args[1]
	now := time.Now()
	fmt.Println(now)
	listAll("C:\\", 0)
	after := time.Now()
	fmt.Println(after)
}

func listAll(path string, curHier int) {
	readerInfos, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, info := range readerInfos {
		if info.IsDir() {
			for tmpheir := curHier; tmpheir > 0; tmpheir-- {
				//fmt.Printf("|\t")
			}
			//fmt.Println(info.Name(), "\\")
			listAll(path+"\\"+info.Name(), curHier+1)
		} else {
			for tmpheir := curHier; tmpheir > 0; tmpheir-- {
				//fmt.Printf("|\t")
			}
			//fmt.Println(info.Name())
		}
	}
}
