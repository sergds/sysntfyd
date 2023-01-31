package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type sysntfyd_cfg struct {
	Topic   string  `json:"ntfyapi_topic"`
	MaxTemp float64 `json:"temp_threshould"`
}

// ===== DEFAULTS! =====
// может быть перезаписано в конфиге
// Топик на ntfy.sh
var topic = "sysntfyd-sergdssrv"

// Максимальная температура в градусах Цельсия.
var temp_threshould = 73.0

var highesttemp = 0.0

var mainconfig sysntfyd_cfg

func dispatch_notification_title(title string, body string, tags string, topic string) {
	var req, _ = http.NewRequest("POST", "https://ntfy.sh/"+mainconfig.Topic, strings.NewReader(body))
	req.Header.Set("Title", title)
	req.Header.Set("Tags", tags)
	http.DefaultClient.Do(req)
}

func dispatch_error_noncritical(err string, topic string) {
	var host, _ = os.Hostname()
	dispatch_notification_title("Ошибка sysntfyd!", err+"\nsysntfyd продолжит работу...", "skull,x,"+host, topic)
}

func dispatch_overheating(temp string, topic string) {
	var host, _ = os.Hostname()
	dispatch_notification_title(host+" плавится!", "Огнетушитель не найдётся?\n"+"temp="+temp, "fire,skull,"+host, topic)
}

func main() {
	log.Print("sysntfyd STARTED!")
	mainconfig = sysntfyd_cfg{
		Topic:   topic,
		MaxTemp: temp_threshould,
	}
	f, err := os.Open("/etc/sysntfyd.json")
	if err != nil {
		f.Close()
		f, err := os.OpenFile("/etc/sysntfyd.json", os.O_WRONLY|os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			panic(err)
		}
		d, err := json.Marshal(mainconfig)
		if err != nil {
			panic(err)
		}
		f.Write(d)
		f.Close()
	} else {
		fi, err := f.Stat()
		if err != nil {
			panic(err)
		}
		buf := make([]byte, fi.Size())
		newcfg := sysntfyd_cfg{}
		f.Read(buf)
		log.Print(string(buf))
		err = json.Unmarshal(buf, &newcfg)
		if err != nil {
			panic(err)
		}
		mainconfig = newcfg
		f.Close()
	}
	log.Printf("Loaded config:\n...\tntfy.sh topic=%s\n...\tTemperature threshould=%f", mainconfig.Topic, mainconfig.MaxTemp)
	var host, _ = os.Hostname()
	var currtime = time.Now().Local().String()
	dispatch_notification_title("sysntfyd запустился", "на "+host+"\n"+currtime+"\n"+"CONFIG:"+"\n"+"...\tntfy.sh topic="+mainconfig.Topic+"\n"+"...\tTemperature threshould="+strconv.FormatFloat(mainconfig.MaxTemp, 'f', 1, 64), "white_check_mark", topic)
	//dispatch_overheating("80", topic)
	for {
		time.Sleep(5 * time.Second)
		var cmd = exec.Command("vcgencmd", "measure_temp")
		var stdout, err = cmd.Output()
		if err != nil {
			dispatch_error_noncritical(err.Error(), topic)
			continue
		}
		b := string(stdout)
		var temp, _ = strconv.ParseFloat(strings.Split(strings.Split(b, "=")[1], "'")[0], 64)
		//log.Println(strconv.FormatFloat(temp, 'f', 1, 64)+"'C")
		if temp > mainconfig.MaxTemp && highesttemp+3 < temp {
			highesttemp = temp
			dispatch_overheating(strconv.FormatFloat(temp, 'f', 1, 64)+"'C", topic)
		}
		if temp < mainconfig.MaxTemp-3 && highesttemp > 0.0 {
			highesttemp = 0.0
			dispatch_notification_title("Перегрев устранён!", "temp="+strconv.FormatFloat(temp, 'f', 1, 64)+"'C", "white_check_mark,"+host, topic)
		}
	}
}
