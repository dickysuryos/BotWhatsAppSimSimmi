package main
// ./go run ../goWhatsApp/examples/receiveMessages/receive.go
import (
	"encoding/gob"
	"fmt"
	"github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp/binary/proto"
	"github.com/Rhymen/go-whatsapp"
	"os"
	"strings"
	"time"
	"os/signal"
	"syscall"
	"net/http"
	"encoding/json"
	"log"
	"io/ioutil"

)
type Validate struct {
	doge string
	btc string
}
type goIndoDax struct {
	ticker Ticker
}

type Ticker struct{
	last string
}

type waHandler struct {
	c *whatsapp.Conn
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (wh *waHandler) HandleError(err error) {
	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(5 * time.Second)
		log.Println("Reconnecting...")
		err := wh.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		log.Printf("error occoured: %v\n", err)
	}
}

//Optional to be implemented. Implement HandleXXXMessage for the types you need.
func (wh *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	// fmt.Printf("%v %v %v %v\n\t%v\n", message.Info.Timestamp, message.Info.Id, message.Info.RemoteJid, message.ContextInfo.QuotedMessageID, message.Text)


if strings.Index(strings.ToLower(message.Text),"!") == 0  {
// 	return
// }
// if validateMessage(message.Text){
var t = strings.ToLower(message.Text)
t = strings.Trim(t, "!")
url := "https://indodax.com/api/"+t+"_idr/ticker"
// payload := http.NewRequest(
fmt.Printf("ini url '%v'",t)
req, reqerr := http.NewRequest("POST", url,strings.NewReader(""))
if reqerr != nil {
    return 
}
req.Header.Add("Content-Type", "application/json")
res, reserr := http.DefaultClient.Do(req)
if reserr != nil {
    return 
}
defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
 	fmt.Printf("ini res '%v'",res)
	fmt.Printf("ini body '%v'",string(body))
	if strings.Contains(string(body),"error"){
		return
	}
 	fmt.Printf("ini body tanpa string '%v'",body)
 	var result map[string]interface{}
 	if body == nil {
		return
	}
	jsonErr := json.Unmarshal([]byte(body), &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		return
	}

	
	if result == nil {
		return
	}

	response := result["ticker"].(map[string]interface{})
	if response == nil {
        return
    }
    var last = ""
    var high = ""
    var low  = ""
	if response != nil {
	for key,value:= range response {
  	// Each value is an interface{} type, that is type asserted as a string
  	if strings.Contains(strings.ToLower(key),"last"){
			last = value.(string)
			}
	if strings.Contains(strings.ToLower(key),"high"){
			high = value.(string)
			}
	if strings.Contains(strings.ToLower(key),"low"){
			low = value.(string)
			}
  		}
  		sendMessage(message,last,high,low,wh)
  	}
  }
	
}

func sendMessage(message whatsapp.TextMessage,valuela string,valuh string,valuelo string,wh *waHandler){
	var t = message.Text
	t = strings.Trim(t, "!")
	previousMessage := message.Text
	quotedMessage := proto.Message{
		Conversation: &previousMessage,
	}

ContextInfo := whatsapp.ContextInfo{
		QuotedMessage:   &quotedMessage,
		QuotedMessageID: message.Text,
		Participant:message.Info.RemoteJid, //Whot sent the original message
	}


	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: message.Info.RemoteJid,

		},
		ContextInfo: ContextInfo,
		Text: t+" last:"+valuela+",high:"+valuh+",low:"+valuelo+" #Bot",

				}
		if _, err := wh.c.Send(msg); err != nil {
		fmt.Fprintf(os.Stderr, "error sending message: %v\n", err)
				}
		}

func validateMessage(value string) bool{
	if strings.Contains(strings.ToLower(value),"harga doge") || strings.Contains(strings.ToLower(value),"harga btc"){
		return true
	}
	return false
}

func main() {
	//create new WhatsApp connection
	wac, err := whatsapp.NewConn(50 * time.Second)
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}

	//Add handler
	wac.AddHandler(&waHandler{wac})

	//login or restore
	if err := login(wac); err != nil {
		log.Fatalf("error logging in: %v\n", err)
	}

	//verifies phone connectivity
	pong, err := wac.AdminTest()

	if !pong || err != nil {
		log.Fatalf("error pinging in: %v\n", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	//Disconnect safe
	fmt.Println("Shutting down now.")
	session, err := wac.Disconnect()
	if err != nil {
		log.Fatalf("error disconnecting: %v\n", err)
	}
	if err := writeSession(session); err != nil {
		log.Fatalf("error saving session: %v", err)
	}
}

func login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v\n", err)
		}
	}

	//save session
	err = writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}
	return nil
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	file, err := os.Create(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}
