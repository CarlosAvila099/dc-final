package api

import(
	"net/http"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
	"encoding/json"
	"go.nanomsg.org/mangos"

	"github.com/CarlosAvila099/dc-final/resources"
)

/*
 * Handles the errors given by different functions of the code
 * @param c refers to the error code defined
 * @param e refers to the error given, can be nil
 * @param w refers to the writer used to connect with the client
**/
func errorHandler(c float64, e error, w http.ResponseWriter){
	var msg []byte
	var str string
	var err error
	switch c{
		/*
		 * Error Code Format:
		 *	Integer Codes:
		 *		1 = Login Errors
		 *		2 = Token Errors
		 *		3 = Image Errors
		 *		4 = Workload Errors
		 *		5 = JSON Errors
		**/
		case 1.1:
			str = "Please enter username and password"
			break;
		case 1.2:
			str = "Invalid username or password"
			break
		case 2.1:
			str = "There was an error while getting the token"
			break
		case 2.2:
			str = "Please enter a token"
			break
		case 2.3:	
			str = "Invalid token"
			break
		case 2.4:
			str = "There was a problem while revoking the token, please try again"
			break
		case 3.1:
			str = "Please enter a png, jpg or jpeg file"
			break
		case 3.2:
			str = "There was a problem while getting the image"
			break
		case 3.3:
			str = "The image id given does not exist"
			break
		case 3.4:
			str = "The type given does is not supported, please choose either original or filtered"
			break
		case 3.5:
			str = "There was an error while downloading the image"
			break
		case 4.1:
			str = "The filter given is not supported, please choose either grayscale or blur"
			break
		case 4.2:
			str = "There was an error while getting the workload id"
			break
		case 4.3:
			str = "The workload id given does not exist"
			break
		case 4.4:
			str = "Please enter a workload name"
			break
		case 4.5:
			str = "The workload already exists"
			break
		case 5.1:
			str = "There was an error while encoding the message"
			break
		case 5.2:
			str = "There was a problem while decoding the data given"
	}
	if msg, err = resources.ErrorMessage(str, e); err != nil{
		errorHandler(-1, err, w)
		return
	}
	w.Write(msg)
}

func login(c * gin.Context){
	var ok bool
	var err error
	var msg []byte
	c.Writer.Header().Set("Content-type", "application/json")
	username, password, ok := c.Request.BasicAuth()
    if !ok {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(1.1, nil, c.Writer)
        return
    }
	session := resources.Session{ username, password, "" }
	ok, err = session.StartSession()
	if !ok{
		c.Writer.WriteHeader(http.StatusUnauthorized)
		if err != nil{
			errorHandler(2.1, err, c.Writer)
			return
		}
		errorHandler(1.2, nil, c.Writer)
		return
	}
    c.Writer.WriteHeader(http.StatusOK)
	message := resources.Message{ "Hi " + username + " welcome to the DPIP System", map[string]string {"Token": session.Token} }
	if msg, err = message.MakeMessage(); err != nil{
		errorHandler(5.1, err, c.Writer)
		return
	}
    c.Writer.Write(msg)
	sessionManager = append(sessionManager, session)
    return
}

func logout(c *gin.Context){
	var err error
	var msg []byte
	token := c.Request.Header.Get("Authorization")
	if len(token) < 7{
		c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.2, nil, c.Writer)
        return
	}
	token = token[7:]
	session, started := resources.InSession(token, sessionManager)
	if !started{
        c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.3, nil, c.Writer)
        return
	}
	revoked := resources.RevokeToken(token, sessionManager)
	if !revoked{
		errorHandler(2.4, nil, c.Writer)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
	message := resources.Message{"Bye " + session.User + ", your token has been revoked", map[string]string{}}
	if msg, err = message.MakeMessage(); err != nil{
		errorHandler(5.1, err, c.Writer)
		return
	}
    c.Writer.Write(msg)
	return
}

func images(c *gin.Context){
	var err error
	var msg []byte
	c.Writer.Header().Set("Content-type", "application/json")
	token := c.Request.Header.Get("Authorization")
	if len(token) < 7{
		c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.2, nil, c.Writer)
        return
	}
	token = token[7:]
	_, started := resources.InSession(token, sessionManager)
	if !started{
        c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.3, nil, c.Writer)
        return
	}
	c.Writer.WriteHeader(http.StatusOK)
	var data resources.ImageJSON
	err = c.ShouldBindJSON(&data)
	if err != nil{
		c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(5.2, err, c.Writer)
		return
	}
	resources.SendToPair(socket, "4&" + strconv.Itoa(data.WorkloadId) + "&" + data.ImagePath + "&" + data.ImageType)
	bytes := resources.ReceiveFromPair(socket)
	response := new(resources.ControllerResponse)
	_ = json.Unmarshal([]byte(bytes), &response)
	if response.Response != ""{
		if response.Response == "-4.1"{
			errorHandler(4.3, nil, c.Writer)
			return
		} else if response.Response == "-4.2"{
			errorHandler(3.4, nil, c.Writer)
			return
		} else if response.Response == "-4.3"{
			errorHandler(3.1, nil, c.Writer)
			return
		} else if response.Response == "-4.4"{
			errorHandler(3.3, nil, c.Writer)
			return
		}
	}
	message := resources.Message{ "The image has been uploaded", map[string]string{"Image ID":response.Response, "Workload ID":strconv.Itoa(response.Work.Id), "Type":data.ImageType} }
	if msg, err = message.MakeMessage(); err != nil{
		errorHandler(5.1, err, c.Writer)
		return
	}
    c.Writer.Write(msg)
	return
}

func ImagesGet(c *gin.Context){
	var err error
	var msg []byte
	c.Writer.Header().Set("Content-type", "application/json")
	token := c.Request.Header.Get("Authorization")
	if len(token) < 7{
		c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.2, nil, c.Writer)
        return
	}
	token = token[7:]
	_, started := resources.InSession(token, sessionManager)
	if !started{
        c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.3, nil, c.Writer)
        return
	}
	c.Writer.WriteHeader(http.StatusOK)
	id, err := strconv.Atoi(c.Param("id")[1:])
	if err != nil{
		errorHandler(3.2, err, c.Writer)
		return
	}
	resources.SendToPair(socket, "5&" + strconv.Itoa(id))
	bytes := resources.ReceiveFromPair(socket)
	response := new(resources.ControllerResponse)
	_ = json.Unmarshal([]byte(bytes), &response)
	if response.Response != ""{
		if response.Response == "-5.1"{
			errorHandler(3.3, nil, c.Writer)
			return
		} else if response.Response == "-5.2"{
			errorHandler(3.5, nil, c.Writer)
			return
		}
	}
	message := resources.Message{ "Image downloaded succesfully, please check for the image in downloads/" + response.Response, map[string]string{ } }
	if msg, err = message.MakeMessage(); err != nil{
		errorHandler(5.1, err, c.Writer)
		return
	}
    c.Writer.Write(msg)
	return
	return
}

func workloads(c *gin.Context){
	var err error
	var msg []byte
	c.Writer.Header().Set("Content-type", "application/json")
	token := c.Request.Header.Get("Authorization")
	if len(token) < 7{
		c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.2, nil, c.Writer)
        return
	}
	token = token[7:]
	_, started := resources.InSession(token, sessionManager)
	if !started{
        c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.3, nil, c.Writer)
        return
	}
	c.Writer.WriteHeader(http.StatusOK)
	var data resources.WorkloadJSON
	err = c.ShouldBindJSON(&data)
	if err != nil{
		c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(5.2, err, c.Writer)
		return
	}
	name := data.WorkloadName
	filter := data.Filter
	resources.SendToPair(socket, "1&" + name + "&" + filter)
	bytes := resources.ReceiveFromPair(socket)
	response := new(resources.ControllerResponse)
	_ = json.Unmarshal([]byte(bytes), &response)
	if response.Response != ""{
		if response.Response == "-1"{
			errorHandler(4.1, nil, c.Writer)
			return
		} else if response.Response == "-1.1"{
			errorHandler(4.4, nil, c.Writer)
			return
		} else if response.Response == "-1.2"{
			errorHandler(4.5, nil, c.Writer)
			return
		}
	}
	work := response.Work
	message := resources.Message{ "The workload has been successfully created", map[string]string{ "Workload ID":strconv.Itoa(work.Id), "Filter":work.GetFilter(), "Workload Name":work.Name,  "Status":work.GetStatus(), "Running Jobs":strconv.Itoa(work.RunningJobs), "Filtered Images": work.GetImages() } }
	if msg, err = message.MakeMessage(); err != nil{
		errorHandler(5.1, err, c.Writer)
		return
	}
    c.Writer.Write(msg)
	return 
}

func workloadsGet(c *gin.Context){
	var err error
	var msg []byte
	c.Writer.Header().Set("Content-type", "application/json")
	token := c.Request.Header.Get("Authorization")
	if len(token) < 7{
		c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.2, nil, c.Writer)
        return
	}
	token = token[7:]
	_, started := resources.InSession(token, sessionManager)
	if !started{
        c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.3, nil, c.Writer)
        return
	}
	c.Writer.WriteHeader(http.StatusOK)
	id, err := strconv.Atoi(c.Param("id")[1:])
	if err != nil{
		errorHandler(4.2, err, c.Writer)
		return
	}
	resources.SendToPair(socket, "2&" + strconv.Itoa(id))
	bytes := resources.ReceiveFromPair(socket)
	response := new(resources.ControllerResponse)
	_ = json.Unmarshal([]byte(bytes), &response)
	if response.Response != ""{
		if response.Response == "-2"{
			errorHandler(4.3, nil, c.Writer)
			return
		}
	}
	work := response.Work
	message := resources.Message{ "Information retrieved successfully", map[string]string{ "Workload ID":strconv.Itoa(work.Id), "Filter":work.GetFilter(), "Workload Name":work.Name,  "Status":work.GetStatus(), "Running Jobs":strconv.Itoa(work.RunningJobs), "Filtered Images": work.GetImages() } }
	if msg, err = message.MakeMessage(); err != nil{
		errorHandler(5.1, err, c.Writer)
		return
	}
    c.Writer.Write(msg)
	return
}

func status(c *gin.Context){
	var err error
	var msg []byte
	token := c.Request.Header.Get("Authorization")
	if len(token) < 7{
		c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.2, nil, c.Writer)
        return
	}
	token = token[7:]
	session, started := resources.InSession(token, sessionManager)
	if !started{
        c.Writer.WriteHeader(http.StatusUnauthorized)
		errorHandler(2.3, nil, c.Writer)
        return
	}
	c.Writer.WriteHeader(http.StatusOK)
	t := time.Now()
	resources.SendToPair(socket, "3&")
	bytes := resources.ReceiveFromPair(socket)
	response := new(resources.ControllerResponse)
	_ = json.Unmarshal([]byte(bytes), &response)
	message := resources.Message{ "Hi " + session.User + ", the DPIP System is Up and Running", map[string]string{ "Time": t.Format("2006-01-02 15:04:05"), "Active Workloads":response.Response } }
	if msg, err = message.MakeMessage(); err != nil{
		errorHandler(5.1, err, c.Writer)
		return
	}
	c.Writer.Write(msg)
    return
}

var sessionManager []resources.Session // Manages all started sessions
var socket mangos.Socket

func Start() {
	gin.SetMode(gin.ReleaseMode)
	url := "localhost:8080"
	socket = resources.GetSocket(false)
	r := gin.Default()
	r.POST("/login", login)
	r.DELETE("/logout", logout)
	r.GET("/status", status)
	r.POST("/images", images)
	r.GET("/images/*id", ImagesGet)
	r.POST("/workloads", workloads)
	r.GET("/workloads/*id", workloadsGet)
	r.Run(url)
}