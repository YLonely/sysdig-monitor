package localprom

var tasks = []PromTask{}

func Register(t PromTask) {
	tasks = append(tasks, t)
}

func RunMetrics() {
	for _, t := range tasks {
		t.Run()
	}
}
