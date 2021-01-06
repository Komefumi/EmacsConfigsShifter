package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/otiai10/copy"

	fiutils "github.com/Komefumi/EmacsConfigShifter/filesutils"
)

var (
	rootInfo     *os.FileInfo
	rootNotExist bool
	err          error
	files        []string
	rootPath string = os.Getenv("HOME") + "/.reserve_emacs_configs"
	swapPath string = path.Join(rootPath, ".swap_for_current")
	errNoConfigUsed error = errors.New("No Config Is Currently Used")
	errNoEmacsConfigExists error = errors.New("No Emacs Configuration currently exists")
)

func exitGracefully(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func check(e error) {
	if e != nil {
		exitGracefully(e)
	}
}

func mkroot() error {
	err := os.Mkdir(rootPath, 0777)
	return err
}

func getRootInfo() (os.FileInfo, error) {
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

func attemptToCreateSwapPath() error {
	f, err := os.Create(swapPath)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func locExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return false, err
}

func ensureDirectory() {
	exists, err := locExists(rootPath)
	check(err)
	if !exists {
		err = mkroot()
	}
	check(err)
	exists, err = locExists(swapPath)
	check(err)
	if !exists {
		err = attemptToCreateSwapPath()
	}
	check(err)
}

func getAvailableConfigs() []os.FileInfo {
	files, err := ioutil.ReadDir(rootPath)
	var dirs []os.FileInfo
	if err != nil {
		exitGracefully(err)
	}
	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, f)
		}
	}

	return dirs
}

func printAvailableConfigs() {
	configDirs := getAvailableConfigs()
	for _, f := range configDirs {
		fmt.Println(f.Name())
	}
	os.Exit(0)
}

func printCurrentConfig() {
	contents, err := ioutil.ReadFile(swapPath)
	if err != nil {
		exitGracefully(err)
	}
	if len(strings.TrimSpace(string(contents))) == 0 {
		fmt.Println("No config is currently being used")
	} else {
		fmt.Printf("Currently used config is: %s\n", contents)
	}
	os.Exit(0)
}

func setConfigFromAvailable(desiredConfigName string)  {
	configDirs := getAvailableConfigs()
	var desiredConfig os.FileInfo
	for _, f := range configDirs {
		if f.Name() == desiredConfigName {
			desiredConfig = f
		}
	}
	if desiredConfig != nil {
		currentConfigName, err := getCurrentConfig()
		if err == errNoConfigUsed {
			err = enableConfig(desiredConfig.Name())
			check(err)
		} else if err == nil {
			err = archiveAs(currentConfigName)
			check(err)
			err = clearCurrentConfig()
			check(err)
			err = enableConfig(desiredConfig.Name())
			check(err)
		} else {
			check(err)
		}
	} else {
		exitGracefully(errors.New("Configuration specified was not found"))
	}

	os.Exit(0)
}

func getCurrentConfig() (string, error) {
	contents, err := ioutil.ReadFile(swapPath)
	if err != nil {
		return "", err
	}
	if len(strings.TrimSpace(string(contents))) == 0 {
		return "", errNoConfigUsed
	}
	return strings.TrimSpace(string(contents)), nil
}

func clearCurrentConfig() error {
	// os.Truncate()
	err := os.Truncate(swapPath, 0)
	return err
}

func archiveAs(name string) error {
	srcDir := os.Getenv("HOME") + "/" + ".emacs.d"
	destDir := rootPath + "/" + name
	itExists, err := locExists(destDir)
	if err != nil {
		return err
	}
	if itExists {
		err = os.RemoveAll(destDir)
		if err != nil {
			return err
		}
		// err = os.Remove(destDir)
		if err != nil {
			return err
		}
	}
	err = copy.Copy(srcDir, destDir)
	// err = fiutils.CopyDir(srcDir, destDir, true)
	if err != nil {
		// fmt.Println("Here")
		return err
	}
	// err = os.Rename(rootPath + "/" + ".emacs.d", rootPath + "/" + name)
	if err != nil {
		return err
	}
	return nil
}

func enableConfig(configName string) error {
	srcDir := rootPath + "/" + configName
	dstDir := os.Getenv("HOME") + "/" + ".emacs.d"
	err := os.RemoveAll(dstDir)
	if err != nil {
		return err
	}
	err = fiutils.CopyDir(srcDir, os.Getenv("HOME") + "/" + ".emacs.d")
	if err != nil {
		fmt.Println("Error occurs here")
		return err
	}
	f, err := os.OpenFile(swapPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%s", configName)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func performPresentConfigArchiving() {
	presentConfigName, err := getCurrentConfig()
	if err != errNoConfigUsed {
		check(err)
	}
	aConfigExists, err := locExists(os.Getenv("HOME") + "/.emacs.d")
	check(err)
	if !aConfigExists {
		exitGracefully(errNoEmacsConfigExists)
	}
	if presentConfigName == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Preparing to archive present configuration. Please enter a valid name for saving the current config as")
		providedName, err := reader.ReadString('\n')
		check(err)
		providedName = strings.TrimSpace(providedName)
		err = archiveAs(providedName)
		check(err)
		err = clearCurrentConfig()
		check(err)
		fmt.Printf("Present configuration successfully saved as %s at %s\n", providedName, rootPath + "/" + providedName)
	} else {
		err = archiveAs(presentConfigName)
		check(err)
		err = clearCurrentConfig()
		check(err)
	}
	os.Exit(0)
}

func main() {
	// var rootNotExist bool = false
	ensureDirectory()


	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s --current|--list|--set=[config]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "--list provides currently available configs\n")
		fmt.Fprintf(os.Stderr, "--current prints currently enabled config\n")
		fmt.Fprintf(os.Stderr, "--set=[config] enables a desired config from the list\n")
		fmt.Fprintf(os.Stderr, "--archive Archives the present config\n")
	}

	listPtr := flag.Bool("list", false, "List available configs")
	setPtr := flag.String("set", "", "Set a config from available configs")
	currentPtr := flag.Bool("current", false, "Shows the currently used config")
	archivePresentConfig := flag.Bool("archive", false, "Archive the presently used config")

	flag.Parse()

	if *listPtr == true {
		printAvailableConfigs()
	}

	if *setPtr != "" {
		setConfigFromAvailable(*setPtr)
	}

	if *currentPtr == true {
		printCurrentConfig()
	}

	if *archivePresentConfig == true {
		performPresentConfigArchiving()
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
