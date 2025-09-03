package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aquasecurity/table"
)

type Task struct {
	Message     string
	CreatedAt   time.Time
	CompletedAt *time.Time
	Done        bool
}

type Tasks []Task

func (tasks *Tasks) add(message string) {
	task := Task{
		Message:     message,
		Done:        false,
		CreatedAt:   time.Now(),
		CompletedAt: nil,
	}
	*tasks = append(*tasks, task)
}

func (tasks *Tasks) validateIndex(index int) error {
	if index < 0 || index >= len(*tasks) {
		err := errors.New("invalid index")
		fmt.Println(err)
		return err
	}
	return nil
}

func (tasks *Tasks) delete(index int) error {
	t := *tasks
	if err := t.validateIndex(index); err != nil {
		return err
	}

	*tasks = append(t[:index], t[index+1:]...)

	return nil
}

func (tasks *Tasks) toggle(index int) error {
	t := *tasks

	if err := t.validateIndex(index); err != nil {
		return err
	}

	done := t[index].Done
	if !done {
		completionTime := time.Now()
		t[index].CompletedAt = &completionTime
	}

	t[index].Done = !done

	return nil
}

func (tasks *Tasks) print() {
	table := table.New(os.Stdout)

	table.SetRowLines(false)
	table.SetHeaders("#", "Message", "Done", "Created at", "Completed at")
	for i, t := range *tasks {
		done := "❌"
		completedAt := ""

		if t.Done {
			done = "✅"
			if t.CompletedAt != nil {
				completedAt = t.CompletedAt.Format("15:04 02/01/06")
			}
		}
		table.AddRow(strconv.Itoa(i), t.Message, done, t.CreatedAt.Format("15:04 02/01/06"), completedAt)
	}

	table.Render()
}
