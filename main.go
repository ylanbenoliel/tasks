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
		message      string
		taskDone     int
		taskToDelete int
		list         bool
	)
	flag.StringVar(&message, "a", "", "Create new task")
	flag.BoolVar(&list, "l", false, "List tasks")
	flag.IntVar(&taskDone, "u", 0, "Set task done")
	flag.IntVar(&taskToDelete, "d", 0, "Delete task")
	flag.Parse()

	file, err := OpenDatabase()
	if err != nil {
		log.Fatalf("%v\n", err)
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
			log.Fatalf("%v\n", err)
		}
	}

	if taskDone != 0 {
		err = setTaskDone(file, taskDone)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		fmt.Println("Task marked as done!")
	}
	if taskToDelete != 0 {
		err = deleteTask(file, taskToDelete)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		fmt.Println("Task deleted!")
	}
}

func OpenDatabase() (*os.File, error) {
	file, err := os.OpenFile("db.csv", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, &TaskError{
			Code: "DatabaseError",
			Msg:  fmt.Sprintf("error opening database %v", err),
		}
	}
	return file, nil
}

func addTask(database *os.File, message string) error {
	// Positioning cursor to start of file
	_, err := database.Seek(0, 0)
	if err != nil {
		return &TaskError{
			Code: "SeekError",
			Msg:  fmt.Sprintf("error seeking start of file %v", err),
		}
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
			return &TaskError{
				Code: "CSVReadError",
				Msg:  fmt.Sprintf("error reading csv %v", err),
			}
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
		return &TaskError{
			Code: "SeekError",
			Msg:  fmt.Sprintf("error seeking start of file %v", err),
		}
	}

	var line = []string{
		strconv.FormatInt(int64(task.id), 10),
		task.message,
		task.created.Format(time.RFC3339),
		strconv.FormatBool(task.done)}

	writer := csv.NewWriter(database)

	err = writer.Write(line)
	if err != nil {
		return &TaskError{
			Code: "CSVWriteError",
			Msg:  fmt.Sprintf("error writing task %v", err),
		}
	}

	writer.Flush()
	if err = writer.Error(); err != nil {
		return &TaskError{
			Code: "CSVWriteError",
			Msg:  fmt.Sprintf("error flushing csv %v", err),
		}
	}

	return nil
}

func listTasks(database *os.File) error {
	lines, err := readAllTasks(database)
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(
		os.Stdout, 0, 2, 4, ' ', 0,
	)

	writer.Write(
		[]byte("ID\tTask\tCreated\tDone\n"),
	)

	for _, line := range lines {
		dateTime, _ := time.Parse(time.RFC3339, line[2])
		dateTimeFormatted := dateTime.Format("15:04:05 02/01/06")
		writer.Write(
			[]byte(fmt.Sprintf("%v\t%v\t%v\t%v\n", line[0], line[1], dateTimeFormatted, line[3])),
		)
	}

	err = writer.Flush()
	if err != nil {
		return &TaskError{
			Code: "OSWriterError",
			Msg:  fmt.Sprintf("error printing tasks %v", err),
		}
	}

	return nil
}

func setTaskDone(database *os.File, id int) error {
	lines, err := readAllTasks(database)
	if err != nil {
		return err
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

	if err = writeAllTasks(database, lines); err != nil {
		return err
	}

	return nil
}

func deleteTask(database *os.File, id int) error {
	lines, err := readAllTasks(database)
	if err != nil {
		return err
	}

	var found bool
	for i, line := range lines {
		taskID, _ := strconv.Atoi(line[0])
		if taskID == id {
			lines = append(lines[:i], lines[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return &TaskError{Code: "NotFound", Msg: fmt.Sprintf("task with id %v not found", id)}
	}

	if err = writeAllTasks(database, lines); err != nil {
		return err
	}

	return nil
}

func readAllTasks(database *os.File) ([][]string, error) {
	_, err := database.Seek(0, 0)
	if err != nil {
		return nil, &TaskError{
			Code: "SeekError",
			Msg:  fmt.Sprintf("error seeking to start of file %v", err),
		}
	}

	var lines [][]string
	reader := csv.NewReader(database)
	reader.Comma = ';'
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, &TaskError{
				Code: "CSVReadError",
				Msg:  fmt.Sprintf("error reading csv %v", err),
			}
		}
		lines = append(lines, record)
	}
	return lines, nil
}

func writeAllTasks(database *os.File, lines [][]string) error {
	err := database.Truncate(0)
	if err != nil {
		return &TaskError{
			Code: "TruncateError",
			Msg:  fmt.Sprintf("error truncate file %v", err),
		}
	}
	_, err = database.Seek(0, 0)
	if err != nil {
		return &TaskError{
			Code: "SeekError",
			Msg:  fmt.Sprintf("error seeking start of file %v", err),
		}
	}

	writer := csv.NewWriter(database)
	writer.Comma = ';'
	for _, line := range lines {
		if err = writer.Write(line); err != nil {
			return &TaskError{
				Code: "CSVWriteError",
				Msg:  fmt.Sprintf("error writing csv %v", err),
			}
		}
	}

	writer.Flush()
	if err = writer.Error(); err != nil {
		return &TaskError{
			Code: "CSVWriteError",
			Msg:  fmt.Sprintf("error flushing csv %v", err),
		}
	}

	return nil
}
