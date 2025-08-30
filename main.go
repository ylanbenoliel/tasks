package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"text/tabwriter"
	"time"
)

type Task struct {
	id      int
	message string
	created time.Time
	done    bool
}

func main() {
	var (
		message string
		list    bool
	)
	flag.StringVar(&message, "a", "", "Create new task")
	flag.BoolVar(&list, "l", false, "List tasks")
	flag.Parse()

	file, err := OpenDatabase()
	if err != nil {
		log.Fatalf("Error open database %v", err)
	}
	defer file.Close()

	if message != "" {
		err = addTask(file, message)
		if err != nil {
			log.Fatalf("Error adding new task %v", err)
		}
		fmt.Print("New task saved!")
	}

	if list {
		err = listTasks(file)
		if err != nil {
			log.Fatalf("Error task list %v", err)
		}
	}
}

func OpenDatabase() (*os.File, error) {
	file, err := os.OpenFile("db.csv", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	return file, err
}

func addTask(database *os.File, message string) error {
	// Positioning cursor to start of file
	_, err := database.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error seeking to start of file %v", err)
	}

	reader := csv.NewReader(database)
	var lastLine []string
	var lastID = 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading csv %v", err)
		}
		lastLine = record
	}

	if len(lastLine) > 0 {
		lastID, _ = strconv.Atoi(lastLine[0])
	}

	task := Task{
		message: message,
		created: time.Now().Local(),
		done:    false,
	}
	task.id = lastID + 1

	_, err = database.Seek(0, 2)
	if err != nil {
		return fmt.Errorf("error seeking end of file %v", err)
	}

	var line = []string{
		strconv.FormatInt(int64(task.id), 10),
		task.message,
		task.created.Format(time.RFC3339),
		strconv.FormatBool(task.done)}

	writer := csv.NewWriter(database)

	err = writer.Write(line)
	if err != nil {
		return fmt.Errorf("error writing record %v", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error writing new task %v", err)
	}
	return nil
}

func listTasks(database *os.File) error {
	// Positioning cursor to start of file
	_, err := database.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error seeking to start of file %v", err)
	}

	var lines [][]string
	reader := csv.NewReader(database)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading csv %v", err)
		}
		lines = append(lines, record)
	}

	writer := tabwriter.NewWriter(
		os.Stdout, 0, 2, 4, ' ', 0,
	)

	writer.Write(
		[]byte("ID\tTask\tCreated\tDone\n"),
	)

	for _, line := range lines {
		writer.Write(
			[]byte(fmt.Sprintf("%v\t%v\t%v\t%v\n", line[0], line[1], line[2], line[3])),
		)
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("error getting tasks %v", err)
	}

	return nil
}
