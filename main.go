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

type TaskError struct {
	Code string
	Msg  string
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Msg)
}

func main() {
	var (
		message  string
		taskDone int

		list bool
	)
	flag.StringVar(&message, "a", "", "Create new task")
	flag.BoolVar(&list, "l", false, "List tasks")
	flag.IntVar(&taskDone, "u", 0, "Set task done")
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
		fmt.Println("New task saved!")
	}

	if list {
		err = listTasks(file)
		if err != nil {
			log.Fatalf("Error task list %v\n", err)
		}
	}

	if taskDone != 0 {
		err = setTaskDone(file, taskDone)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		fmt.Println("Task marked as done!")
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

func setTaskDone(database *os.File, id int) error {
	// Positioning cursor to start of file
	_, err := database.Seek(0, 0)
	if err != nil {
		return &TaskError{Code: "SeekError", Msg: fmt.Sprintf("error seeking to start of file %v", err)}
	}

	var lines [][]string
	reader := csv.NewReader(database)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return &TaskError{Code: "CSVReadError", Msg: fmt.Sprintf("error reading csv %v", err)}
		}
		lines = append(lines, record)
	}

	var found bool
	for i, line := range lines {
		taskID, _ := strconv.Atoi(line[0])
		if taskID == id {
			lines[i][3] = "true"
			found = true
			break
		}
	}
	if !found {
		return &TaskError{Code: "NotFound", Msg: fmt.Sprintf("task with id %v not found", id)}
	}

	// Truncate and write updated lines
	err = database.Truncate(0)
	if err != nil {
		return &TaskError{Code: "TruncateError", Msg: fmt.Sprintf("error truncating file %v", err)}
	}
	_, err = database.Seek(0, 0)
	if err != nil {
		return &TaskError{Code: "SeekError", Msg: fmt.Sprintf("error seeking to start of file %v", err)}
	}

	writer := csv.NewWriter(database)
	for _, line := range lines {
		if err := writer.Write(line); err != nil {
			return &TaskError{Code: "CSVWriteError", Msg: fmt.Sprintf("error writing csv %v", err)}
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return &TaskError{Code: "CSVWriteError", Msg: fmt.Sprintf("error flushing csv %v", err)}
	}

	return nil
}
