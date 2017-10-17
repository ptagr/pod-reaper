package main

import (
	"fmt"
	"time"
)

func main(){
	for true {
		fmt.Printf("\n%s Hello from pod reaper! Hide all the pods!", time.Now().Format(time.UnixDate))
		time.Sleep(30 * time.Second)
	}

}