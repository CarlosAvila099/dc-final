package resources

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"github.com/dgrijalva/jwt-go"
	"encoding/json"
	
	"go.nanomsg.org/mangos"
	"go.nanomsg.org/mangos/protocol/pair"
)

const(
	NORMAL = 0
	GRAYSCALE = 1
	BLUR = 2
	SCHEDULING = 0
	RUNNING = 1
	COMPLETED = 2
	
	CONTROLLER = "tcp://localhost:40899"
	API = "tcp://localhost:40900"

	LISTEN = 0
	DIAL = 1
)

type WorkloadJSON struct {
	Filter string `form:"filter" json:"filter" xml:"filter" binding:"-"`
	WorkloadName string `form:"workload_name" json:"workload_name" xml:"workload_name" binding:"-"`
}

type ImageJSON struct {
	WorkloadId int `form:"workload_id" json:"workload_id" xml:"workload_id" binding:"-"`
}

type Workload struct {
	Id int
	Filter int
	Name string
	Status int
	RunningJobs int
	FilteredImages []int
}

type ControllerResponse struct{
	Work Workload
	Response string
}

type Session struct{
	User string
	Pass string
	Token string
}

type Message struct{
	Message string
	Info map[string]string
}

func (s *Session) StartSession() (bool, error){
	var err error
	if !s.isAuthorised(){
		return false, nil
	}
	if err = s.getToken(); err != nil{
		return false, err
	}
	return true, nil
}

func (s *Session) isAuthorised() bool{
	authorization := map[string]string{
		"username": "password",
		"root": "",
	}
	if pass, ok := authorization[s.User]; ok{
		return pass == s.Pass
	}
	return false
}

func (s *Session) getToken() (error){
	secretKey := "MOOTCKTPOXOOTCK"
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = s.User
	atClaims["exp"] = time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	_, err := token.SignedString([]byte(secretKey))
	if err != nil{
		return err
	}
	s.Token = "yeet" //signedToken
	return nil
}

func RevokeToken(t string, sm []Session) bool{
	for n, s := range sm{
		if t == s.Token{
			sm = remove(sm, n)
			break
		}
	}
	_, ok := InSession(t, sm)
	return !ok
}

func InSession(t string, sm []Session) (Session, bool){
	for _, s := range sm{
		if t == s.Token{
			return s, true
		}
	}
	return Session{"", "", ""}, false
}

func remove(s []Session, i int) []Session {
    s[len(s)-1], s[i] = s[i], s[len(s)-1]
    return s[:len(s)-1]
}

func (w *Workload) GetImages() string{
	files := ""
	for _,value := range w.FilteredImages{
		files = files + strconv.Itoa(value) + ", "
	}
	if len(w.FilteredImages) > 0{
		files = files[:len(files) - 2]
	}
	return files
}

func (w *Workload) GetStatus() string{
	var status string
	if w.Status == SCHEDULING{
		status = "Scheduling"
	} else if w.Status == RUNNING{
		status = "Running"
	} else if w.Status == COMPLETED{
		status = "Completed"
	}
	return status
}

func (w *Workload) GetFilter() string{
	var filter string
	if w.Filter == NORMAL{
		filter = "Original"
	} else if w.Filter == GRAYSCALE{
		filter = "Grayscale"
	} else if w.Filter == BLUR{
		filter = "Blur"
	}
	return filter
}

func Die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func GetSocket(first bool) mangos.Socket{
	var socket mangos.Socket
	var err error
	if socket, err = pair.NewSocket(); err != nil{
		Die("Can't Create a pair socket:", err.Error())
	}
	if first{
		if err = socket.Listen(API); err != nil{
			Die("Can't Listen to API:", err.Error())
		}
	} else{
		if err = socket.Dial(API); err != nil{
			Die("Can't Dial to API:", err.Error())
		}
	}
	return socket
}

func ReceiveFromPair(socket mangos.Socket) string{
	var bytes []byte
	var err error
	if bytes, err = socket.Recv(); err != nil {
		Die("There was an error receiving the information:", err.Error())
	}
	message := string(bytes)
	return message
}

func SendToPair(socket mangos.Socket, message string){
	var err error
	if err = socket.Send([]byte(message)); err != nil {
		Die("There was an error sending the information:", err.Error())
	}
}

func ErrorMessage(s string, e error) ([]byte, error){
	var message map[string]string
	if e != nil{
		message = map[string]string { "Message":s, "Error":e.Error() }
	} else{
		message = map[string]string { "Message":s }
	}
	jsonValue, err := json.MarshalIndent(message, " ", "    ")
	if err != nil{
		return make([]byte, 0), err
	}
	return jsonValue, nil
}

func (m *Message) MakeMessage() ([]byte, error){
	message := map[string]string { "Message":m.Message }
	for key, element := range m.Info{
		message[key] = element
	}
	jsonValue, err := json.MarshalIndent(message, " ", "    ")
	if err != nil{
		return make([]byte, 0), err
	}
	return jsonValue, nil
}