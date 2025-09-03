package main

func main() {
	tasks := Tasks{}
	storage := NewStorage[Tasks]("tasks.json")
	storage.Load(&tasks)
	cmdFlags := NewCmdFlags()
	cmdFlags.Execute(&tasks)
	storage.Save(tasks)
}
