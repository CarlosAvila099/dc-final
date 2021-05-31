package resources

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"bufio"
	"strings"
	"io/ioutil"
	"io"
	"github.com/dgrijalva/jwt-go"
	"encoding/json"
	
	"go.nanomsg.org/mangos"
	"go.nanomsg.org/mangos/protocol/pair"
	"go.nanomsg.org/mangos/protocol/pub"
)

const(
	GRAYSCALE = 0
	BLUR = 1

	SCHEDULING = 0
	RUNNING = 1
	COMPLETED = 2
	
	CONTROLLER = "tcp://localhost:40899"
	API = "tcp://localhost:40900"
	SCHEDULER = "tcp://localhost:50051"

	LISTEN = 0
	DIAL = 1
)

type WorkloadJSON struct {
	Filter string `form:"filter" json:"filter" xml:"filter" binding:"-"`
	WorkloadName string `form:"workload_name" json:"workload_name" xml:"workload_name" binding:"-"`
}

type ImageJSON struct {
	WorkloadId int `form:"workload_id" json:"workload_id" xml:"workload_id" binding:"-"`
	ImagePath string `form:"image" json:"image" xml:"image" binding:"-"`
	ImageType string `form:"type" json:"type" xml:"type" binding:"-"`
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

type Job struct {
	Address string
	RPCName string //Filter
	ImagePath string
	WorkloadPath string
	CurrentId int
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
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil{
		return err
	}
	s.Token = signedToken
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

func (w *Workload) SaveWorkload() bool{
	path := "./images/" + w.Name
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
		os.Mkdir(path + "/images", os.ModePerm)
		w.CreateFile()
		return true
	}
	return false
}

func (w *Workload) CreateFile(){
	information := strconv.Itoa(w.Id) + "&" + strconv.Itoa(w.Filter) + "&" + w.Name
    ioutil.WriteFile("./images/" + w.Name + "/info.txt", []byte(information), 0644)
}

func (w *Workload) GetFilter() string{
	var filter string
	if w.Filter == GRAYSCALE{
		filter = "Grayscale"
	} else if w.Filter == BLUR{
		filter = "Blur"
	}
	return filter
}

func (w *Workload) CopyImage(image string, counter int, works []Workload) (int, bool) {
	splitter := strings.Split(image, ".")
	ext := splitter[len(splitter) - 1]
	id := counter + 1
	source, err := os.Open(image)
    if err != nil {
            return -1, false
    }
    defer source.Close()
	newPath := "images/" + w.Name + "/images/o" + strconv.Itoa(id) + "." + ext
    destination, err := os.Create(newPath)
    if err != nil {
		fmt.Println(err.Error())
        return 0, false
    }
    defer destination.Close()
    _, err = io.Copy(destination, source)
	if err != nil{
		return -2, false
	}
	works[w.Id].FilteredImages = w.FilteredImages
    return id, true
}

func ReadWorkloads() ([]Workload, int){
	var works []Workload
	var dirs []string 
	var wl Workload
	counter := 0
	path := "./images"
	files, _ := ioutil.ReadDir(path)
    for _, f := range files {
        if f.IsDir(){
			dirs = append(dirs, f.Name())
		}
    }
	for _, name := range dirs{
		filepath := path + "/" + name + "/info.txt"
		file, _ := os.Open(filepath)
		scanner := bufio.NewScanner(file)
    	for scanner.Scan() {
        	splitted := strings.Split(scanner.Text(), "&")
			id, _ := strconv.Atoi(splitted[0])
			filter, _ := strconv.Atoi(splitted[1])
			workName := splitted[2]
			wl = Workload { id, filter, workName, SCHEDULING, 0, make([]int, 0) }
    	}
		images := path + "/" + name + "/images"
		files, _ := ioutil.ReadDir(images)
    	for _, f := range files {
			name := f.Name()[0:1]
			if name == "f"{
				splitted := strings.Split(f.Name()[1:], ".")
				id, _ := strconv.Atoi(splitted[0])
				wl.FilteredImages = append(wl.FilteredImages, id)
			}
			counter++
    	}
		works = append(works, wl)
		file.Close()
	}
	return works, counter
}

func SearchImage(wm []Workload, s string) (string, string, bool){
	for _, wk := range wm{
		images :=  "images/" + wk.Name + "/images"
		files, _ := ioutil.ReadDir(images)
    	for _, f := range files {
			id := strings.Split(f.Name(), ".")[0]
			if id == s{
				return images + "/", f.Name(), true
			}
    	}	
	}
	return "", "", false
}

func DownloadImage(p string, n string) bool {
	path := "./downloads"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
	source, err := os.Open(p + n)
    if err != nil {
		fmt.Println("S", err)
        return false
    }
    defer source.Close()
	newPath := path + "/" + n
    destination, err := os.Create(newPath)
    if err != nil {
		fmt.Println("D", err)
        return false
    }
    defer destination.Close()
    _, err = io.Copy(destination, source)
	if err != nil{
		fmt.Println("C", err)
		return false
	}
    return true
}

func Exists(wm []Workload, s string) bool{
	for _, wk := range wm{
		if wk.Name == s{
			return true
		}
	}
	return false
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

func GetPublisher() mangos.Socket{
	var sock mangos.Socket
	var err error
	if sock, err = pub.NewSocket(); err != nil {
		Die("can't get new pub socket: %s", err)
	}
	if err = sock.Listen(CONTROLLER); err != nil {
		Die("can't listen on pub socket: %s", err.Error())
	}
	return sock
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