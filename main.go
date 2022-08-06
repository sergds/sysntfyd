package main

import (
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Топик на ntfy.sh
var topic = "sysntfyd-sergdssrv"

// Максимальная температура в градусах Цельсия.
var temp_threshould = 70.0

var highesttemp = 0.0

func dispatch_notification_title(title string, body string, tags string, topic string) {
	var req, _ = http.NewRequest("POST", "https://ntfy.sh/"+topic, strings.NewReader(body))
	req.Header.Set("Title", title)
	req.Header.Set("Tags", tags)
	http.DefaultClient.Do(req)
}

func dispatch_error_noncritical(err string, topic string) {
	var host, _ = os.Hostname()
	dispatch_notification_title("Error while running sysntfyd!", err+"\nsysntfyd will still be running!", "skull,x,"+host, topic)
}

func dispatch_overheating(temp string, topic string) {
	var host, _ = os.Hostname()
	dispatch_notification_title(host+" плавится!", "Онетушитель не найдётся?\n"+"temp="+temp, "fire,skull,"+host, topic)
}

func main() {
	var host, _ = os.Hostname()
	var currtime = time.Now().Local().String()
	dispatch_notification_title("sysntfyd запустился", "на "+host+"\n"+currtime, "white_check_mark", topic)
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
		if temp > temp_threshould && highesttemp+3 < temp {
			highesttemp = temp
			dispatch_overheating(strconv.FormatFloat(temp, 'f', 1, 64)+"'C", topic)
		}
		if temp < temp_threshould-3 && highesttemp > 0.0 {
			highesttemp = 0.0
			dispatch_notification_title("Перегрев устранён!", "temp="+strconv.FormatFloat(temp, 'f', 1, 64)+"'C", "white_check_mark,"+host, topic)
		}
	}
}
