package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/stretchr/objx"
)

var (
	link, msgid, token, channelid, cookie string
	ids                                   = []string{}
)

func getBody() *http.Response {

	client := &http.Client{}
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {

	}
	req.Header.Set("cookie", "session="+cookie)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	return resp
}

func updateConfig(toWrite string) {
	f, err := os.Create("data.json")
	if err != nil {
		log.Fatal(err)
		f.Close()
	}
	f.WriteString(toWrite)
	f.Close()
}

type config struct {
	Token     string   `json:"token"`
	Msgid     string   `json:"msgid"`
	Link      string   `json:"link"`
	Channelid string   `json:"channelid"`
	Cookie    string   `json:"cookie"`
	Ids       []string `json:"ids"`
}

func loadConfig() {
	content, err := ioutil.ReadFile("./data.json")
	var c config
	err = json.Unmarshal(content, &c)
	if err != nil {
		fmt.Print(err)
		return
	}
	token = c.Token
	link = c.Link
	msgid = c.Msgid
	channelid = c.Channelid
	cookie = c.Cookie
	ids = c.Ids
}

func isInArray(arr []string, thing string) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] == thing {
			return true
		}
	}
	return false
}

func convert(str int) string {
	return fmt.Sprint(str)
}

func convertInt(str string) int {
	kek, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println(err)
	}
	return kek
}

func reverseInts(input []int) []int {
	var final []int
	for i := len(input) - 1; i > 0; i-- {
		final = append(final, input[i])
	}
	return final
}

func parseAllData(data *http.Response) []string {
	var end []string

	var dresp map[string]interface{}
	defer data.Body.Close()
	bodyBytes, err := ioutil.ReadAll(data.Body)
	bodyString := string(bodyBytes)
	err = json.Unmarshal([]byte(bodyString), &dresp)
	if err != nil {
		log.Fatal(err)
	}

	var ints []int
	for _, id := range ids {
		o := objx.New(dresp)
		thingy := o.Get("members." + id).ObjxMap()
		var name string
		var score int
		for key, value := range thingy {
			if key == "name" {
				name = fmt.Sprint(value)
			} else if key == "local_score" {
				score = convertInt(fmt.Sprint(value))

			}
		}
		ints = append(ints, score)

		kek := fmt.Sprintf("%s: %s", name, fmt.Sprint(score))
		end = append(end, kek)
	}
	sort.Ints(ints)
	ints = reverseInts(ints)
	var final []string
	var already []string
	for _, kekw := range ints {
		for _, info := range end {
			if strings.HasSuffix(info, ": "+convert(kekw)) && isInArray(already, strings.Split(info, ": ")[0]) == false {
				already = append(already, strings.Split(info, ": ")[0])
				final = append(final, info)
			}
		}
	}
	return final
}

func ready(session *discordgo.Session, event *discordgo.Ready) {
	_, err := session.ChannelMessage(channelid, msgid)
	if err != nil {
		msg, err := session.ChannelMessageSend(channelid, "Monkey!")
		if err != nil {
			fmt.Println(err)
			return
		}
		msgid = msg.ID
		jsonToWrite := fmt.Sprintf("{\"token\":\"%s\",\"msgid\": \"%s\",\"link\":\"%s\",\"channelid\":\"%s\",\"cookie\": \"%s\"}", token, msgid, link, channelid, cookie)
		updateConfig(jsonToWrite)
		for {
			data := getBody()
			thing := parseAllData(data)
			_, err := session.ChannelMessageEdit(channelid, msgid, "```"+strings.Join(thing, "\n")+"```")
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(15 * time.Minute)
		}
	} else {
		for {
			data := getBody()
			thing := parseAllData(data)
			_, err := session.ChannelMessageEdit(channelid, msgid, "```"+strings.Join(thing, "\n")+"```")
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(15 * time.Minute)
		}
	}
}

func main() {
	loadConfig()
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	bot.AddHandler(ready)
	err = bot.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	_ = bot.Close()
}
