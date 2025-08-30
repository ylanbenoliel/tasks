package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

type Task struct {
	id      int
	message string
	created time.Time
	done    bool
}

func main() {
	var message string
	flag.StringVar(&message, "a", "", "Create new task")
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
