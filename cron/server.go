package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"./task"
)

var taskList task.List

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Usage: go run server.go [port]")
	}
	port := os.Args[1]
	fmt.Println("Gocron listening on port: " + port)

	http.HandleFunc("/", listenHandle)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Println(err)
	}
}

func listenHandle(res http.ResponseWriter, req *http.Request) {
	type Sucjson struct {
		Ok bool   `json:"ok"`
		ID string `json:"id"`
	}
	type Failjson struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}
	urlpath := req.URL.Path
	method := req.Method

	if urlpath == "/" {
		switch method {
		case "POST":
			data := getBody(req.Body)
			saved := saveTask(&data)
			if saved {
				sucjson, _ := json.Marshal(Sucjson{Ok: true, ID: data.ID})
				go execCmd(&data)
				log.Println("Run cmd: ", data.Cmd)
				res.Write(sucjson)
			} else {
				failjson, _ := json.Marshal(Failjson{Ok: false, Error: "The task " + data.ID + " already exists."})
				res.WriteHeader(http.StatusConflict)
				res.Write(failjson)
			}
		case "DELETE":
			data := getBody(req.Body)
			deleted := delTask(data)
			if deleted {
				sucjson, _ := json.Marshal(Sucjson{Ok: true, ID: data.ID})
				log.Println("Delete ", data.ID, " task successd.")
				res.Write(sucjson)
			} else {
				failjson, _ := json.Marshal(Failjson{Ok: false, Error: "The task " + data.ID + " is not found."})
				res.WriteHeader(http.StatusConflict)
				res.Write(failjson)
			}
		}
	}
}

func getBody(body io.ReadCloser) task.Task {
	var data task.Task
	dec := json.NewDecoder(body)
	dec.Decode(&data)

	return data
}

func saveTask(data *task.Task) (saved bool) {
	saved = taskList.SaveTask(data)
	return
}

func delTask(data task.Task) (deleted bool) {
	deleted = taskList.DelTask(data.ID)
	return
}

func execCmd(data *task.Task) {
	cmdText := data.Cmd
	interval := data.Interval
	args := data.Args
	sig := data.Sig
	tickC := time.Tick(time.Duration(interval) * time.Millisecond)

	for {
		// task's signal
		if <-sig {
			<-tickC
			sig <- true
			// signal have 2 cache
			// deleted task in ticker duration will not delay
			if <-sig {
				sig <- true
				cmd := exec.Command(cmdText, args...)
				stdout, err := cmd.StdoutPipe()
				if err != nil {
					log.Fatalln(err)
				}

				cmd.Start()
				reader := bufio.NewReader(stdout)
				for {
					line, err := reader.ReadString('\n')
					if err != nil || io.EOF == err {
						break
					}
					log.Print(data.ID, ": ", line)
				}

				cmd.Wait()
			} else {
				break
			}
		} else {
			break
		}
	}
}
