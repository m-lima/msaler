package main

import (
	"fmt"
	"os"
)

func main() {
	var err error
	if len(os.Args) == 1 {
		err = GetClientToken(os.Args[1:])
	} else {
		modeParam := os.Args[1]

		if "token" == modeParam {
			err = GetClientToken(os.Args[2:])
		} else if "new" == modeParam {
			err = NewClient(os.Args[2:])
		} else if "delete" == modeParam {
			err = DeleteClient(os.Args[2:])
		} else if "print" == modeParam {
			err = PrintClient(os.Args[2:])
		} else if "config" == modeParam {
			var path string
			path, err = ConfigPath()
			if err == nil {
				fmt.Println(path)
			}
		} else {
			err = GetClientToken(os.Args[1:])
			if err != nil {
				err = fmt.Errorf("Error while trying to interpret command as user: %v", err)
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
}
