package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/eiannone/keyboard"
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

func chooseFiles(files []string) {
	fmt.Println("Use arrow keys to select a file and press Enter to open it.")

	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer keyboard.Close()

	selectedIndex := 0

	for {
		clearScreen()

		fmt.Println("\nFiles found:")
		for i, file := range files {
			prefix := " "
			if i == selectedIndex {
				prefix = ">"
			}
			fmt.Printf("%s %d: %s\n", prefix, i+1, file)
		}

		_, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Printf("Error reading key: %v\n", err)
			continue
		}

		if key == keyboard.KeyArrowUp {
			if selectedIndex > 0 {
				selectedIndex--
			}
		} else if key == keyboard.KeyArrowDown {
			if selectedIndex < len(files)-1 {
				selectedIndex++
			}
		} else if key == keyboard.KeyEnter {
			fileToOpen := files[selectedIndex]

			if selectedIndex == len(files) - 1 {
				break
			}

			fmt.Printf("Opening file: %s\n", fileToOpen)
			err = openFile(fileToOpen)
			if err != nil {
				fmt.Printf("Error opening file: %v\n", err)
			}
			
		}
	}
}

func selectDir() string {
	fmt.Println("Select directory")

	root, err := dialog.Directory().Title("Select Directory").Browse()
	if err != nil {
		fmt.Printf("Error selecting directory: %v\n", err)
		selectDir()
	}

	return root
}

func selectFilePtr() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter file name \n")
	pattern, _ := reader.ReadString('\n')
	pattern = strings.TrimSpace(pattern)

	return pattern
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func main() {
	root := selectDir()
	pattern := selectFilePtr()

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

	files = append(files, "Quit")

	chooseFiles(files)
}


