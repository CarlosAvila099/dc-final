package api

import(
	"net/http"
	"time"
	"strconv"
	"github.com/dgrijalva/jwt-go"
	"log"
)

/*
 * Represents a session, with information needed to operate
**/
type Session struct{
	User string
	Token string
}

/*
 * Handles the errors given by different functions of the code
 * @param c refers to the error code defined
 * @param e refers to the error given, can be nil
 * @param w refers to the writer used to connect with the client
**/
func errorHandler(c int, e error, w http.ResponseWriter){
	switch c{
		case 1:
			message := `
{
	"message": "There was an error while getting the token"
	"error": "` + e.Error() + `"
}
`
			w.Write([]byte(message))
			break
		case 2:
			message := `
{
	"message": "There was a problem while revoking the token, please try again"
}
`
			w.Write([]byte(message))
			break
		case 3:
			message := `
{
	"message": "Please enter a username and password"
}
`
			w.Write([]byte(message))
			break;
		case 4:
			message := `
{
	"message": "Invalid username or password"
}
`
			w.Write([]byte(message))
			break
		case 5:	
			message := `
{
	"message": "Please enter a token"
}
`
			w.Write([]byte(message))
			break
		case 6:	
			message := `
{
	"message": "Invalid token"
}
`
			w.Write([]byte(message))
			break
		case 7:
			message := `
{
	"message": "There was an error uploading the image"
	"error": "` + e.Error() +  `"
}
`
			w.Write([]byte(message))
			break
		case 8:
			message := `
{
	"message": "The image surpasses the limit of 10mb"
}
`
			w.Write([]byte(message))
			break
	}
}

/*
 * Checks if username and password exist
 * @param u refers to the username
 * @param p refers to the password
 * @return bool that checks if the username and password are correct
**/
func isAuthorised(u string, p string) bool{
	authorization := map[string]string{
		"username": "password",
		"root": "",
	}
	if pass, ok := authorization[u]; ok{
		return p == pass
	}
	return false
}

/*
 * Creates a token
 * @param u refers to the username
 * @return string that refers to the token and an error if necesary, nil otherwise
 * @see https://learn.vonage.com/blog/2020/03/13/using-jwt-for-authentication-in-a-golang-application-dr/
**/
func getToken(u string) (string, error){
	secretKey := "MOOTCKTPOXOOTCK"
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = u
	atClaims["exp"] = time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil{
		return "", err
	}
	return signedToken, nil
}

/*
 * Checks if session is already started
 * @param t refers to the token
 * @return Session related to the token given and bool to check if session is started
**/
func inSession(t string) (Session, bool){
	for _, s := range sessionManager{
		if t == s.Token{
			return s, true
		}
	}
	return Session{"", ""}, false
}

/* Removes an item from an array
 * @param s refers to array that will have an item removed
 * @param i refers to the position of the item to be removed
 * @return array with the item removed
**/
func remove(s []Session, i int) []Session {
    s[len(s)-1], s[i] = s[i], s[len(s)-1]
    return s[:len(s)-1]
}

/*
 * Revokes token, ends active session 
 * @param t refers to token to be removed
 * @return bool to check if session has ended
**/
func revokeToken(t string) bool{
	for n, s := range sessionManager{
		if t == s.Token{
			sessionManager = remove(sessionManager, n)
			break
		}
	}
	for _, s := range sessionManager{
		if t == s.Token{
			return false
		}
	}
	return true
}

/*
 * Converts the bytes received to kb, mg depending on the occassion
 * @param size refers to the bytes received
 * @return string that has the size in its converted form, bool that refers if the size is in the limit defined
**/
func getSize(size int64) (string, bool){
	var KB, MB, LIMIT float64 = 1024, 1048576, 10485760
	fSize := float64(size)
	if fSize < KB{
		return strconv.FormatFloat(fSize, 'f', 2, 64) + "b", true
	} else if fSize >= KB && fSize < MB{
		return strconv.FormatFloat(fSize/KB, 'f', 2, 64) + "kb", true
	} else if fSize >= MB && fSize <= LIMIT{
		return strconv.FormatFloat(fSize/MB, 'f', 2, 64) + "mb", true
	} else{
		return "", false
	}
}

/*
 * Starts session
 * @param w refers to the writer connected to the client
 * @param r refers to the requests made
 * @see https://blog.umesh.wtf/how-to-implement-http-basic-auth-in-gogolang
**/
func login(w http.ResponseWriter, r *http.Request){
	w.Header().Add("Content-type", "application/json")
	username, password, ok := r.BasicAuth()
    if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		errorHandler(3, nil, w)
        return
    }
	
	if !isAuthorised(username, password) {
		w.WriteHeader(http.StatusUnauthorized)
		errorHandler(4, nil, w)
		return
	}
	
    w.WriteHeader(http.StatusOK)
	token, err := getToken(username)
	if err != nil{
		errorHandler(1, err, w)
		return
	}
	message := `
{
	"message": "Hi ` + username + ` welcome to the DPIP System"
	"token" "` + token + `"
}
`
    w.Write([]byte(message))
	sessionManager = append(sessionManager, Session{username, token})
	log.Println("User '" + username + "' has started a session.")
    return
}

/*
 * Ends session
 * @param w refers to the writer connected to the client
 * @param r refers to the requests made
 * @see https://golang.org/pkg/net/http/#Header.Get
**/
func logout(w http.ResponseWriter, r *http.Request){
	token := r.Header.Get("Authorization")
	if len(token) < 7{
		w.WriteHeader(http.StatusUnauthorized)
		errorHandler(5, nil, w)
        return
	}
	token = token[7:]
	session, started := inSession(token)
	if !started{
        w.WriteHeader(http.StatusUnauthorized)
		errorHandler(6, nil, w)
        return
	}
	revoked := revokeToken(token)
	if !revoked{
		errorHandler(2, nil, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	message := `
{
	"message": "Bye ` + session.User + `, your token has been revoked"
}
`
	w.Write([]byte(message))
	log.Println("User '" + session.User + "' has ended the session.")
	return
}

/*
 * Uploads an image to the server
 * @param w refers to the writer connected to the client
 * @param r refers to the requests made
 * @see https://golang.org/pkg/net/http/#Request.FormFile
 * @see https://golang.org/pkg/mime/multipart/#FileHeader
**/
func upload(w http.ResponseWriter, r *http.Request){
	token := r.Header.Get("Authorization")
	if len(token) < 7{
		w.WriteHeader(http.StatusUnauthorized)
		errorHandler(5, nil, w)
        return
	}
	token = token[7:]
	session, started := inSession(token)
	if !started{
        w.WriteHeader(http.StatusUnauthorized)
		errorHandler(6, nil, w)
        return
	}
	w.WriteHeader(http.StatusOK)
	r.ParseMultipartForm(10 << 20)
	_, header, err  := r.FormFile("data")
	if err != nil{
		w.WriteHeader(http.StatusUnauthorized)
		errorHandler(7, err, w)
		return
	}
	size, ok := getSize(header.Size)
	if !ok{
		errorHandler(8, nil, w)
		return
	}
	message := `
{
	"message": "An image has been successfully uploaded"
	"filename": "` + header.Filename + `"
	"size": "` + size + `"
}
`
    w.Write([]byte(message))
	log.Println("User '" + session.User + "' uploaded '" + header.Filename + "'.")
	return 
}

/*
 * Gives session status
 * @param w refers to the writer connected to the client
 * @param r refers to the requests made
 * @see https://golang.org/pkg/net/http/#Header.Get
**/
func status(w http.ResponseWriter, r *http.Request){
	token := r.Header.Get("Authorization")
	if len(token) < 7{
		w.WriteHeader(http.StatusUnauthorized)
		errorHandler(5, nil, w)
        return
	}
	token = token[7:]
	session, started := inSession(token)
	if !started{
        w.WriteHeader(http.StatusUnauthorized)
		errorHandler(6, nil, w)
        return
	}
	w.WriteHeader(http.StatusOK)
	t := time.Now()
	message := `
{
	"message": "Hi ` + session.User + `, the DPIP System is Up and Running "
	"time": "` + t.Format("2006-01-02 15:04:05") + `"
}
`
    w.Write([]byte(message))
	log.Println("User '" + session.User + "' wants to know the status.")
    return
}

var sessionManager []Session // Manages all started sessions

func Start() {
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/status", status)
	http.ListenAndServe("localhost:8080", nil)
}
