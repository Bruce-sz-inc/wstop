package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/communaute-cimi/glay"
	"github.com/communaute-cimi/linuxproc"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Configuration struct {
	MaxFailure   int                `json:"maxfailure"`
	Applications []glay.Application `json:"apps"`
}

const VERSION = "1.0.0"

var (
	flhttpd   *bool   = flag.Bool("httpd", false, "httpd")
	flprocess *string = flag.String("process", "", "liste,process,listen")
	fllisten  *string = flag.String("listen", ":8080", "Listen ip:port or :port")
)

type App struct {
	Name   string
	Pid    int
	VmData int
	VmPeak int
	State  string
}

func listProcess() (apps []App) {
	// todo: fonction bof à revoir. c'est vraiment pour que ca marche
	d, err := os.Open("/proc")
	if err != nil {
		return
	}
	defer d.Close()
	fi, _ := d.Readdir(0)
	for _, f := range fi {
		pid, err := strconv.Atoi(f.Name())
		if err != nil {
			continue
		}
		p, err := linuxproc.FindProcess(pid)
		if err != nil {
			continue
		}
		vmdata, err := p.VmData()
		if err != nil {
			continue
		}
		vmpeak, err := p.VmPeak()
		if err != nil {
			continue
		}
		vmstat, err := p.State()
		if err != nil {
			continue
		}

		for _, n := range strings.Split(*flprocess, ",") {
			if strings.Contains(p.Name, strings.TrimSpace(n)) {
				if len(apps) == 0 {
					apps = append(apps, App{p.Name, p.Pid, vmdata, vmpeak, vmstat})
				} else {
					exist := false
					for i, a := range apps {
						if a.Name == p.Name {
							vmdata = a.VmData + vmdata
							vmpeak = a.VmPeak + vmpeak
							// del ancien pid du slice pour ajouter avec les nouvelles values
							apps = append(apps[:i], apps[i+1:]...)
							apps = append(apps, App{a.Name, a.Pid, vmdata, vmpeak, vmstat})
							exist = true
						}
					}
					if !exist {
						apps = append(apps, App{p.Name, p.Pid, vmdata, vmpeak, vmstat})
					}
				}
			}
		}

	}
	return
}

func mainHandler() http.Handler {

	type Data struct {
		Apps     []App
		MemTotal float32
		MemFree  float32
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl, err := template.ParseFiles("tpl/main.html")
		if err != nil {
			log.Printf("%s", err)
		}

		data := new(Data)
		memory := new(linuxproc.Memory)
		memFree, _ := memory.MemFree()
		data.MemFree = float32(memFree) * 9.53674316406E-7
		memTotal, _ := memory.MemTotal()
		data.MemTotal = float32(memTotal)*9.53674316406E-7 - data.MemFree
		data.Apps = listProcess()
		tpl.Execute(w, data)
	})
}

func wsMemoryConsoHandler(ws *websocket.Conn) {
	for {
		memory := new(linuxproc.Memory)
		mf, _ := memory.MemFree()
		mt, _ := memory.MemTotal()
		// utiliser json.MArshal car là, c'est moche ...
		msg := fmt.Sprintf("[{\"name\":\"free\",\"y\":%.2f,\"color\":\"#D0FA58\"},{\"name\":\"occ\",\"y\":%.2f,\"color\":\"#F78181\"}]", float32(mf)*9.53674316406E-7, float32(mt)*9.53674316406E-7-float32(mf)*9.53674316406E-7)
		name := []byte(msg)
		ws.Write(name)
		time.Sleep(250 * time.Millisecond)
	}
}

func wsMemoryProcessGraphHandler(ws *websocket.Conn) {
	i := 0
	for {
		lp := listProcess()
		lpoints := []string{}
		for _, p := range lp {
			t := time.Now().Unix()
			// utiliser json.MArshal car là, c'est moche ...
			lpoints = append(lpoints, fmt.Sprintf("{\"x\":%d,\"y\":%.2f,\"name\":\"%s\"}", t, float32(p.VmPeak)*9.53674316406E-7, p.Name))
		}
		msg, _ := json.Marshal(lpoints)
		ws.Write(msg)
		time.Sleep(250 * time.Millisecond)
		i += 1
	}
}

func wsMemoryProcessHandler(ws *websocket.Conn) {
	i := 0
	for {
		lp := listProcess()
		lpoints := []string{}
		for _, p := range lp {
			// t := time.Now().Unix()
			// utiliser json.MArshal car là, c'est moche ...
			lpoints = append(lpoints, fmt.Sprintf("{\"peak\":%.2f,\"data\":%.2f,\"name\":\"%s\",\"pid\":%d,\"state\":\"%s\"}", float32(p.VmData)*9.53674316406E-7, float32(p.VmPeak)*9.53674316406E-7, p.Name, p.Pid, p.State))
		}
		msg, _ := json.Marshal(lpoints)
		ws.Write(msg)
		time.Sleep(2000 * time.Millisecond)
		i += 1
	}
}

func main() {
	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
	}

	if *flhttpd && *flprocess != "" {
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
		http.Handle("/wsmemoryconsograph/", websocket.Handler(wsMemoryConsoHandler))
		http.Handle("/wsmemoryprocessgraph/", websocket.Handler(wsMemoryProcessGraphHandler))
		http.Handle("/wsmemoryprocess/", websocket.Handler(wsMemoryProcessHandler))
		http.Handle("/", mainHandler())
		http.ListenAndServe(*fllisten, nil)
	}

	flag.Usage()
}
