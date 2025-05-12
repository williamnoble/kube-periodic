output := "output"

run:
	go run . > {{ output }}.d2 && d2 {{ output }}.d2 --watch
