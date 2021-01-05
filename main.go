package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	rootInfo     *os.FileInfo
	rootNotExist bool
	err          error
	files        []string
)

func mkroot(rootPath string) error {
	err := os.Mkdir(rootPath, 0777)
	return err
}

func getRootInfo(rootPath string) (os.FileInfo, error) {
	rootPathInfo, errRootPath := os.Stat(rootPath)
	if errRootPath != nil {
		if os.IsNotExist(err) {
			fmt.Println("Root does not exist.")
			fmt.Println("Creating root")
		}

		return rootPathInfo, errRootPath
	}

	return rootPathInfo, nil
}

func main() {
	root := os.Getenv("HOME") + "/.reserve_emacs_configs"
	// var rootNotExist bool = false
	rootInfo, err := getRootInfo(root)

	if err != nil {
		fmt.Println("Root does not exist. Attempting to create root")
		err = mkroot(root)
	}

	if err != nil {
		fmt.Println("Root does not exist, also failed to create root. Exiting")
		fmt.Println(err)
		return
	}

	rootInfo, err = getRootInfo(root)

	if err != nil {
		fmt.Println("Failed to create root again, exiting")
		return
	}

	fmt.Println("Root information:")
	fmt.Println(rootInfo)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s --current|--list|--set=[config]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "--list provides currently available configs")
		fmt.Fprintf(os.Stderr, "--current prints currently enabled config")
		fmt.Fprintf(os.Stderr, "--set=[config] enables a desired config from the list")
	}

	/*
		    err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		        files = append(files, path)
		        return nil
		    })
		    if err != nil {
		        panic(err)
		    }
		    for _, file := range files {
		        fmt.Println(file)
				}
	*/
}
