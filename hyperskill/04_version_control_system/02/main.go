package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type user struct {
	name string
}

type repository struct {
	user          user
	rootDir       string
	repositoryDir string
	configFile    string
	index         []string
	indexFile     string
	commands      map[string]string
}

func newRepository() *repository {
	return &repository{
		commands: map[string]string{
			"config":   "Get and set a username.",
			"add":      "Add a file to the index.",
			"log":      "Show commit logs.",
			"commit":   "Save changes.",
			"checkout": "Restore a file.",
			"--help": `These are SVCS commands:
config     Get and set a username.
add        Add a file to the index.
log        Show commit logs.
commit     Save changes.
checkout   Restore a file.`,
		},
	}
}

const (
	vcsDirName     = "vcs"
	configFileName = "config.txt"
	indexFileName  = "index.txt"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (r *repository) ensurePathExists(path string, isDir bool) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if isDir {
			err = os.MkdirAll(path, os.ModePerm)
		} else {
			_, err = os.Create(path)
		}
		handleErr(err)
	}
}

func (r *repository) init() {
	var err error
	r.rootDir, err = os.Getwd()
	handleErr(err)

	r.repositoryDir = filepath.Join(r.rootDir, vcsDirName)

	r.ensurePathExists(r.repositoryDir, true)

	r.configFile = filepath.Join(r.repositoryDir, configFileName)
	r.ensurePathExists(r.configFile, false)
	if fileExists(r.configFile) {
		r.readConfig()
	}

	r.indexFile = filepath.Join(r.repositoryDir, indexFileName)
	r.ensurePathExists(r.indexFile, false)
	if fileExists(r.indexFile) {
		r.readIndex()
	}
}

func (r *repository) readConfig() {
	file, err := os.Open(r.configFile)
	handleErr(err)

	defer file.Close()

	stat, err := file.Stat()
	handleErr(err)

	if stat.Size() == 0 {
		return
	}

	_, err = fmt.Fscanf(file, "%s", &r.user.name)
	handleErr(err)
}

func (r *repository) writeConfig() {
	file, err := os.OpenFile(r.configFile, os.O_WRONLY, 0644)
	handleErr(err)

	defer file.Close()

	_, err = file.WriteString(r.user.name)
	handleErr(err)
}

func (r *repository) setConfig(userName string) {
	r.user.name = userName
	r.writeConfig()
	r.printConfig()
}

func (r *repository) printConfig() {
	if r.user.name == "" {
		fmt.Println("Please, tell me who you are.")
		return
	}

	fmt.Printf("The username is %s.\n", r.user.name)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func (r *repository) readIndex() {
	file, err := os.Open(r.indexFile)
	handleErr(err)

	defer file.Close()

	stat, err := file.Stat()
	handleErr(err)

	if stat.Size() == 0 {
		return
	}

	var fileName string
	for {
		_, err = fmt.Fscanf(file, "%s", &fileName)
		if err != nil {
			break
		}
		r.index = append(r.index, fileName)
	}
}

func (r *repository) writeIndex() {
	file, err := os.OpenFile(r.indexFile, os.O_WRONLY, 0644)
	handleErr(err)

	defer file.Close()

	for _, fileName := range r.index {
		_, err = file.WriteString(fileName + "\n")
		handleErr(err)
	}
}

func (r *repository) addToIndex(fileName string) {
	r.index = append(r.index, fileName)
	r.writeIndex()
}

func (r *repository) removeFromIndex(fileName string) {
	var index []string
	for _, file := range r.index {
		if file != fileName {
			index = append(index, file)
		}
	}
	r.index = index
	r.writeIndex()
}

func (r *repository) containsInIndex(fileName string) bool {
	for _, file := range r.index {
		if file == fileName {
			return true
		}
	}
	return false
}

func (r *repository) add(fileName string) {
	file := filepath.Join(r.rootDir, fileName)

	if !fileExists(file) {
		fmt.Printf("Can't find '%s'.\n", fileName)
		return
	}

	if !r.containsInIndex(fileName) {
		r.addToIndex(fileName)
	}

	fmt.Printf("The file '%s' is tracked.\n", fileName)
}

func (r *repository) printIndex() {
	if len(r.index) == 0 {
		fmt.Printf("Add a file to the index.\n")
		return
	}

	fmt.Println("Tracked files:")
	for _, file := range r.index {
		fmt.Println(file)
	}
}

func printError(command string) {
	fmt.Printf("'%s' is not a SVCS command.\n", command)
}

func (r *repository) printDescription(command string) {
	fmt.Println(r.commands[command])
}

type Args struct {
	Command string
	Params  []string
}

func parseArgs(args []string) Args {
	if len(args) < 2 {
		return Args{Command: "--help"}
	}

	command := args[1]
	params := args[2:]

	return Args{Command: command, Params: params}
}

func main() {
	args := parseArgs(os.Args)

	r := newRepository()
	r.init()

	if _, ok := r.commands[args.Command]; !ok {
		printError(args.Command)
		return
	}

	if args.Command == "config" {
		if len(args.Params) == 0 {
			r.printConfig()
		} else if len(args.Params) == 1 {
			r.setConfig(args.Params[0])
		} else {
			printError(args.Command)
		}
	} else if args.Command == "add" {
		if len(args.Params) == 0 {
			r.printIndex()
		} else if len(args.Params) == 1 {
			r.add(args.Params[0])
		} else {
			printError(args.Command)
		}
	} else {
		r.printDescription(args.Command)
	}
}
