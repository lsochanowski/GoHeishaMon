package main

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/xid"
	"go.bug.st/serial"
)

var panasonicQuery []byte = []byte{0x71, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
var PANASONICQUERYSIZE int = 110

//should be the same number
var NUMBER_OF_TOPICS int = 92
var AllTopics [92]TopicData
var MqttKeepalive time.Duration
var CommandsToSend map[xid.ID][]byte
var GPIO map[string]string
var actData [92]string
var config Config
var sending bool
var Serial serial.Port
var err error
var goodreads float64
var totalreads float64
var readpercentage float64
var SwitchTopics map[string]AutoDiscoverStruct
var ClimateTopics map[string]AutoDiscoverStruct

type command_struct struct {
	value  [128]byte
	length int
}

type TopicData struct {
	TopicNumber        int
	TopicName          string
	TopicBit           int
	TopicFunction      string
	TopicUnit          string
	TopicA2M           string
	TopicType          string
	TopicDisplayUnit   string
	TopicValueTemplate string
}

type Config struct {
	Readonly               bool
	Loghex                 bool
	Device                 string
	ReadInterval           int
	MqttServer             string
	MqttPort               string
	MqttLogin              string
	Aquarea2mqttCompatible bool
	Mqtt_topic_base        string
	Mqtt_set_base          string
	Aquarea2mqttPumpID     string
	MqttPass               string
	MqttClientID           string
	MqttKeepalive          int
	ForceRefreshTime       int
	EnableCommand          bool
	SleepAfterCommand      int
	HAAutoDiscover         bool
}

var cfgfile *string
var topicfile *string
var configfile string

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func SetGPIODebug() {
	err := ioutil.WriteFile("/sys/class/gpio/export", []byte("2"), 0200)
	err = ioutil.WriteFile("/sys/class/gpio/export", []byte("3"), 0200)
	err = ioutil.WriteFile("/sys/class/gpio/export", []byte("13"), 0200)
	err = ioutil.WriteFile("/sys/class/gpio/export", []byte("15"), 0200)
	err = ioutil.WriteFile("/sys/class/gpio/export", []byte("10"), 0200)
	err = ioutil.WriteFile("/sys/class/gpio/export", []byte("0"), 0200)
	err = ioutil.WriteFile("/sys/class/gpio/export", []byte("1"), 0200)
	err = ioutil.WriteFile("/sys/class/gpio/export", []byte("16"), 0200)

	if err != nil {
		fmt.Println(err.Error())
	}
}

func GetGPIOStatus() {
	readFile, err := os.Open("/sys/kernel/debug/gpio")
	//readFile, err := os.Open("FakeKernel.txt")
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileTextLines []string

	for fileScanner.Scan() {
		fileTextLines = append(fileTextLines, fileScanner.Text())
	}

	readFile.Close()

	for _, eachline := range fileTextLines {
		s := strings.Fields(eachline)
		if len(s) > 3 {
			GPIO[s[0]] = s[4]
		}

	}
	if len(GPIO) > 1 {
		fmt.Println(GPIO)
		if GPIO["gpio-0"] == "lo" && GPIO["gpio-1"] == "lo" && GPIO["gpio-16"] == "hi" {
			err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("high"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("high"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("high"), 644)
		}
		if GPIO["gpio-0"] == "hi" || GPIO["gpio-1"] == "hi" || GPIO["gpio-16"] == "lo" {
			err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("high"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("low"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("low"), 644)
		}
		if GPIO["gpio-0"] == "hi" && GPIO["gpio-1"] == "hi" {
			err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("low"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("high"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("high"), 644)
		}
		if GPIO["gpio-0"] == "hi" && GPIO["gpio-16"] == "lo" {
			err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("low"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("high"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("high"), 644)
		}
		if GPIO["gpio-1"] == "hi" && GPIO["gpio-16"] == "lo" {
			err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("low"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("high"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("high"), 644)
		}
		if GPIO["gpio-0"] == "hi" && GPIO["gpio-1"] == "hi" && GPIO["gpio-16"] == "lo" {
			err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("low"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("low"), 644)
			err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("high"), 644)
			cmd := exec.Command("fwupdate", "sw")
			out, err := cmd.CombinedOutput()
			fmt.Println(out)
			cmd = exec.Command("sync")
			out, err = cmd.CombinedOutput()
			fmt.Println(out)
			cmd = exec.Command("reboot")
			out, err = cmd.CombinedOutput()
			fmt.Println(out)
			if err != nil {
				fmt.Println(err)
			}

		}
		if GPIO["gpio-10"] == "hi" {
			err := ioutil.WriteFile("/sys/class/gpio/gpio3/direction", []byte("low"), 644)
			if err != nil {
				fmt.Println(err)
			}

		}
		if GPIO["gpio-10"] == "lo" {
			err := ioutil.WriteFile("/sys/class/gpio/gpio3/direction", []byte("high"), 644)
			if err != nil {
				fmt.Println(err)
			}

		}

	}
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(time.Nanosecond * 500000000)

}

func ReadConfig() Config {

	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}
	return config
}

func UpdateConfig(configfile string) bool {
	fmt.Printf("try to update configfile: %s", configfile)
	out, err := exec.Command("/usr/bin/usb_mount.sh").Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(out)
	_, err = os.Stat("/mnt/usb/GoHeishaMonConfig.new")
	if err != nil {
		_, _ = exec.Command("/usr/bin/usb_umount.sh").Output()
		return false
	}
	if GetFileChecksum(configfile) != GetFileChecksum("/mnt/usb/GoHeishaMonConfig.new") {
		fmt.Printf("checksum of configfile and new configfile diffrent: %s ", configfile)

		_, _ = exec.Command("/bin/cp", "/mnt/usb/GoHeishaMonConfig.new", configfile).Output()
		if err != nil {
			fmt.Printf("can't update configfile %s", configfile)
			return false
		}
		_, _ = exec.Command("sync").Output()

		_, _ = exec.Command("/usr/bin/usb_umount.sh").Output()
		_, _ = exec.Command("reboot").Output()
		return true
	}
	_, _ = exec.Command("/usr/bin/usb_umount.sh").Output()

	return true
}

func EncodeTopicsToTOML(topnr int, data TopicData) {
	f, err := os.Create(fmt.Sprintf("data/%d", topnr))
	if err != nil {
		// failed to create/open the file
		log.Fatal(err)
	}
	if err := toml.NewEncoder(f).Encode(data); err != nil {
		// failed to encode
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		// failed to close the file
		log.Fatal(err)

	}

}
func UpdatePassword() bool {
	_, err = os.Stat("/mnt/usb/GoHeishaMonPassword.new")
	if err != nil {
		return true
	} else {
		_, _ = exec.Command("chmod", "+x", "/root/pass.sh").Output()
		dat, _ := ioutil.ReadFile("/mnt/usb/GoHeishaMonPassword.new")
		fmt.Printf("updejtuje haslo na: %s", string(dat))
		o, err := exec.Command("/root/pass.sh", string(dat)).Output()
		if err != nil {
			fmt.Println(err)
			fmt.Println(o)

			return false
		}
		fmt.Println(o)

		_, _ = exec.Command("/bin/rm", "/mnt/usb/GoHeishaMonPassword.new").Output()
	}
	return true
}

func GetFileChecksum(f string) string {
	input := strings.NewReader(f)

	hash := md5.New()
	if _, err := io.Copy(hash, input); err != nil {
		log.Fatal(err)
	}
	sum := hash.Sum(nil)

	return fmt.Sprintf("%x\n", sum)

}

func UpdateConfigLoop(configfile string) {
	for {
		UpdateConfig(configfile)
		time.Sleep(time.Minute * 5)

	}
}

func PublishTopicsToAutoDiscover(mclient mqtt.Client, token mqtt.Token) {
	for k, v := range AllTopics {
		var m AutoDiscoverStruct
		m.UID = fmt.Sprintf("Aquarea-%s-%d", config.MqttLogin, k)
		if v.TopicType == "" {
			v.TopicType = "sensor"
		}
		m.ValueTemplate = v.TopicValueTemplate

		m.UnitOfM = v.TopicDisplayUnit

		if v.TopicType == "binary_sensor" {
			m.UnitOfM = ""
			m.PayloadOn = "1"
			m.PayloadOff = "0"
			m.ValueTemplate = `{{ value }}`
		}
		if v.TopicDisplayUnit == "°C" {
			m.DeviceClass = "temperature"
		}
		if v.TopicDisplayUnit == "W" {
			m.DeviceClass = "power"
		}
		m.StateTopic = fmt.Sprintf("%s/%s", config.Mqtt_topic_base, v.TopicName)
		m.Name = fmt.Sprintf("TEST-%s", v.TopicName)
		Topic_Value, err := json.Marshal(m)
		//Topic_Value = []byte("")
		//v.TopicType = "sensor"

		fmt.Println(err)
		TOP := fmt.Sprintf("%s/%s/%s/config", config.Mqtt_topic_base, v.TopicType, strings.ReplaceAll(m.Name, " ", "_"))
		fmt.Println("Publikuje do ", TOP, "warosc", string(Topic_Value))
		token = mclient.Publish(TOP, byte(0), false, Topic_Value)
		if token.Wait() && token.Error() != nil {
			fmt.Printf("Fail to publish, %v", token.Error())
		}

	}

	for _, vs := range SwitchTopics {
		if vs.ValueTemplate == "" {
			vs.PayloadOff = "0"
			vs.PayloadOn = "1"
		}
		vs.Optimistic = "true"
		Topic_Value, err := json.Marshal(vs)
		//Topic_Value = []byte("")

		fmt.Println(err)
		TOP := fmt.Sprintf("%s/%s/%s/config", config.Mqtt_topic_base, "switch", strings.ReplaceAll(vs.Name, " ", "_"))
		fmt.Println("Publikuje do ", TOP, "warosc", string(Topic_Value))
		token = mclient.Publish(TOP, byte(0), false, Topic_Value)
		if token.Wait() && token.Error() != nil {
			fmt.Printf("Fail to publish, %v", token.Error())
		}
	}

}

type AutoDiscoverStruct struct {
	DeviceClass   string `json:"device_class,omitempty"`
	Name          string `json:"name,omitempty"`
	StateTopic    string `json:"state_topic,omitempty"`
	UnitOfM       string `json:"unit_of_measurement,omitempty"`
	ValueTemplate string `json:"value_template,omitempty"`
	CommandTopic  string `json:"command_topic,omitempty"`
	UID           string `json:"unique_id,omitempty"`
	PayloadOn     string `json:"payload_on,omitempty"`
	PayloadOff    string `json:"payload_off,omitempty"`
	Optimistic    string `json:"optimistic,omitempty"`
	StateON       string `json:"state_on,omitempty"`
	StateOff      string `json:"state_off,omitempty"`
}

func UpdateGPIOStat() {

	// watcher := gpio.NewWatcher()
	// //watcher.AddPin(0)
	// watcher.AddPin(1)
	// watcher.AddPin(2)
	// watcher.AddPin(3)
	// watcher.AddPin(4)
	// watcher.AddPin(5)
	// watcher.AddPin(6)
	// watcher.AddPin(7)
	// watcher.AddPin(8)
	// watcher.AddPin(9)
	// watcher.AddPin(10)
	// watcher.AddPin(11)
	// watcher.AddPin(12)
	// watcher.AddPin(13)
	// watcher.AddPin(14)
	// watcher.AddPin(15)
	// watcher.AddPin(16)

	// defer watcher.Close()

	// go func() {
	// 	var v string
	// 	for {
	// 		pin, value := watcher.Watch()
	// 		if value == 1 {
	// 			v = "hi"
	// 		} else {
	// 			v = "lo"
	// 		}
	// 		GPIO[fmt.Sprintf("gpio-%d", pin)] = v
	// 		fmt.Printf("read %d from gpio %d\n", value, pin)
	// 	}
	// }()

	GPIO = make(map[string]string)
	SetGPIODebug()
	for {
		GetGPIOStatus()
		//time.Sleep(time.Nanosecond * 500000000)
	}
}

func ExecuteGPIOCommand() {
	for {
		var err error
		if len(GPIO) > 1 {
			fmt.Println(GPIO)
			if GPIO["gpio-0"] == "lo" && GPIO["gpio-1"] == "lo" && GPIO["gpio-16"] == "hi" {
				err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("high"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("high"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("high"), 644)
			}
			if GPIO["gpio-0"] == "hi" || GPIO["gpio-1"] == "hi" || GPIO["gpio-16"] == "lo" {
				err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("high"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("low"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("low"), 644)
			}
			if GPIO["gpio-0"] == "hi" && GPIO["gpio-1"] == "hi" {
				err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("low"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("high"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("high"), 644)
			}
			if GPIO["gpio-0"] == "hi" && GPIO["gpio-16"] == "lo" {
				err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("low"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("high"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("high"), 644)
			}
			if GPIO["gpio-1"] == "hi" && GPIO["gpio-16"] == "lo" {
				err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("low"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("high"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("high"), 644)
			}
			if GPIO["gpio-0"] == "hi" && GPIO["gpio-1"] == "hi" && GPIO["gpio-16"] == "lo" {
				err = ioutil.WriteFile("/sys/class/gpio/gpio2/direction", []byte("low"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio13/direction", []byte("low"), 644)
				err = ioutil.WriteFile("/sys/class/gpio/gpio15/direction", []byte("high"), 644)
				cmd := exec.Command("fwupdate", "sw")
				out, err := cmd.CombinedOutput()
				fmt.Println(out)
				cmd = exec.Command("sync")
				out, err = cmd.CombinedOutput()
				fmt.Println(out)
				cmd = exec.Command("reboot")
				out, err = cmd.CombinedOutput()
				fmt.Println(out)
				fmt.Println(err)

			}
			if GPIO["gpio-10"] == "hi" {
				err := ioutil.WriteFile("/sys/class/gpio/gpio3/direction", []byte("low"), 644)
				fmt.Println(err)

			}
			if GPIO["gpio-10"] == "lo" {
				err := ioutil.WriteFile("/sys/class/gpio/gpio3/direction", []byte("high"), 644)
				fmt.Println(err)

			}

		}
		//time.Sleep(time.Nanosecond * 500000000)
		fmt.Println(err)

	}
}

func main() {
	SwitchTopics = make(map[string]AutoDiscoverStruct)

	//	cfgfile = flag.String("c", "config", "a config file patch")
	//	topicfile = flag.String("t", "Topics.csv", "a topic file patch")
	flag.Parse()
	if runtime.GOOS != "windows" {
		//	go UpdateGPIOStat()
		configfile = "/etc/gh/config"

	} else {
		configfile = "config"

	}
	_, err := os.Stat(configfile)
	if err != nil {
		fmt.Printf("Config file is missing: %s ", configfile)
		UpdateConfig(configfile)
	}
	go UpdateConfigLoop(configfile)
	c1 := make(chan bool, 1)
	go ClearActData()
	CommandsToSend = make(map[xid.ID][]byte)
	var in int
	config = ReadConfig()
	// if config.Readonly != true {
	// 	log_message("Not sending this command. Heishamon in listen only mode! - this POC version don't support writing yet....")
	// 	os.Exit(0)
	// }
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	for _, port := range ports {
		fmt.Printf("Found port: %v\n", port)
	}
	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.EvenParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	Serial, err = serial.Open(config.Device, mode)
	if err != nil {
		fmt.Println(err)
	}
	PoolInterval := time.Second * time.Duration(config.ReadInterval)
	ParseTopicList3()
	MqttKeepalive = time.Second * time.Duration(config.MqttKeepalive)
	MC, MT := MakeMQTTConn()
	if config.HAAutoDiscover == true {
		PublishTopicsToAutoDiscover(MC, MT)
	}
	for {
		if MC.IsConnected() != true {
			MC, MT = MakeMQTTConn()
		}
		if len(CommandsToSend) > 0 {
			fmt.Println("jest wiecej niz jedna komenda tj", len(CommandsToSend))
			in = 1
			for key, value := range CommandsToSend {
				if in == 1 {

					send_command(value, len(value))
					delete(CommandsToSend, key)
					in++
					time.Sleep(time.Second * time.Duration(config.SleepAfterCommand))

				} else {
					fmt.Println("numer komenty  ", in, " jest za duzy zrobie to w nastepnym cyklu")
					break
				}
				fmt.Println("koncze range po tablicy z komendami ")

			}

		} else {
			send_command(panasonicQuery, PANASONICQUERYSIZE)
		}
		go func() {
			tbool := readSerial(MC, MT)
			c1 <- tbool
		}()

		select {
		case res := <-c1:
			fmt.Println("read ma status", res)
		case <-time.After(5 * time.Second):
			fmt.Println("out of time for read :(")
		}

		time.Sleep(PoolInterval)

	}

}

func ClearActData() {
	for {
		time.Sleep(time.Second * time.Duration(config.ForceRefreshTime))
		for k, _ := range actData {
			actData[k] = "nil" //funny i know ;)
		}

	}
}

func MakeMQTTConn() (mqtt.Client, mqtt.Token) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s:%s", "tcp", config.MqttServer, config.MqttPort))
	opts.SetPassword(config.MqttPass)
	opts.SetUsername(config.MqttLogin)
	opts.SetClientID(config.MqttClientID)
	opts.SetWill(config.Mqtt_set_base+"/LWT", "Offline", 1, true)
	opts.SetKeepAlive(MqttKeepalive)
	opts.SetOnConnectHandler(startsub)
	opts.SetConnectionLostHandler(connLostHandler)

	// connect to broker
	client := mqtt.NewClient(opts)
	//defer client.Disconnect(uint(2))

	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		fmt.Printf("Fail to connect broker, %v", token.Error())
	}
	return client, token
}

func connLostHandler(c mqtt.Client, err error) {
	fmt.Printf("Connection lost, reason: %v\n", err)

	//Perform additional action...
}

func startsub(c mqtt.Client) {
	var t AutoDiscoverStruct

	c.Subscribe("aquarea/+/+/set", 2, HandleMSGfromMQTT)
	c.Subscribe(config.Mqtt_set_base+"/SetHeatpump", 2, HandleSetHeatpump)
	t.Name = "TEST-SetHeatpump"
	t.CommandTopic = config.Mqtt_set_base + "/SetHeatpump"
	t.StateTopic = config.Mqtt_topic_base + "/Heatpump_State"
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	SwitchTopics["SetHeatpump"] = t
	t = AutoDiscoverStruct{}
	c.Subscribe(config.Mqtt_set_base+"/SetQuietMode", 2, HandleSetQuietMode)
	c.Subscribe(config.Mqtt_set_base+"/SetZ1HeatRequestTemperature", 2, HandleSetZ1HeatRequestTemperature)
	c.Subscribe(config.Mqtt_set_base+"/SetZ1CoolRequestTemperature", 2, HandleSetZ1CoolRequestTemperature)
	c.Subscribe(config.Mqtt_set_base+"/SetZ2HeatRequestTemperature", 2, HandleSetZ2HeatRequestTemperature)
	c.Subscribe(config.Mqtt_set_base+"/SetZ2CoolRequestTemperature", 2, HandleSetZ2CoolRequestTemperature)
	c.Subscribe(config.Mqtt_set_base+"/SetOperationMode", 2, HandleSetOperationMode)
	c.Subscribe(config.Mqtt_set_base+"/SetForceDHW", 2, HandleSetForceDHW)
	t.Name = "TEST-SetForceDHW"
	t.CommandTopic = config.Mqtt_set_base + "/SetForceDHW"
	t.StateTopic = config.Mqtt_topic_base + "/Force_DHW_State"
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	SwitchTopics["SetForceDHW"] = t
	t = AutoDiscoverStruct{}
	c.Subscribe(config.Mqtt_set_base+"/SetForceDefrost", 2, HandleSetForceDefrost)
	t.Name = "TEST-SetForceDefrost"
	t.CommandTopic = config.Mqtt_set_base + "/SetForceDefrost"
	t.StateTopic = config.Mqtt_topic_base + "/Defrosting_State"
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	SwitchTopics["SetForceDefrost"] = t
	t = AutoDiscoverStruct{}
	c.Subscribe(config.Mqtt_set_base+"/SetForceSterilization", 2, HandleSetForceSterilization)
	t.Name = "TEST-SetForceSterilization"
	t.CommandTopic = config.Mqtt_set_base + "/SetForceSterilization"
	t.StateTopic = config.Mqtt_topic_base + "/Sterilization_State"
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	SwitchTopics["SetForceSterilization"] = t
	t = AutoDiscoverStruct{}
	c.Subscribe(config.Mqtt_set_base+"/SetHolidayMode", 2, HandleSetHolidayMode)
	t.Name = "TEST-SetHolidayMode"
	t.CommandTopic = config.Mqtt_set_base + "/SetHolidayMode"
	t.StateTopic = config.Mqtt_topic_base + "/Holiday_Mode_State"
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	SwitchTopics["SetHolidayMode"] = t
	t = AutoDiscoverStruct{}
	c.Subscribe(config.Mqtt_set_base+"/SetPowerfulMode", 2, HandleSetPowerfulMode)

	t.Name = "TEST-SetPowerfulMode-30min"
	t.CommandTopic = config.Mqtt_set_base + "/SetPowerfulMode"
	t.StateTopic = config.Mqtt_topic_base + "/Powerful_Mode_Time"
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	t.PayloadOn = "1"
	t.StateON = "on"
	t.StateOff = "off"
	t.ValueTemplate = `{%- if value == "1" -%} on {%- else -%} off {%- endif -%}`
	SwitchTopics["SetPowerfulMode1"] = t
	t = AutoDiscoverStruct{}
	t.Name = "TEST-SetPowerfulMode-60min"
	t.CommandTopic = config.Mqtt_set_base + "/SetPowerfulMode"
	t.StateTopic = config.Mqtt_topic_base + "/Powerful_Mode_Time"
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	t.PayloadOn = "2"
	t.StateON = "on"
	t.StateOff = "off"
	t.ValueTemplate = `{%- if value == "2" -%} on {%- else -%} off {%- endif -%}`
	SwitchTopics["SetPowerfulMode2"] = t
	t = AutoDiscoverStruct{}
	t.Name = "TEST-SetPowerfulMode-90min"
	t.CommandTopic = config.Mqtt_set_base + "/SetPowerfulMode"
	t.StateTopic = config.Mqtt_topic_base + "/Powerful_Mode_Time"
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	t.PayloadOn = "3"
	t.StateON = "on"
	t.StateOff = "off"
	t.ValueTemplate = `{%- if value == "3" -%} on {%- else -%} off {%- endif -%}`
	SwitchTopics["SetPowerfulMode3"] = t
	t = AutoDiscoverStruct{}

	c.Subscribe(config.Mqtt_set_base+"/SetDHWTemp", 2, HandleSetDHWTemp)
	c.Subscribe(config.Mqtt_set_base+"/SendRawValue", 2, HandleSendRawValue)
	if config.EnableCommand == true {
		c.Subscribe(config.Mqtt_set_base+"/OSCommand", 2, HandleOSCommand)
	}

	//Perform additional action...
}

func HandleMSGfromMQTT(mclient mqtt.Client, msg mqtt.Message) {

}

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func HandleOSCommand(mclient mqtt.Client, msg mqtt.Message) {
	var cmd *exec.Cmd
	var out2 string
	s := strings.Split(string(msg.Payload()), " ")
	if len(s) < 2 {
		cmd = exec.Command(s[0])
	} else {
		cmd = exec.Command(s[0], s[1:]...)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		// TODO: handle error more gracefully
		out2 = fmt.Sprintf("%s", err)
	}
	comout := fmt.Sprintf("%s - %s", out, out2)
	TOP := fmt.Sprintf("%s/OSCommand/out", config.Mqtt_set_base)
	fmt.Println("Publikuje do ", TOP, "warosc", string(comout))
	token := mclient.Publish(TOP, byte(0), false, comout)
	if token.Wait() && token.Error() != nil {
		fmt.Printf("Fail to publish, %v", token.Error())
	}

}

func HandleSendRawValue(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	cts := strings.TrimSpace(string(msg.Payload()))
	command, err = hex.DecodeString(cts)
	if err != nil {
		fmt.Println(err)
	}

	CommandsToSend[xid.New()] = command
}

func HandleSetOperationMode(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var set_mode byte
	a, _ := strconv.Atoi(string(msg.Payload()))

	switch a {
	case 0:
		set_mode = 82
	case 1:
		set_mode = 83
	case 2:
		set_mode = 89
	case 3:
		set_mode = 33
	case 4:
		set_mode = 98
	case 5:
		set_mode = 99
	case 6:
		set_mode = 104
	default:
		set_mode = 0
	}

	fmt.Printf("set heat pump mode to  %d", set_mode)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, set_mode, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetDHWTemp(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpump_state byte

	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	e := a + 128
	heatpump_state = byte(e)
	fmt.Printf("set DHW temperature to   %d", a)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, heatpump_state, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetPowerfulMode(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpump_state byte

	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	e := a + 73
	heatpump_state = byte(e)
	fmt.Printf("set powerful mode to  %d", a)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, heatpump_state, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetHolidayMode(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpump_state byte
	e := 16
	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	if a == 1 {
		e = 32
	}
	heatpump_state = byte(e)
	fmt.Printf("set holiday mode to  %d", heatpump_state)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, heatpump_state, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetForceSterilization(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpump_state byte
	e := 0
	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	if a == 1 {
		e = 4
	}
	heatpump_state = byte(e)
	fmt.Printf("set force sterilization  mode to %d", heatpump_state)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, heatpump_state, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetForceDefrost(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpump_state byte
	e := 0
	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	if a == 1 {
		e = 2
	}
	heatpump_state = byte(e)
	fmt.Printf("set force defrost mode to %d", heatpump_state)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, heatpump_state, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetForceDHW(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpump_state byte
	e := 64
	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	if a == 1 {
		e = 128
	}
	heatpump_state = byte(e)
	fmt.Printf("set force DHW mode to %d", heatpump_state)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, heatpump_state, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetZ1HeatRequestTemperature(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var request_temp byte
	e, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		e = int(f)
	}

	e = e + 128
	request_temp = byte(e)
	fmt.Printf("set z1 heat request temperature to %d", request_temp-128)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, request_temp, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetZ1CoolRequestTemperature(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var request_temp byte
	e, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		e = int(f)
	}
	e = e + 128
	request_temp = byte(e)
	fmt.Printf("set z1 cool request temperature to %d", request_temp-128)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, request_temp, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetZ2HeatRequestTemperature(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var request_temp byte
	e, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		e = int(f)
	}
	e = e + 128
	request_temp = byte(e)
	fmt.Printf("set z2 heat request temperature to %d", request_temp-128)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, request_temp, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetZ2CoolRequestTemperature(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var request_temp byte
	e, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		e = int(f)
	}
	e = e + 128
	request_temp = byte(e)
	fmt.Printf("set z2 cool request temperature to %d", request_temp-128)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, request_temp, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetQuietMode(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var quiet_mode byte

	e, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		e = int(f)
	}
	e = (e + 1) * 8

	quiet_mode = byte(e)
	fmt.Printf("set Quiet mode to %d", quiet_mode/8-1)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, quiet_mode, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func HandleSetHeatpump(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpump_state byte

	e := 1
	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}
	if a == 1 {
		e = 2
	}

	heatpump_state = byte(e)
	fmt.Printf("set heatpump state to %d", heatpump_state)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, heatpump_state, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	CommandsToSend[xid.New()] = command
}

func log_message(a string) {
	fmt.Println(a)
}

func logHex(command []byte, length int) {
	fmt.Printf("% X \n", string(command))

}

func calcChecksum(command []byte, length int) byte {
	var chk byte
	chk = 0
	for i := 0; i < length; i++ {
		chk += command[i]
	}
	chk = (chk ^ 0xFF) + 01
	return chk
}

func ParseTopicList2() {

	// Loop through lines & turn into object
	for key, _ := range AllTopics {
		var data TopicData
		if _, err := toml.DecodeFile(configfile, &data); err != nil {
			log.Fatal(err)
		}
		AllTopics[key] = data
		//a	fmt.Println(data)
		//EncodeTopicsToTOML(TNUM, data)

	}
}

func send_command(command []byte, length int) bool {

	var chk byte
	chk = calcChecksum(command, length)
	var bytesSent int

	bytesSent, err := Serial.Write(command) //first send command
	_, err = Serial.Write([]byte{chk})      //then calculcated checksum byte afterwards
	if err != nil {
		fmt.Println(err)
	}
	log_msg := fmt.Sprintf("sent bytes: %d with checksum: %d ", bytesSent, int(chk))
	log_message(log_msg)

	if config.Loghex == true {
		logHex(command, length)
	}
	//readSerial()
	//allowreadtime = millis() + SERIALTIMEOUT //set allowreadtime when to timeout the answer of this command
	return true
}

// func pushCommandBuffer(command []byte , length int) {
// 	if (commandsInBuffer < MAXCOMMANDSINBUFFER) {
// 	  command_struct* newCommand = new command_struct;
// 	  newCommand->length = length;
// 	  for (int i = 0 ; i < length ; i++) {
// 		newCommand->value[i] = command[i];
// 	  }
// 	  newCommand->next = commandBuffer;
// 	  commandBuffer = newCommand;
// 	  commandsInBuffer++;
// 	}
// 	else {
// 	  log_message("Too much commands already in buffer. Ignoring this commands.");
// 	}
//   }

func readSerial(MC mqtt.Client, MT mqtt.Token) bool {

	data_length := 203

	totalreads++
	data := make([]byte, data_length)
	n, err := Serial.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	if n == 0 {
		fmt.Println("\nEOF")

	}

	//panasonic read is always 203 on valid receive, if not yet there wait for next read
	log_message("Received 203 bytes data\n")
	if config.Loghex {
		logHex(data, data_length)
	}
	if !isValidReceiveHeader(data) {
		log_message("Received wrong header!\n")
		data_length = 0 //for next attempt;
		return false
	}
	if !isValidReceiveChecksum(data) {
		log_message("Checksum received false!")
		data_length = 0 //for next attempt
		return false
	}
	log_message("Checksum and header received ok!")
	data_length = 0 //for next attempt
	goodreads++
	readpercentage = ((goodreads / totalreads) * 100)
	log_msg := fmt.Sprintf("Total reads : %f and total good reads : %f (%.2f %%)", totalreads, goodreads, readpercentage)
	log_message(log_msg)
	decode_heatpump_data(data, MC, MT)
	token := MC.Publish(fmt.Sprintf("%s/LWT", config.Mqtt_set_base), byte(0), false, "Online")
	if token.Wait() && token.Error() != nil {
		fmt.Printf("Fail to publish, %v", token.Error())
	}
	return true

}

func isValidReceiveHeader(data []byte) bool {
	return ((data[0] == 0x71) && (data[1] == 0xC8) && (data[2] == 0x01) && (data[3] == 0x10))
}

func isValidReceiveChecksum(data []byte) bool {
	var chk byte
	chk = 0
	for i := 0; i < len(data); i++ {
		chk += data[i]
	}
	return (chk == 0) //all received bytes + checksum should result in 0
}

func CallTopicFunction(data byte, f func(data byte) string) string {
	return f(data)
}

func getBit7and8(input byte) string {
	return fmt.Sprintf("%d", (input&0b11)-1)
}

func getBit3and4and5(input byte) string {
	return fmt.Sprintf("%d", ((input>>3)&0b111)-1)
}

func getIntMinus1Times10(input byte) string {
	value := int(input) - 1
	return fmt.Sprintf("%d", value*10)

}

func getIntMinus1Times50(input byte) string {
	value := int(input) - 1
	return fmt.Sprintf("%d", value*50)

}

func unknown(input byte) string {
	return "-1"
}

func getIntMinus128(input byte) string {
	value := int(input) - 128
	return fmt.Sprintf("%d", value)
}

func getIntMinus1Div5(input byte) string {
	value := int(input) - 1
	var out float32
	out = float32(value) / 5
	return fmt.Sprintf("%.2f", out)

}

func getRight3bits(input byte) string {
	return fmt.Sprintf("%d", (input&0b111)-1)

}

func getBit1and2(input byte) string {
	return fmt.Sprintf("%d", (input>>6)-1)

}

func getOpMode(input byte) string {
	switch int(input) {
	case 82:
		return "0"
	case 83:
		return "1"
	case 89:
		return "2"
	case 97:
		return "3"
	case 98:
		return "4"
	case 99:
		return "5"
	case 105:
		return "6"
	case 90:
		return "7"
	case 106:
		return "8"
	default:
		return "-1"
	}
}

func getIntMinus1(input byte) string {
	value := int(input) - 1
	return fmt.Sprintf("%d", value)
}

func getEnergy(input byte) string {
	value := (int(input) - 1) * 200
	return fmt.Sprintf("%d", value)
}

func getBit3and4(input byte) string {
	return fmt.Sprintf("%d", ((input>>4)&0b11)-1)

}

func getBit5and6(input byte) string {
	return fmt.Sprintf("%d", ((input>>2)&0b11)-1)

}

func getPumpFlow(data []byte) string { // TOP1 //
	PumpFlow1 := int(data[170])
	PumpFlow2 := ((float64(data[169]) - 1) / 256)
	PumpFlow := float64(PumpFlow1) + PumpFlow2
	//return String(PumpFlow,2);
	return fmt.Sprintf("%.2f", PumpFlow)
}

func getErrorInfo(data []byte) string { // TOP44 //
	Error_type := int(data[113])
	Error_number := int(data[114]) - 17
	var Error_string string
	switch Error_type {
	case 177: //B1=F type error
		Error_string = fmt.Sprintf("F%02X", Error_number)

	case 161: //A1=H type error
		Error_string = fmt.Sprintf("H%02X", Error_number)

	default:
		Error_string = fmt.Sprintf("No error")

	}
	return Error_string
}

func decode_heatpump_data(data []byte, mclient mqtt.Client, token mqtt.Token) {

	var updatenow bool = false
	m := map[string]func(byte) string{
		"getBit7and8":         getBit7and8,
		"unknown":             unknown,
		"getRight3bits":       getRight3bits,
		"getIntMinus1Div5":    getIntMinus1Div5,
		"getIntMinus1Times50": getIntMinus1Times50,
		"getIntMinus1Times10": getIntMinus1Times10,
		"getBit3and4and5":     getBit3and4and5,
		"getIntMinus128":      getIntMinus128,
		"getBit1and2":         getBit1and2,
		"getOpMode":           getOpMode,
		"getIntMinus1":        getIntMinus1,
		"getEnergy":           getEnergy,
		"getBit5and6":         getBit5and6,

		"getBit3and4": getBit3and4,
	}

	// 	if (millis() > nextalldatatime) {
	// 	  updatenow = true;
	// 	  nextalldatatime = millis() + UPDATEALLTIME;
	// 	}
	for k, v := range AllTopics {
		var Input_Byte byte
		var Topic_Value string
		var value string
		switch k {
		case 1:
			Topic_Value = getPumpFlow(data)
		case 11:
			d := make([]byte, 2)
			d[0] = data[183]
			d[1] = data[182]
			Topic_Value = fmt.Sprintf("%d", int(binary.BigEndian.Uint16(d))-1)
		case 12:
			d := make([]byte, 2)
			d[0] = data[180]
			d[1] = data[179]
			Topic_Value = fmt.Sprintf("%d", int(binary.BigEndian.Uint16(d))-1)
		case 90:
			d := make([]byte, 2)
			d[0] = data[186]
			d[1] = data[185]
			Topic_Value = fmt.Sprintf("%d", int(binary.BigEndian.Uint16(d))-1)
		case 91:
			d := make([]byte, 2)
			d[0] = data[189]
			d[1] = data[188]
			Topic_Value = fmt.Sprintf("%d", int(binary.BigEndian.Uint16(d))-1)
		case 44:
			Topic_Value = getErrorInfo(data)
		default:
			Input_Byte = data[v.TopicBit]
			if _, ok := m[v.TopicFunction]; ok {
				Topic_Value = CallTopicFunction(Input_Byte, m[v.TopicFunction])
			} else {
				fmt.Println("NIE MA FUNKCJI", v.TopicFunction)
			}

		}

		if (updatenow) || (actData[k] != Topic_Value) {
			actData[k] = Topic_Value
			fmt.Printf("received TOP%d %s: %s \n", k, v.TopicName, Topic_Value)
			if config.Aquarea2mqttCompatible {
				TOP := "aquarea/state/" + fmt.Sprintf("%s/%s", config.Aquarea2mqttPumpID, v.TopicA2M)
				value = strings.TrimSpace(Topic_Value)
				value = strings.ToUpper(Topic_Value)
				fmt.Println("Publikuje do ", TOP, "warosc", string(value))
				token = mclient.Publish(TOP, byte(0), false, value)
				if token.Wait() && token.Error() != nil {
					fmt.Printf("Fail to publish, %v", token.Error())
				}
			}
			TOP := fmt.Sprintf("%s/%s", config.Mqtt_topic_base, v.TopicName)
			fmt.Println("Publikuje do ", TOP, "warosc", string(Topic_Value))
			token = mclient.Publish(TOP, byte(0), false, Topic_Value)
			if token.Wait() && token.Error() != nil {
				fmt.Printf("Fail to publish, %v", token.Error())
			}

		}

	}

}

func ParseTopicList3() {
	AllTopics[0].TopicNumber = 0
	AllTopics[0].TopicName = "Heatpump_State"
	AllTopics[0].TopicType = "binary_sensor"
	AllTopics[0].TopicBit = 4
	AllTopics[0].TopicFunction = "getBit7and8"
	AllTopics[0].TopicUnit = "OffOn"
	AllTopics[0].TopicA2M = "RunningStatus"
	AllTopics[1].TopicNumber = 1
	AllTopics[1].TopicName = "Pump_Flow"
	AllTopics[1].TopicBit = 0
	AllTopics[1].TopicDisplayUnit = "L/m"

	AllTopics[1].TopicFunction = "unknown"
	AllTopics[1].TopicUnit = "LitersPerMin"
	AllTopics[1].TopicA2M = "WaterFlow"
	AllTopics[10].TopicNumber = 10
	AllTopics[10].TopicName = "DHW_Temp"
	AllTopics[10].TopicBit = 141
	AllTopics[10].TopicDisplayUnit = "°C"

	AllTopics[10].TopicFunction = "getIntMinus128"
	AllTopics[10].TopicUnit = "Celsius"
	AllTopics[10].TopicA2M = "DailyWaterTankActualTemperature"
	AllTopics[11].TopicNumber = 11
	AllTopics[11].TopicName = "Operations_Hours"
	AllTopics[11].TopicBit = 0
	AllTopics[11].TopicFunction = "unknown"
	AllTopics[11].TopicDisplayUnit = "h"

	AllTopics[11].TopicUnit = "Hours"
	AllTopics[11].TopicA2M = ""
	AllTopics[12].TopicNumber = 12
	AllTopics[12].TopicName = "Operations_Counter"
	AllTopics[12].TopicBit = 0
	AllTopics[12].TopicFunction = "unknown"
	AllTopics[12].TopicUnit = "Counter"
	AllTopics[12].TopicA2M = ""
	AllTopics[13].TopicNumber = 13
	AllTopics[13].TopicName = "Main_Schedule_State"
	AllTopics[13].TopicBit = 5
	AllTopics[13].TopicType = "binary_sensor"
	AllTopics[13].TopicFunction = "getBit1and2"
	AllTopics[13].TopicUnit = "DisabledEnabled"
	AllTopics[13].TopicA2M = ""
	AllTopics[14].TopicNumber = 14
	AllTopics[14].TopicName = "Outside_Temp"
	AllTopics[14].TopicDisplayUnit = "°C"

	AllTopics[14].TopicBit = 142
	AllTopics[14].TopicFunction = "getIntMinus128"
	AllTopics[14].TopicUnit = "Celsius"
	AllTopics[14].TopicA2M = "OutdoorTemperature"
	AllTopics[15].TopicNumber = 15
	AllTopics[15].TopicName = "Heat_Energy_Production"
	AllTopics[15].TopicBit = 194
	AllTopics[15].TopicDisplayUnit = "W"

	AllTopics[15].TopicFunction = "getEnergy"
	AllTopics[15].TopicUnit = "Watt"
	AllTopics[15].TopicA2M = ""
	AllTopics[16].TopicNumber = 16
	AllTopics[16].TopicName = "Heat_Energy_Consumption"
	AllTopics[16].TopicBit = 193
	AllTopics[16].TopicFunction = "getEnergy"
	AllTopics[16].TopicDisplayUnit = "W"
	AllTopics[16].TopicUnit = "Watt"
	AllTopics[16].TopicA2M = ""
	AllTopics[17].TopicNumber = 17
	AllTopics[17].TopicName = "Powerful_Mode_Time"
	AllTopics[17].TopicBit = 7
	AllTopics[17].TopicDisplayUnit = "Min"
	AllTopics[17].TopicValueTemplate = `{{ (value | int) * 30 }}`

	AllTopics[17].TopicFunction = "getRight3bits"
	AllTopics[17].TopicUnit = "Powerfulmode"
	AllTopics[17].TopicA2M = ""
	AllTopics[18].TopicNumber = 18
	AllTopics[18].TopicName = "Quiet_Mode_Level"
	AllTopics[18].TopicBit = 7
	AllTopics[18].TopicFunction = "getBit3and4and5"
	AllTopics[18].TopicUnit = "Quietmode"
	AllTopics[18].TopicValueTemplate = `{%- if value == "4" -%} Scheduled {%- else -%} {{ value }} {%- endif -%}`

	AllTopics[18].TopicA2M = ""
	AllTopics[19].TopicNumber = 19
	AllTopics[19].TopicName = "Holiday_Mode_State"
	AllTopics[19].TopicBit = 5
	AllTopics[19].TopicType = "binary_sensor"
	AllTopics[19].TopicFunction = "getBit3and4"
	AllTopics[19].TopicUnit = "HolidayState"
	AllTopics[19].TopicA2M = ""
	AllTopics[2].TopicNumber = 2
	AllTopics[2].TopicName = "Force_DHW_State"
	AllTopics[2].TopicBit = 4
	AllTopics[2].TopicType = "binary_sensor"

	AllTopics[2].TopicFunction = "getBit1and2"
	AllTopics[2].TopicUnit = "DisabledEnabled"
	AllTopics[2].TopicA2M = ""
	AllTopics[20].TopicNumber = 20
	AllTopics[20].TopicName = "ThreeWay_Valve_State"
	AllTopics[20].TopicBit = 111
	AllTopics[20].TopicFunction = "getBit7and8"
	AllTopics[20].TopicValueTemplate = `{%- if value == "0" -%} Room {%- elif value == "1" -%} Tank {%- endif -%}`

	AllTopics[20].TopicUnit = "Valve"
	AllTopics[20].TopicA2M = ""
	AllTopics[21].TopicNumber = 21
	AllTopics[21].TopicName = "Outside_Pipe_Temp"
	AllTopics[21].TopicBit = 158
	AllTopics[21].TopicFunction = "getIntMinus128"
	AllTopics[21].TopicUnit = "Celsius"
	AllTopics[21].TopicDisplayUnit = "°C"

	AllTopics[21].TopicA2M = ""
	AllTopics[22].TopicNumber = 22
	AllTopics[22].TopicName = "DHW_Heat_Delta"
	AllTopics[22].TopicBit = 99
	AllTopics[22].TopicFunction = "getIntMinus128"
	AllTopics[22].TopicUnit = "Kelvin"
	AllTopics[22].TopicDisplayUnit = "°K"

	AllTopics[22].TopicA2M = ""
	AllTopics[23].TopicNumber = 23
	AllTopics[23].TopicName = "Heat_Delta"
	AllTopics[23].TopicBit = 84
	AllTopics[23].TopicFunction = "getIntMinus128"
	AllTopics[23].TopicUnit = "Kelvin"
	AllTopics[23].TopicDisplayUnit = "°K"

	AllTopics[23].TopicA2M = ""
	AllTopics[24].TopicNumber = 24
	AllTopics[24].TopicName = "Cool_Delta"
	AllTopics[24].TopicBit = 94
	AllTopics[24].TopicFunction = "getIntMinus128"
	AllTopics[24].TopicUnit = "Kelvin"
	AllTopics[24].TopicA2M = ""
	AllTopics[24].TopicDisplayUnit = "°K"

	AllTopics[25].TopicNumber = 25
	AllTopics[25].TopicName = "DHW_Holiday_Shift_Temp"
	AllTopics[25].TopicBit = 44
	AllTopics[25].TopicFunction = "getIntMinus128"
	AllTopics[25].TopicUnit = "Kelvin"
	AllTopics[25].TopicDisplayUnit = "°K"

	AllTopics[25].TopicA2M = ""
	AllTopics[26].TopicNumber = 26
	AllTopics[26].TopicName = "Defrosting_State"
	AllTopics[26].TopicType = "binary_sensor"

	AllTopics[26].TopicBit = 111
	AllTopics[26].TopicFunction = "getBit5and6"
	AllTopics[26].TopicUnit = "DisabledEnabled"
	AllTopics[26].TopicA2M = ""
	AllTopics[27].TopicNumber = 27
	AllTopics[27].TopicName = "Z1_Heat_Request_Temp"
	AllTopics[27].TopicBit = 38
	AllTopics[27].TopicDisplayUnit = "°C"

	AllTopics[27].TopicFunction = "getIntMinus128"
	AllTopics[27].TopicUnit = "Celsius"
	AllTopics[27].TopicA2M = "Zone1SetpointTemperature"
	AllTopics[28].TopicNumber = 28
	AllTopics[28].TopicName = "Z1_Cool_Request_Temp"
	AllTopics[28].TopicBit = 39
	AllTopics[28].TopicFunction = "getIntMinus128"
	AllTopics[28].TopicDisplayUnit = "°C"

	AllTopics[28].TopicUnit = "Celsius"
	AllTopics[28].TopicA2M = ""
	AllTopics[29].TopicNumber = 29
	AllTopics[29].TopicName = "Z1_Heat_Curve_Target_High_Temp"
	AllTopics[29].TopicBit = 75
	AllTopics[29].TopicDisplayUnit = "°C"

	AllTopics[29].TopicFunction = "getIntMinus128"
	AllTopics[29].TopicUnit = "Celsius"
	AllTopics[29].TopicA2M = ""
	AllTopics[3].TopicNumber = 3
	AllTopics[3].TopicName = "Quiet_Mode_Schedule"
	AllTopics[3].TopicBit = 7
	AllTopics[3].TopicType = "binary_sensor"
	AllTopics[3].TopicFunction = "getBit1and2"
	AllTopics[3].TopicUnit = "DisabledEnabled"
	AllTopics[3].TopicA2M = ""
	AllTopics[30].TopicNumber = 30
	AllTopics[30].TopicName = "Z1_Heat_Curve_Target_Low_Temp"
	AllTopics[30].TopicBit = 76
	AllTopics[30].TopicDisplayUnit = "°C"

	AllTopics[30].TopicFunction = "getIntMinus128"
	AllTopics[30].TopicUnit = "Celsius"
	AllTopics[30].TopicA2M = ""
	AllTopics[31].TopicNumber = 31
	AllTopics[31].TopicName = "Z1_Heat_Curve_Outside_High_Temp"
	AllTopics[31].TopicBit = 78
	AllTopics[31].TopicDisplayUnit = "°C"

	AllTopics[31].TopicFunction = "getIntMinus128"
	AllTopics[31].TopicUnit = "Celsius"
	AllTopics[31].TopicA2M = ""
	AllTopics[32].TopicNumber = 32
	AllTopics[32].TopicName = "Z1_Heat_Curve_Outside_Low_Temp"
	AllTopics[32].TopicBit = 77
	AllTopics[32].TopicDisplayUnit = "°C"

	AllTopics[32].TopicFunction = "getIntMinus128"
	AllTopics[32].TopicUnit = "Celsius"
	AllTopics[32].TopicA2M = ""
	AllTopics[33].TopicNumber = 33
	AllTopics[33].TopicName = "Room_Thermostat_Temp"
	AllTopics[33].TopicBit = 156
	AllTopics[33].TopicDisplayUnit = "°C"

	AllTopics[33].TopicFunction = "getIntMinus128"
	AllTopics[33].TopicUnit = "Celsius"
	AllTopics[33].TopicA2M = ""
	AllTopics[34].TopicNumber = 34
	AllTopics[34].TopicName = "Z2_Heat_Request_Temp"
	AllTopics[34].TopicBit = 40
	AllTopics[34].TopicDisplayUnit = "°C"

	AllTopics[34].TopicFunction = "getIntMinus128"
	AllTopics[34].TopicUnit = "Celsius"
	AllTopics[34].TopicA2M = "Zone2SetpointTemperature"
	AllTopics[35].TopicNumber = 35
	AllTopics[35].TopicName = "Z2_Cool_Request_Temp"
	AllTopics[35].TopicBit = 41
	AllTopics[35].TopicDisplayUnit = "°C"

	AllTopics[35].TopicFunction = "getIntMinus128"
	AllTopics[35].TopicUnit = "Celsius"
	AllTopics[35].TopicA2M = ""
	AllTopics[36].TopicNumber = 36
	AllTopics[36].TopicName = "Z1_Water_Temp"
	AllTopics[36].TopicBit = 145
	AllTopics[36].TopicFunction = "getIntMinus128"
	AllTopics[36].TopicUnit = "Celsius"
	AllTopics[36].TopicDisplayUnit = "°C"

	AllTopics[36].TopicA2M = "Zone1WaterTemperature"
	AllTopics[37].TopicNumber = 37
	AllTopics[37].TopicName = "Z2_Water_Temp"
	AllTopics[37].TopicBit = 146
	AllTopics[37].TopicFunction = "getIntMinus128"
	AllTopics[37].TopicUnit = "Celsius"
	AllTopics[37].TopicDisplayUnit = "°C"

	AllTopics[37].TopicA2M = "Zone2WaterTemperature"
	AllTopics[38].TopicNumber = 38
	AllTopics[38].TopicName = "Cool_Energy_Production"
	AllTopics[38].TopicBit = 196
	AllTopics[38].TopicDisplayUnit = "W"

	AllTopics[38].TopicFunction = "getEnergy"
	AllTopics[38].TopicUnit = "Watt"
	AllTopics[38].TopicA2M = ""
	AllTopics[39].TopicNumber = 39
	AllTopics[39].TopicName = "Cool_Energy_Consumption"
	AllTopics[39].TopicBit = 195
	AllTopics[39].TopicDisplayUnit = "W"

	AllTopics[39].TopicFunction = "getEnergy"
	AllTopics[39].TopicUnit = "Watt"
	AllTopics[39].TopicA2M = ""
	AllTopics[4].TopicNumber = 4
	AllTopics[4].TopicName = "Operating_Mode_State"
	AllTopics[4].TopicBit = 6
	AllTopics[4].TopicValueTemplate = `{%- if value == "0" -%} Heat {%- elif value == "1" -%} Cool {%- elif value == "2" -%} Auto(Heat) {%- elif value == "3" -%} DHW {%- elif value == "4" -%} Heat+DHW {%- elif value == "5" -%} Cool+DHW {%- elif value == "6" -%} Auto(Heat)+DHW {%- elif value == "7" -%} Auto(Cool) {%- elif value == "8" -%} Auto(Cool)+DHW {%- endif -%}`
	AllTopics[4].TopicFunction = "getOpMode"
	AllTopics[4].TopicUnit = "OpModeDesc"
	AllTopics[4].TopicA2M = "WorkingMode"
	AllTopics[40].TopicNumber = 40
	AllTopics[40].TopicName = "DHW_Energy_Production"
	AllTopics[40].TopicBit = 198
	AllTopics[40].TopicFunction = "getEnergy"
	AllTopics[40].TopicUnit = "Watt"
	AllTopics[40].TopicDisplayUnit = "W"

	AllTopics[40].TopicA2M = ""
	AllTopics[41].TopicNumber = 41
	AllTopics[41].TopicName = "DHW_Energy_Consumption"
	AllTopics[41].TopicBit = 197
	AllTopics[41].TopicFunction = "getEnergy"
	AllTopics[41].TopicUnit = "Watt"
	AllTopics[41].TopicDisplayUnit = "W"

	AllTopics[41].TopicA2M = ""
	AllTopics[42].TopicNumber = 42
	AllTopics[42].TopicName = "Z1_Water_Target_Temp"
	AllTopics[42].TopicBit = 147
	AllTopics[42].TopicFunction = "getIntMinus128"
	AllTopics[42].TopicUnit = "Celsius"
	AllTopics[42].TopicA2M = ""
	AllTopics[42].TopicDisplayUnit = "°C"

	AllTopics[43].TopicNumber = 43
	AllTopics[43].TopicName = "Z2_Water_Target_Temp"
	AllTopics[43].TopicBit = 148
	AllTopics[43].TopicFunction = "getIntMinus128"
	AllTopics[43].TopicUnit = "Celsius"
	AllTopics[43].TopicDisplayUnit = "°C"

	AllTopics[43].TopicA2M = ""
	AllTopics[44].TopicNumber = 44
	AllTopics[44].TopicName = "Error"
	AllTopics[44].TopicBit = 0
	AllTopics[44].TopicFunction = "unknown"
	AllTopics[44].TopicUnit = "ErrorState"
	AllTopics[44].TopicA2M = ""
	AllTopics[45].TopicNumber = 45
	AllTopics[45].TopicName = "Room_Holiday_Shift_Temp"
	AllTopics[45].TopicBit = 43
	AllTopics[45].TopicFunction = "getIntMinus128"
	AllTopics[45].TopicUnit = "Kelvin"
	AllTopics[45].TopicDisplayUnit = "°K"

	AllTopics[45].TopicA2M = ""
	AllTopics[46].TopicNumber = 46
	AllTopics[46].TopicName = "Buffer_Temp"
	AllTopics[46].TopicBit = 149
	AllTopics[46].TopicFunction = "getIntMinus128"
	AllTopics[46].TopicUnit = "Celsius"
	AllTopics[46].TopicDisplayUnit = "°C"

	AllTopics[46].TopicA2M = "BufferTankTemperature"
	AllTopics[47].TopicNumber = 47
	AllTopics[47].TopicName = "Solar_Temp"
	AllTopics[47].TopicBit = 150
	AllTopics[47].TopicFunction = "getIntMinus128"
	AllTopics[47].TopicUnit = "Celsius"
	AllTopics[47].TopicDisplayUnit = "°C"

	AllTopics[47].TopicA2M = ""
	AllTopics[48].TopicNumber = 48
	AllTopics[48].TopicName = "Pool_Temp"
	AllTopics[48].TopicBit = 151
	AllTopics[48].TopicFunction = "getIntMinus128"
	AllTopics[48].TopicUnit = "Celsius"
	AllTopics[48].TopicDisplayUnit = "°C"

	AllTopics[48].TopicA2M = ""
	AllTopics[49].TopicNumber = 49
	AllTopics[49].TopicName = "Main_Hex_Outlet_Temp"
	AllTopics[49].TopicBit = 154
	AllTopics[49].TopicDisplayUnit = "°C"

	AllTopics[49].TopicFunction = "getIntMinus128"
	AllTopics[49].TopicUnit = "Celsius"
	AllTopics[49].TopicA2M = ""
	AllTopics[5].TopicNumber = 5
	AllTopics[5].TopicName = "Main_Inlet_Temp"
	AllTopics[5].TopicBit = 143
	AllTopics[5].TopicFunction = "getIntMinus128"
	AllTopics[5].TopicUnit = "Celsius"
	AllTopics[5].TopicDisplayUnit = "°C"
	AllTopics[5].TopicA2M = "WaterInleet"
	AllTopics[50].TopicNumber = 50
	AllTopics[50].TopicName = "Discharge_Temp"
	AllTopics[50].TopicBit = 155
	AllTopics[50].TopicFunction = "getIntMinus128"
	AllTopics[50].TopicUnit = "Celsius"
	AllTopics[50].TopicDisplayUnit = "°C"

	AllTopics[50].TopicA2M = ""
	AllTopics[51].TopicNumber = 51
	AllTopics[51].TopicName = "Inside_Pipe_Temp"
	AllTopics[51].TopicBit = 157
	AllTopics[51].TopicFunction = "getIntMinus128"
	AllTopics[51].TopicUnit = "Celsius"
	AllTopics[51].TopicDisplayUnit = "°C"

	AllTopics[51].TopicA2M = ""
	AllTopics[52].TopicNumber = 52
	AllTopics[52].TopicName = "Defrost_Temp"
	AllTopics[52].TopicBit = 159
	AllTopics[52].TopicFunction = "getIntMinus128"
	AllTopics[52].TopicUnit = "Celsius"
	AllTopics[52].TopicA2M = ""
	AllTopics[52].TopicDisplayUnit = "°C"

	AllTopics[53].TopicNumber = 53
	AllTopics[53].TopicDisplayUnit = "°C"

	AllTopics[53].TopicName = "Eva_Outlet_Temp"
	AllTopics[53].TopicBit = 160
	AllTopics[53].TopicFunction = "getIntMinus128"
	AllTopics[53].TopicUnit = "Celsius"
	AllTopics[53].TopicA2M = ""
	AllTopics[54].TopicNumber = 54
	AllTopics[54].TopicName = "Bypass_Outlet_Temp"
	AllTopics[54].TopicBit = 161
	AllTopics[54].TopicDisplayUnit = "°C"

	AllTopics[54].TopicFunction = "getIntMinus128"
	AllTopics[54].TopicUnit = "Celsius"
	AllTopics[54].TopicA2M = ""
	AllTopics[55].TopicNumber = 55
	AllTopics[55].TopicName = "Ipm_Temp"
	AllTopics[55].TopicBit = 162
	AllTopics[55].TopicDisplayUnit = "°C"

	AllTopics[55].TopicFunction = "getIntMinus128"
	AllTopics[55].TopicUnit = "Celsius"
	AllTopics[55].TopicA2M = ""
	AllTopics[56].TopicNumber = 56
	AllTopics[56].TopicName = "Z1_Temp"
	AllTopics[56].TopicBit = 139
	AllTopics[56].TopicFunction = "getIntMinus128"
	AllTopics[56].TopicUnit = "Celsius"
	AllTopics[56].TopicDisplayUnit = "°C"

	AllTopics[56].TopicA2M = "Zone1ActualTemperature"
	AllTopics[57].TopicNumber = 57
	AllTopics[57].TopicName = "Z2_Temp"
	AllTopics[57].TopicBit = 140
	AllTopics[57].TopicFunction = "getIntMinus128"
	AllTopics[57].TopicUnit = "Celsius"
	AllTopics[57].TopicDisplayUnit = "°C"

	AllTopics[57].TopicA2M = "Zone2ActualTemperature"
	AllTopics[58].TopicNumber = 58
	AllTopics[58].TopicName = "DHW_Heater_State"
	AllTopics[58].TopicBit = 9
	AllTopics[58].TopicType = "binary_sensor"

	AllTopics[58].TopicFunction = "getBit5and6"
	AllTopics[58].TopicUnit = "BlockedFree"
	AllTopics[58].TopicA2M = ""
	AllTopics[59].TopicNumber = 59
	AllTopics[59].TopicName = "Room_Heater_State"
	AllTopics[59].TopicBit = 9
	AllTopics[59].TopicType = "binary_sensor"

	AllTopics[59].TopicFunction = "getBit7and8"
	AllTopics[59].TopicUnit = "BlockedFree"
	AllTopics[59].TopicA2M = ""
	AllTopics[6].TopicNumber = 6
	AllTopics[6].TopicName = "Main_Outlet_Temp"
	AllTopics[6].TopicBit = 144
	AllTopics[6].TopicFunction = "getIntMinus128"
	AllTopics[6].TopicUnit = "Celsius"
	AllTopics[6].TopicDisplayUnit = "°C"

	AllTopics[6].TopicA2M = "WaterOutleet"
	AllTopics[60].TopicNumber = 60
	AllTopics[60].TopicType = "binary_sensor"

	AllTopics[60].TopicName = "Internal_Heater_State"
	AllTopics[60].TopicBit = 112
	AllTopics[60].TopicFunction = "getBit7and8"
	AllTopics[60].TopicUnit = "InactiveActive"
	AllTopics[60].TopicA2M = ""
	AllTopics[61].TopicNumber = 61
	AllTopics[61].TopicName = "External_Heater_State"
	AllTopics[61].TopicBit = 112
	AllTopics[61].TopicFunction = "getBit5and6"
	AllTopics[61].TopicUnit = "InactiveActive"
	AllTopics[61].TopicA2M = ""
	AllTopics[61].TopicType = "binary_sensor"

	AllTopics[62].TopicNumber = 62
	AllTopics[62].TopicName = "Fan1_Motor_Speed"
	AllTopics[62].TopicBit = 173
	AllTopics[62].TopicDisplayUnit = "R/min"

	AllTopics[62].TopicFunction = "getIntMinus1Times10"
	AllTopics[62].TopicUnit = "RotationsPerMin"
	AllTopics[62].TopicA2M = ""
	AllTopics[63].TopicNumber = 63
	AllTopics[63].TopicName = "Fan2_Motor_Speed"
	AllTopics[63].TopicBit = 174
	AllTopics[63].TopicDisplayUnit = "R/min"
	AllTopics[63].TopicFunction = "getIntMinus1Times10"
	AllTopics[63].TopicUnit = "RotationsPerMin"
	AllTopics[63].TopicA2M = ""
	AllTopics[64].TopicNumber = 64
	AllTopics[64].TopicName = "High_Pressure"
	AllTopics[64].TopicBit = 163
	AllTopics[64].TopicDisplayUnit = "Kgf/cm2"
	AllTopics[64].TopicFunction = "getIntMinus1Div5"
	AllTopics[64].TopicUnit = "Pressure"
	AllTopics[64].TopicA2M = ""
	AllTopics[65].TopicNumber = 65
	AllTopics[65].TopicDisplayUnit = "R/mini"

	AllTopics[65].TopicName = "Pump_Speed"
	AllTopics[65].TopicBit = 171
	AllTopics[65].TopicFunction = "getIntMinus1Times50"
	AllTopics[65].TopicUnit = "RotationsPerMin"
	AllTopics[65].TopicA2M = "PumpSpeed"
	AllTopics[66].TopicNumber = 66
	AllTopics[66].TopicName = "Low_Pressure"
	AllTopics[66].TopicBit = 164
	AllTopics[66].TopicDisplayUnit = "Kgf/cm2"

	AllTopics[66].TopicFunction = "getIntMinus1"
	AllTopics[66].TopicUnit = "Pressure"
	AllTopics[66].TopicA2M = ""
	AllTopics[67].TopicNumber = 67
	AllTopics[67].TopicName = "Compressor_Current"
	AllTopics[67].TopicBit = 165
	AllTopics[67].TopicDisplayUnit = "A"

	AllTopics[67].TopicFunction = "getIntMinus1Div5"
	AllTopics[67].TopicUnit = "Ampere"
	AllTopics[67].TopicA2M = ""
	AllTopics[68].TopicNumber = 68
	AllTopics[68].TopicName = "Force_Heater_State"
	AllTopics[68].TopicBit = 5
	AllTopics[68].TopicType = "binary_sensor"
	AllTopics[68].TopicFunction = "getBit5and6"
	AllTopics[68].TopicUnit = "InactiveActive"
	AllTopics[68].TopicA2M = ""
	AllTopics[69].TopicNumber = 69
	AllTopics[69].TopicName = "Sterilization_State"
	AllTopics[69].TopicBit = 117
	AllTopics[69].TopicType = "binary_sensor"
	AllTopics[69].TopicFunction = "getBit5and6"
	AllTopics[69].TopicUnit = "InactiveActive"
	AllTopics[69].TopicA2M = ""
	AllTopics[7].TopicNumber = 7
	AllTopics[7].TopicName = "Main_Target_Temp"
	AllTopics[7].TopicBit = 153
	AllTopics[7].TopicFunction = "getIntMinus128"
	AllTopics[7].TopicUnit = "Celsius"
	AllTopics[7].TopicDisplayUnit = "°C"

	AllTopics[7].TopicA2M = ""
	AllTopics[70].TopicNumber = 70
	AllTopics[70].TopicName = "Sterilization_Temp"
	AllTopics[70].TopicBit = 100
	AllTopics[70].TopicDisplayUnit = "°C"

	AllTopics[70].TopicFunction = "getIntMinus128"
	AllTopics[70].TopicUnit = "Celsius"
	AllTopics[70].TopicA2M = ""
	AllTopics[71].TopicNumber = 71
	AllTopics[71].TopicName = "Sterilization_Max_Time"
	AllTopics[71].TopicBit = 101
	AllTopics[71].TopicFunction = "getIntMinus1"
	AllTopics[71].TopicUnit = "Minutes"
	AllTopics[71].TopicDisplayUnit = "min"

	AllTopics[71].TopicA2M = ""
	AllTopics[72].TopicNumber = 72
	AllTopics[72].TopicName = "Z1_Cool_Curve_Target_High_Temp"
	AllTopics[72].TopicBit = 86
	AllTopics[72].TopicFunction = "getIntMinus128"
	AllTopics[72].TopicUnit = "Celsius"
	AllTopics[72].TopicA2M = ""
	AllTopics[72].TopicDisplayUnit = "°C"

	AllTopics[73].TopicNumber = 73
	AllTopics[73].TopicName = "Z1_Cool_Curve_Target_Low_Temp"
	AllTopics[73].TopicBit = 87
	AllTopics[73].TopicFunction = "getIntMinus128"
	AllTopics[73].TopicUnit = "Celsius"
	AllTopics[73].TopicDisplayUnit = "°C"

	AllTopics[73].TopicA2M = ""
	AllTopics[74].TopicNumber = 74
	AllTopics[74].TopicName = "Z1_Cool_Curve_Outside_High_Temp"
	AllTopics[74].TopicBit = 88
	AllTopics[74].TopicFunction = "getIntMinus128"
	AllTopics[74].TopicUnit = "Celsius"
	AllTopics[74].TopicDisplayUnit = "°C"

	AllTopics[74].TopicA2M = ""
	AllTopics[75].TopicNumber = 75
	AllTopics[75].TopicName = "Z1_Cool_Curve_Outside_Low_Temp"
	AllTopics[75].TopicBit = 89
	AllTopics[75].TopicFunction = "getIntMinus128"
	AllTopics[75].TopicUnit = "Celsius"
	AllTopics[75].TopicDisplayUnit = "°C"

	AllTopics[75].TopicA2M = ""
	AllTopics[76].TopicNumber = 76
	AllTopics[76].TopicName = "Heating_Mode"
	AllTopics[76].TopicBit = 28
	AllTopics[76].TopicFunction = "getBit7and8"
	AllTopics[76].TopicUnit = "HeatCoolModeDesc"
	AllTopics[76].TopicA2M = ""
	AllTopics[77].TopicNumber = 77
	AllTopics[77].TopicName = "Heating_Off_Outdoor_Temp"
	AllTopics[77].TopicBit = 83
	AllTopics[77].TopicFunction = "getIntMinus128"
	AllTopics[77].TopicUnit = "Celsius"
	AllTopics[77].TopicDisplayUnit = "°C"

	AllTopics[77].TopicA2M = ""
	AllTopics[78].TopicNumber = 78
	AllTopics[78].TopicName = "Heater_On_Outdoor_Temp"
	AllTopics[78].TopicBit = 85
	AllTopics[78].TopicFunction = "getIntMinus128"
	AllTopics[78].TopicUnit = "Celsius"
	AllTopics[78].TopicA2M = ""
	AllTopics[78].TopicDisplayUnit = "°C"

	AllTopics[79].TopicNumber = 79
	AllTopics[79].TopicName = "Heat_To_Cool_Temp"
	AllTopics[79].TopicBit = 95
	AllTopics[79].TopicFunction = "getIntMinus128"
	AllTopics[79].TopicUnit = "Celsius"
	AllTopics[79].TopicDisplayUnit = "°C"

	AllTopics[79].TopicA2M = ""
	AllTopics[8].TopicNumber = 8
	AllTopics[8].TopicName = "Compressor_Freq"
	AllTopics[8].TopicBit = 166
	AllTopics[8].TopicFunction = "getIntMinus1"
	AllTopics[8].TopicUnit = "Hertz"
	AllTopics[8].TopicDisplayUnit = "hz"

	AllTopics[8].TopicA2M = ""
	AllTopics[80].TopicNumber = 80
	AllTopics[80].TopicName = "Cool_To_Heat_Temp"
	AllTopics[80].TopicBit = 96
	AllTopics[80].TopicDisplayUnit = "°C"

	AllTopics[80].TopicFunction = "getIntMinus128"
	AllTopics[80].TopicUnit = "Celsius"
	AllTopics[80].TopicA2M = ""
	AllTopics[81].TopicNumber = 81
	AllTopics[81].TopicName = "Cooling_Mode"
	AllTopics[81].TopicBit = 28
	AllTopics[81].TopicFunction = "getBit5and6"
	AllTopics[81].TopicUnit = "HeatCoolModeDesc"
	AllTopics[81].TopicA2M = ""
	AllTopics[82].TopicNumber = 82
	AllTopics[82].TopicName = "Z2_Heat_Curve_Target_High_Temp"
	AllTopics[82].TopicBit = 79
	AllTopics[82].TopicFunction = "getIntMinus128"
	AllTopics[82].TopicUnit = "Celsius"
	AllTopics[82].TopicA2M = ""
	AllTopics[82].TopicDisplayUnit = "°C"

	AllTopics[83].TopicNumber = 83
	AllTopics[83].TopicName = "Z2_Heat_Curve_Target_Low_Temp"
	AllTopics[83].TopicBit = 80
	AllTopics[83].TopicFunction = "getIntMinus128"
	AllTopics[83].TopicUnit = "Celsius"
	AllTopics[83].TopicDisplayUnit = "°C"

	AllTopics[83].TopicA2M = ""
	AllTopics[84].TopicNumber = 84
	AllTopics[84].TopicName = "Z2_Heat_Curve_Outside_High_Temp"
	AllTopics[84].TopicBit = 81
	AllTopics[84].TopicDisplayUnit = "°C"

	AllTopics[84].TopicFunction = "getIntMinus128"
	AllTopics[84].TopicUnit = "Celsius"
	AllTopics[84].TopicA2M = ""
	AllTopics[85].TopicNumber = 85
	AllTopics[85].TopicName = "Z2_Heat_Curve_Outside_Low_Temp"
	AllTopics[85].TopicBit = 82
	AllTopics[85].TopicDisplayUnit = "°C"

	AllTopics[85].TopicFunction = "getIntMinus128"
	AllTopics[85].TopicUnit = "Celsius"
	AllTopics[85].TopicA2M = ""
	AllTopics[86].TopicNumber = 86
	AllTopics[86].TopicName = "Z2_Cool_Curve_Target_High_Temp"
	AllTopics[86].TopicBit = 90
	AllTopics[86].TopicFunction = "getIntMinus128"
	AllTopics[86].TopicUnit = "Celsius"
	AllTopics[86].TopicDisplayUnit = "°C"

	AllTopics[86].TopicA2M = ""
	AllTopics[87].TopicNumber = 87
	AllTopics[87].TopicName = "Z2_Cool_Curve_Target_Low_Temp"
	AllTopics[87].TopicBit = 91
	AllTopics[87].TopicDisplayUnit = "°C"

	AllTopics[87].TopicFunction = "getIntMinus128"
	AllTopics[87].TopicUnit = "Celsius"
	AllTopics[87].TopicA2M = ""
	AllTopics[88].TopicNumber = 88
	AllTopics[88].TopicName = "Z2_Cool_Curve_Outside_High_Temp"
	AllTopics[88].TopicBit = 92
	AllTopics[88].TopicDisplayUnit = "°C"

	AllTopics[88].TopicFunction = "getIntMinus128"
	AllTopics[88].TopicUnit = "Celsius"
	AllTopics[88].TopicA2M = ""
	AllTopics[89].TopicNumber = 89
	AllTopics[89].TopicName = "Z2_Cool_Curve_Outside_Low_Temp"
	AllTopics[89].TopicBit = 93
	AllTopics[89].TopicFunction = "getIntMinus128"
	AllTopics[89].TopicUnit = "Celsius"
	AllTopics[89].TopicDisplayUnit = "°C"

	AllTopics[89].TopicA2M = ""
	AllTopics[9].TopicNumber = 9
	AllTopics[9].TopicName = "DHW_Target_Temp"
	AllTopics[9].TopicBit = 42
	AllTopics[9].TopicFunction = "getIntMinus128"
	AllTopics[9].TopicUnit = "Celsius"
	AllTopics[9].TopicDisplayUnit = "°C"

	AllTopics[9].TopicA2M = "DailyWaterTankSetpointTemperature"
	AllTopics[90].TopicNumber = 90
	AllTopics[90].TopicName = "Room_Heater_Operations_Hours"
	AllTopics[90].TopicBit = 0
	AllTopics[90].TopicDisplayUnit = "h"
	AllTopics[90].TopicFunction = "unknown"
	AllTopics[90].TopicUnit = "Hours"
	AllTopics[90].TopicA2M = ""
	AllTopics[91].TopicNumber = 91
	AllTopics[91].TopicName = "DHW_Heater_Operations_Hours"
	AllTopics[91].TopicBit = 0
	AllTopics[91].TopicDisplayUnit = "h"

	AllTopics[91].TopicFunction = "unknown"
	AllTopics[91].TopicUnit = "Hours"
	AllTopics[91].TopicA2M = ""

}
