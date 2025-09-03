package main

import (
	"errors"
	"fmt"
	"time"
)

type Task struct {
	message     string
	createdAt   time.Time
	completedAt *time.Time
	done        bool
}

type Tasks []Task

func (tasks *Tasks) add(message string) {
	task := Task{
		message:     message,
		done:        false,
		createdAt:   time.Now(),
		completedAt: nil,
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

	done := t[index].done
	if !done {
		completionTime := time.Now()
		t[index].completedAt = &completionTime
	}

	t[index].done = !done

	return nil
}
