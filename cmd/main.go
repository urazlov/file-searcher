package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/sqweek/dialog"
)

func searchFiles(root string, pattern string, wg *sync.WaitGroup, filesChan chan<- string) {
    defer wg.Done()

    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if !info.IsDir() && strings.Contains(filepath.Base(path), pattern) {
            filesChan <- path
        }

        return nil
    })

    if err != nil {
        fmt.Printf("Error walking the path %q: %v\n", root, err)
    }
}

func printFiles(filesChan *chan string) []string {
	files := make([]string, 0, 10)
	for file := range *filesChan {
		fmt.Println("Found :", file)
		files = append(files, file)
	}

	return files
}

func openFile(filePath string) error {
	cmd := exec.Command("open", filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	fmt.Println("Select directory")

	root, err := dialog.Directory().Title("Select Directory").Browse()
	if err != nil {
		fmt.Printf("Error selecting directory: %v\n", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter file name \n")
	pattern, _ := reader.ReadString('\n')
	pattern = strings.TrimSpace(pattern)

    filesChan := make(chan string)
    var wg sync.WaitGroup

    wg.Add(1)

    go func() {
        wg.Wait()
        close(filesChan)
    }()

    go searchFiles(root, pattern, &wg, filesChan)

    files := printFiles(&filesChan)

	if len(files) == 0 {
		fmt.Println("Files not found")
	}

	for {
		fmt.Println("\nFiles found:")
		for i, file := range files {
			fmt.Printf("%d: %s\n", i+1, file)
		}

		fmt.Print("Enter the number of the file to open (or 'q' to quit): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "q" {
			break
		}

		index, err := strconv.Atoi(choice)
		if err != nil || index < 1 || index > len(files) {
			fmt.Println("Invalid choice, please try again.")
			continue
		}

		fileToOpen := files[index-1]
		fmt.Printf("Opening file: %s\n", fileToOpen)
		err = openFile(fileToOpen)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
		}
	}
}


