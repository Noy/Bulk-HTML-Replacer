package main

import (
	"flag"
	"os"
	"strings"
	"io"
	"sync"
	"time"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"runtime"
	"github.com/PuerkitoBio/goquery"
)

var (
	// Our flag. which in gogland we can configure in the 'Edit Configurations' section.
	sourceDir = flag.String("source", "source", "This is the directory that contains the files you want to edit")
)

// Target directories.
var targets []string

func main() {
	// Set the accessible CPUs
	accessibleCPUs := runtime.NumCPU()

	// Returns logical amount of CPUs usable by the current process.
	runtime.GOMAXPROCS(accessibleCPUs)
	// This parses the command line flag given which in our case will be --source.
	flag.Parse()

	// Print out our initial message.
	fmt.Printf("[+] reading all files in dir \"%s\" to modify\n", *sourceDir)

	/* go through the files in the source directory to figure out what we need to do.
		this will actually modify the global "targets" variable, look at the var above.
		 We use the "targets" var as a list of all files we intend to modify and is used in the actual "runAllTasks" function as well.
	*/

	// Walk through our source directory and call the prepareTask function - see below .
	err := filepath.Walk(*sourceDir, prepareTask)

	// if we do encounter an error, we don't want to continue, panic function is called.
	if err != nil {
		panic(err)
	}

	// we're ready!

	// Print how many items we have to process..
	fmt.Printf("[+] ready with %d items \n", len(targets))

	// set variable start equal to the current time.
	start := time.Now()

	// Call our runAllTasks method (which takes the param of the number of CPUs.
	runAllTasks(accessibleCPUs)

	// The duration will be the time since the (start variable), but we convert it to seconds here.
	duration := float64(time.Since(start)) / float64(time.Second)
	// Finishing print, saying how many processed items, with the duration and length of directories (target var above).
	fmt.Printf("[%.02fs] processed all items, wrote %d updates\n", duration, len(targets))

	/*
		Interesting point:
			- When I ran the task over more than 13,000 directories, which within them, contained more, this was our result from the print statement above:
			- [161.04s] processed all items, wrote 27046 updates
			- That roughly translates to 2 minutes and 41 seconds. Changing 27,046 .html files.
			- Asynchronously completing this task (parallel) took just under 3 minutes, imagine how long it'd take recursively?
	*/
}

// This functions is the way it changes the information needed.
func changeFile(file io.Reader) (newContent string, err error) {
	// goquery - https://github.com/PuerkitoBio/goquery
	// Get the document from the file
	doc, err := goquery.NewDocumentFromReader(file)
	// Error checking
	if err != nil {
		return
	}
	// Here we find the 'title' tag.
	doc.Find("title").Each(func(i int, sel *goquery.Selection) {
		// Get the text from within the selection e.g. (<title>HOMEPAGE OF SITE</title>.
		text := sel.Text()
		// Replace the selected text (not the tag itself) with the string below.
		// e.g. This example would change it in the file to: <title>Homepage Of Site</title>.
		sel.ReplaceWithHtml("<title>" + strings.ToLower(text) + "</title>")
		// There is a TON more you can do with goquery, this is just a small example.
	})

	// Set the content to the html file, we will write it in later.. (This is a different task!).
	newContent, err = doc.Html()
	// Stop, returning nothing.
	return
}

// Runs the tasks provided (Asynchronously)
func runAllTasks(parallel int) {
	// Creating our channel, to queue the task for the worker goroutines.
	taskChan := make(chan string)

	// This is to wait for all the goroutines to finish.
	var wg sync.WaitGroup

	// Iterate through the CPUs in this case, as we called above.
	for i := 0; i < parallel; i++ {
		// Add 1 to the wait group.
		wg.Add(1)

		go func() {
			// Call done when it's finished.
			defer wg.Done()

			// Run through all of the tasks.
			for path := range taskChan {
				doTask(path)
			}
		}()
	}

	// Send all the tasks as quickly as possible.
	for _, task := range targets {
		// Send task back to the original channel.
		taskChan <- task
	}
	// Once we've sent them all, close the channel (kills the worker routines once they finish current job.
	// Needs to be done otherwise it'll just keep expecting more
	close(taskChan)

	// Wait for all workers to finish
	wg.Wait()
}

// Prepares the task
func prepareTask(path string, fi os.FileInfo, errIn error) (err error) {
	//not sure what this is even for (why would os give us an error through this)
	//but if we get an error in, it becomes the error out
	if errIn != nil {
		err = errIn
		return
	}

	// Only append if we should process the file, our utility method below.
	if !shouldProcessFile(path, fi) {
		return
	}
	// Add the targets (directories) to the path.
	targets = append(targets, path)
	return
}

// Doing the actual task.
func doTask(path string) {
	// The current time.
	start := time.Now()

	// start by opening the source file (which would be our flag we configured).
	file, err := openFile(path)
	if err != nil {
		return
	}

	// If we have absolutely no issue with opening the file, we should close it when the task is complete.
	// Now, process the file into it's new contents.
	newContents, err := changeFile(file)
	// Close when done.
	file.Close()
	if err != nil {
		return
	}

	// Here we finally write it to our file.
	err = ioutil.WriteFile(path, []byte(newContents), 0)
	if err != nil {
		return
	}

	// The duration, like above.
	duration := float64(time.Since(start)) / float64(time.Millisecond)

	// Log that things happened.
	// The amount of files the program changed.
	changed, err := filepath.Rel(*sourceDir, path)

	if err == nil {
		fmt.Printf("[%.02fms] processed \"%s\"\n", duration, changed)
	} else {
		fmt.Printf("[%.02fms] processed \"%s\"\n", duration, path)
		err = nil
	}
	return
}

// Utility functions

// This function checks if the file is an HTML file, if not, ignore it.
func shouldProcessFile(path string, fi os.FileInfo) bool {
	// This only listens to files which are html, you can change this to your liking obviously..
	return !fi.IsDir() && strings.HasSuffix(fi.Name(), ".html")
}

// This function opens the path to the file.
func openFile(path string) (file io.ReadWriteCloser, err error) {
	file, err = os.Open(path)
	return
}

// Empty directory checker, not used in this example, but useful to have.
func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)

	// If the file is EOF, the directory is empty.
	if err == io.EOF {
		return true, nil
	}
	return false, err
}