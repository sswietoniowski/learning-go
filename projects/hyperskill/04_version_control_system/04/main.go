package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type user struct {
	name string
}

type Repository struct {
	user          user
	rootDir       string
	repositoryDir string
	configFile    string
	index         []string
	indexFile     string
	commands      map[string]string
	commitsDir    string
	logFile       string
}

// NewRepository creates a new instance of the Repository struct and returns a pointer to it.
func NewRepository() *Repository {
	return &Repository{
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
	commitsDirName = "commits"
	configFileName = "config.txt"
	indexFileName  = "index.txt"
	logFileName    = "log.txt"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (r *Repository) ensurePathExists(path string, isDir bool) {
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

// Init initializes the repository by setting up necessary directories and files.
func (r *Repository) Init() {
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

	r.commitsDir = filepath.Join(r.repositoryDir, commitsDirName)
	r.ensurePathExists(r.commitsDir, true)

	r.logFile = filepath.Join(r.repositoryDir, logFileName)
	r.ensurePathExists(r.logFile, false)
}

func (r *Repository) readConfig() {
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

func (r *Repository) writeConfig() {
	file, err := os.OpenFile(r.configFile, os.O_WRONLY, 0644)
	handleErr(err)

	defer file.Close()

	_, err = file.WriteString(r.user.name)
	handleErr(err)
}

// SetConfig sets the username in the repository's configuration.
func (r *Repository) SetConfig(userName string) {
	r.user.name = userName
	r.writeConfig()
	r.PrintConfig()
}

// PrintConfig prints the current username from the repository's configuration.
func (r *Repository) PrintConfig() {
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

func (r *Repository) readIndex() {
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

func (r *Repository) writeIndex() {
	file, err := os.OpenFile(r.indexFile, os.O_WRONLY, 0644)
	handleErr(err)

	defer file.Close()

	for _, fileName := range r.index {
		_, err = file.WriteString(fileName + "\n")
		handleErr(err)
	}
}

func (r *Repository) addToIndex(fileName string) {
	r.index = append(r.index, fileName)
	r.writeIndex()
}

func (r *Repository) removeFromIndex(fileName string) {
	var index []string
	for _, file := range r.index {
		if file != fileName {
			index = append(index, file)
		}
	}
	r.index = index
	r.writeIndex()
}

func (r *Repository) containsInIndex(fileName string) bool {
	for _, file := range r.index {
		if file == fileName {
			return true
		}
	}
	return false
}

// Add adds a file to the repository's index for tracking.
func (r *Repository) Add(fileName string) {
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

// PrintIndex prints the list of files currently being tracked in the repository's index.
func (r *Repository) PrintIndex() {
	if len(r.index) == 0 {
		fmt.Printf("Add a file to the index.\n")
		return
	}

	fmt.Println("Tracked files:")
	for _, file := range r.index {
		fmt.Println(file)
	}
}

// IsValidCommand checks if a given command is valid in the context of the repository.
func (r *Repository) IsValidCommand(command string) bool {
	_, ok := r.commands[command]
	return ok
}

// PrintDescription prints the description of a given command.
func (r *Repository) PrintDescription(command string) {
	fmt.Println(r.commands[command])
}

// PrintError prints an error message for an invalid command.
func (r *Repository) PrintError(command string) {
	fmt.Printf("'%s' is not a SVCS command.\n", command)
}

type commit struct {
	hash    string
	author  string
	message string
}

func (c *commit) String() string {
	return fmt.Sprintf("%s %s %s\n", c.hash, c.author, c.message)
}

func computeHashForFile(fileName string) string {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	md5Hash := md5.New()
	if _, err := io.Copy(md5Hash, file); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", md5Hash.Sum(nil))
}

func (r *Repository) generateHash() string {
	md5Hash := md5.New()
	for _, f := range r.index {
		file := filepath.Join(r.rootDir, f)
		md5Hash.Write([]byte(computeHashForFile(file)))
	}

	return fmt.Sprintf("%x", md5Hash.Sum(nil))
}

func (r *Repository) containsCommit(hash string) bool {
	file, err := os.Open(r.logFile)
	handleErr(err)
	defer file.Close()

	stat, err := file.Stat()
	handleErr(err)

	if stat.Size() == 0 {
		return false
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			panic("Invalid log entry!")
		}

		if parts[0] == hash {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		handleErr(err)
	}

	return false
}

func copyFile(source, destination string) {
	sourceFile, err := os.Open(source)
	handleErr(err)

	defer sourceFile.Close()

	destinationFile, err := os.Create(destination)
	handleErr(err)

	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	handleErr(err)
}

func (r *Repository) saveCommit(commit *commit) {
	commitDir := filepath.Join(r.commitsDir, commit.hash)
	r.ensurePathExists(commitDir, true)

	for _, f := range r.index {
		source := filepath.Join(r.rootDir, f)
		destination := filepath.Join(commitDir, f)
		copyFile(source, destination)
	}
}

// Commit saves the changes in the repository.
func (r *Repository) Commit(message string) {
	if len(r.index) == 0 {
		fmt.Println("Nothing to commit.")
		return
	}

	hash := r.generateHash()

	if r.containsCommit(hash) {
		fmt.Println("Nothing to commit.")
		return
	}

	c := &commit{
		hash:    hash,
		message: message,
		author:  r.user.name,
	}

	r.saveCommit(c)

	r.writeLog(c)

	fmt.Printf("Changes are committed.\n")
}

func (r *Repository) writeLog(commit *commit) {
	file, err := os.OpenFile(r.logFile, os.O_WRONLY|os.O_APPEND, 0644)
	handleErr(err)

	defer file.Close()

	_, err = file.WriteString(commit.String())
	handleErr(err)
}

// PrintLog prints the commit logs.
func (r *Repository) PrintLog() {
	file, err := os.Open(r.logFile)
	handleErr(err)
	defer file.Close()

	stat, err := file.Stat()
	handleErr(err)

	if stat.Size() == 0 {
		fmt.Println("No commits yet.")
		return
	}

	scanner := bufio.NewScanner(file)
	var logs []string
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			panic("Invalid log entry!")
		}

		hash, author, message := parts[0], parts[1], parts[2]
		formattedMessage := fmt.Sprintf("commit %s\nAuthor: %s\n%s", hash, author, message)
		logs = append(logs, formattedMessage)
	}

	if err := scanner.Err(); err != nil {
		handleErr(err)
	}

	for i := len(logs) - 1; i >= 0; i-- {
		fmt.Println(logs[i])
	}
}

// PrintHelp prints the help message.
func (r *Repository) PrintHelp() {
	fmt.Println(r.commands["--help"])
}

// PrintCommitError prints an error message for an invalid commit command.
func (r *Repository) PrintCommitError() {
	fmt.Printf("Message was not passed.\n")
}

// Checkout restores a file from a given commit.
func (r *Repository) Checkout(hash string) {
	commitDir := filepath.Join(r.commitsDir, hash)
	if !fileExists(commitDir) {
		fmt.Println("Commit does not exist.")
		return
	}

	for _, f := range r.index {
		destination := filepath.Join(r.rootDir, f)
		source := filepath.Join(commitDir, f)
		copyFile(source, destination)
	}

	fmt.Printf("Switched to commit %s.\n", hash)
}

// PrintCheckoutError prints an error message for an invalid checkout command.
func (r *Repository) PrintCheckoutError() {
	fmt.Printf("Commit id was not passed.\n")
}

type arguments struct {
	command string
	params  []string
}

func parseArgs(args []string) arguments {
	if len(args) < 2 {
		return arguments{command: "--help"}
	}

	command := args[1]
	params := args[2:]

	return arguments{command: command, params: params}
}

func main() {
	args := parseArgs(os.Args)

	r := NewRepository()
	r.Init()

	if !r.IsValidCommand(args.command) {
		r.PrintError(args.command)
		return
	}

	if args.command == "config" {
		if len(args.params) == 0 {
			r.PrintConfig()
		} else if len(args.params) == 1 {
			r.SetConfig(args.params[0])
		} else {
			r.PrintError(args.command)
		}
	} else if args.command == "add" {
		if len(args.params) == 0 {
			r.PrintIndex()
		} else if len(args.params) == 1 {
			r.Add(args.params[0])
		} else {
			r.PrintError(args.command)
		}
	} else if args.command == "commit" {
		if len(args.params) == 1 {
			r.Commit(args.params[0])
		} else {
			r.PrintCommitError()
		}
	} else if args.command == "log" {
		r.PrintLog()
	} else if args.command == "checkout" {
		if len(args.params) == 1 {
			r.Checkout(args.params[0])
		} else {
			r.PrintCheckoutError()
		}
	} else if args.command == "--help" {
		r.PrintHelp()
	} else {
		r.PrintDescription(args.command)
	}
}
