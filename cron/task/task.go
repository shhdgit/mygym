package task

// List is a storage list of task
type List []*Task

// Task is a task type struct
type Task struct {
	ID       string   `json:"id"`
	Cmd      string   `json:"cmd"`
	Args     []string `json:"args"`
	Interval int      `json:"interval"`
	Sig      chan bool
}

// SaveTask can save a Task into List
func (list *List) SaveTask(task *Task) (saved bool) {
	if _, _, ok := list.SearchList(task.ID); ok {
		saved = false
	} else {
		save(list, task)
		task.Sig = make(chan bool, 2)
		task.Sig <- true
		saved = true
	}

	return
}

// DelTask can delete a Task from List
func (list *List) DelTask(taskid string) (deleted bool) {
	if task, i, ok := list.SearchList(taskid); ok {
		remove(list, i)
		task.Sig <- false
		deleted = true
	} else {
		deleted = false
	}

	return
}

// SearchList search a task from TaskList
func (list *List) SearchList(taskid string) (*Task, int, bool) {
	for i, task := range *list {
		if task.ID == taskid {
			return task, i, true
		}
	}

	return &Task{}, 0, false
}

// saveTo save task into list
func save(list *List, task *Task) {
	*list = append(*list, task)
}

// remove delete task from list
func remove(list *List, index int) {
	l := *list
	copy(l[index:], l[index+1:])
	l = l[:len(l)-1]
}
