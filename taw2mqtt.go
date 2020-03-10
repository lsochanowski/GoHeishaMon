package main

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.bug.st/serial"
)

var panasonicQuery []byte = []byte{0x71, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
var PANASONICQUERYSIZE int = 110

//should be the same number
var NUMBER_OF_TOPICS int = 92
var AllTopics [92]TopicData
var MqttKeepalive time.Duration

var actData [92]string
var config Config
var sending bool
var Serial serial.Port
var err error
var goodreads float64
var totalreads float64
var readpercentage float64

type command_struct struct {
	value  [128]byte
	length int
}

type TopicData struct {
	TopicNumber   int
	TopicName     string
	TopicBit      int
	TopicFunction string
	TopicUnit     string
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
	Aquarea2mqttPumpID     string
	MqttPass               string
	MqttClientID           string
	MqttKeepalive          int
}

func ReadConfig() Config {
	var configfile = "config"
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

func main() {
	config = ReadConfig()
	if config.Readonly != true {
		log_message("Not sending this command. Heishamon in listen only mode! - this POC version don't support writing yet....")
		os.Exit(0)
	}
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
	ParseTopicList()
	MqttKeepalive = time.Second * time.Duration(config.MqttKeepalive)
	MC, MT := MakeMQTTConn()

	for {
		send_command(panasonicQuery, PANASONICQUERYSIZE)
		readSerial(MC, MT)
		time.Sleep(PoolInterval)

	}

}

func MakeMQTTConn() (mqtt.Client, mqtt.Token) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s:%s", "tcp", config.MqttServer, config.MqttPort))
	opts.SetPassword(config.MqttPass)
	opts.SetUsername(config.MqttLogin)
	opts.SetClientID(config.MqttClientID)

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
	c.Subscribe("aquarea/+/+/set", 2, HandleMSGfromMQTT)

	//Perform additional action...
}

func HandleMSGfromMQTT(mclient mqtt.Client, msg mqtt.Message) {

}

func log_message(a string) {
	fmt.Println(a)
}

func logHex(command []byte, length int) {
	fmt.Println(command)
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

func ParseTopicList() {
	lines, err := ReadCsv("Topics.csv")
	if err != nil {
		panic(err)
	}

	// Loop through lines & turn into object
	for _, line := range lines {
		TB, _ := strconv.Atoi(line[2])
		TNUM, _ := strconv.Atoi(line[0])
		data := TopicData{
			TopicNumber:   TNUM,
			TopicName:     line[1],
			TopicBit:      TB,
			TopicFunction: line[3],
			TopicUnit:     line[4],
		}
		AllTopics[TNUM] = data
		//a	fmt.Println(data)
	}

}

func ReadCsv(filename string) ([][]string, error) {

	// Open CSV file
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return lines, nil
}

func send_command(command []byte, length int) bool {

	if sending {
		log_message("Already sending data. Buffering this send request")
		//	pushCommandBuffer(command, length)
		return false
	}
	sending = true //simple semaphore to only allow one send command at a time, semaphore ends when answered data is received
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

	totalreads++
	data := make([]byte, 203)
	n, err := Serial.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	if n == 0 {
		fmt.Println("\nEOF")

	}

	data_length := 203
	//panasonic read is always 203 on valid receive, if not yet there wait for next read
	log_message("Received 203 bytes data")
	sending = false //we received an answer after our last command so from now on we can start a new send request again
	if config.Loghex {
		logHex(data, data_length)
	}
	if !isValidReceiveHeader(data) {
		log_message("Received wrong header!")
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
	return fmt.Sprintf("%f", out)

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
	return fmt.Sprintf("%f", PumpFlow)
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
				fmt.Println("NIE MA FUNKCJI", v.TopicFunction, "\n")
			}

		}

		if (updatenow) || (actData[k] != Topic_Value) {
			actData[k] = Topic_Value
			fmt.Printf("received TOP%d %s: %s \n", k, v.TopicName, Topic_Value)
			value = strings.TrimSpace(Topic_Value)
			value = strings.ToUpper(Topic_Value)
			if config.Aquarea2mqttCompatible {
				TOP := "aquarea/state/" + fmt.Sprintf("%s/%s", config.Aquarea2mqttPumpID, v.TopicName)
				fmt.Println("Publikuje do ", TOP, "warosc", value)
				token = mclient.Publish(TOP, byte(0), false, value)
				if token.Wait() && token.Error() != nil {
					fmt.Printf("Fail to publish, %v", token.Error())
				}
			}
			TOP := fmt.Sprintf("%s/%s", config.Mqtt_topic_base, v.TopicName)
			fmt.Println("Publikuje do ", TOP, "warosc", value)
			token = mclient.Publish(TOP, byte(0), false, value)
			if token.Wait() && token.Error() != nil {
				fmt.Printf("Fail to publish, %v", token.Error())
			}

		}

	}

	// 	for (unsigned int Topic_Number = 0 ; Topic_Number < NUMBER_OF_TOPICS ; Topic_Number++) {
	// 	  byte Input_Byte;
	// 	  String Topic_Value;
	// 	  switch (Topic_Number) { //switch on topic numbers, some have special needs
	// 		case 1:
	// 		  Topic_Value = getPumpFlow(data);
	// 		  break;
	// 		case 11:
	// 		  Topic_Value = String(word(data[183], data[182]) - 1);
	// 		  break;
	// 		case 12:
	// 		  Topic_Value = String(word(data[180], data[179]) - 1);
	// 		  break;
	// 		case 90:
	// 		  Topic_Value = String(word(data[186], data[185]) - 1);
	// 		  break;
	// 		case 91:
	// 		  Topic_Value = String(word(data[189], data[188]) - 1);
	// 		  break;
	// 		case 44:
	// 		  Topic_Value = getErrorInfo(data);
	// 		  break;
	// 		default:
	// 		  Input_Byte = data[topicBytes[Topic_Number]];
	// 		  Topic_Value = topicFunctions[Topic_Number](Input_Byte);
	// 		  break;
	// 	  }
	// 	  if ((updatenow) || ( actData[Topic_Number] != Topic_Value )) {
	// 		actData[Topic_Number] = Topic_Value;
	// 		sprintf(log_msg, "received TOP%d %s: %s", Topic_Number, topics[Topic_Number], Topic_Value.c_str()); log_message(log_msg);
	// 		sprintf(mqtt_topic, "%s/%s", mqtt_topic_base, topics[Topic_Number]); mqtt_client.publish(mqtt_topic, Topic_Value.c_str(), MQTT_RETAIN_VALUES);
	// 	  }
	// 	}

}
