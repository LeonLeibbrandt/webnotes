package server

import (
	"fmt"
	"net/http"
	"time"

	"web/global"
)

const SESSIONKEY = "session"

type Server struct {
	*http.Server
	global *global.Global
}

func NewServer(g *global.Global) (*Server, error) {

	getIP := func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Getting IP")
			r.Header.Set("IPAddress", getIPAddress(r))
			f(w, r)
		}
	}

	auth := func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Authing")
			// First check session cookie
			cookie, err := r.Cookie(SESSIONKEY)
			if err != nil {
				// Check basic auth
				username, password, ok := r.BasicAuth()
				if !ok {
					// If not logged in send login page
					w.Write([]byte(LOGINFORM))
					return
				}
				token, known := g.DB.Auth(username, password, r.Header.Get("IPAddress"))
				if !known {
					http.Error(w, "Unknown User or Password Incorrect", http.StatusUnauthorized)
					return
				}
				http.SetCookie(w, &http.Cookie{
					Name:    SESSIONKEY,
					Value:   token,
					Path:    "/",
					Expires: time.Now().Add(time.Hour * 24),
				})
				f(w, r)
				return
			}
			fmt.Printf("Cookie %#+v\n", cookie)
			known := g.DB.Session(cookie.Value, r.Header.Get("IPAddress"))
			if !known {
				http.Error(w, "Unknown User or Password Incorrect", http.StatusUnauthorized)
				return
			}
			f(w, r)
		}
	}

	s := &Server{
		global: g,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/pages", getIP(auth(s.Pages)))
	mux.HandleFunc("/", getIP(auth(s.Pages)))

	httpServer := &http.Server{
		Addr:    g.Config.ServerPort,
		Handler: mux,
	}

	s.Server = httpServer

	return s, nil
}

func (s *Server) Pages(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Ip Address for Login : %s\n", r.Header.Get("IPAddress"))
	w.Write([]byte("OK"))
}

var LOGINFORM = `<!DOCTYPE html>
<html>
	<header>
		<style type="text/css">
		 /* Bordered form */
		 form {
			 border: 3px solid #f1f1f1;
		 }
		 
		 /* Full-width inputs */
		 input[type=text], input[type=password] {
			 width: 100%;
			 padding: 12px 20px;
			 margin: 8px 0;
			 display: inline-block;
			 border: 1px solid #ccc;
			 box-sizing: border-box;
		 }
		 
		 /* Set a style for all buttons */
		 button {
			 background-color: #4CAF50;
			 color: white;
			 padding: 14px 20px;
			 margin: 8px 0;
			 border: none;
			 cursor: pointer;
			 width: 100%;
		 }
		 
		 /* Add a hover effect for buttons */
		 button:hover {
			 opacity: 0.8;
		 }
		 
		 /* Extra style for the cancel button (red) */
		 .cancelbtn {
			 width: auto;
			 padding: 10px 18px;
			 background-color: #f44336;
		 }
		 
		 /* Center the avatar image inside this container */
		 .imgcontainer {
			 text-align: center;
			 margin: 24px 0 12px 0;
		 }
		 
		 /* Avatar image */
		 img.avatar {
			 width: 40%;
			 border-radius: 50%;
		 }
		 
		 /* Add padding to containers */
		 .container {
			 padding: 16px;
		 }
		 
		 /* The "Forgot password" text */
		 span.psw {
			 float: right;
			 padding-top: 16px;
		 }
		 
		 /* Change styles for span and cancel button on extra small screens */
		 @media screen and (max-width: 300px) {
			 span.psw {
				 display: block;
				 float: none;
			 }
			 .cancelbtn {
				 width: 100%;
			 }
		 }
		</style>
		<script type="text/javascript">
		 function onSubmit() {
          
			 console.log("Here");
             return false;
		 }
		</script>
		<title>Login</title>
	</header>
	<body>
		<form method="get" onsubmit="return onSubmit()">
			<div class="imgcontainer">
				<img src="img_avatar2.png" alt="Avatar" class="avatar">
			</div>
			
			<div class="container">
				<label for="uname"><b>Username</b></label>
				<input type="text" placeholder="Enter Username" name="uname" required>
				
				<label for="psw"><b>Password</b></label>
				<input type="password" placeholder="Enter Password" name="psw" required>
				
				<button type="submit">Login</button>
				<label>
					<input type="checkbox" checked="checked" name="remember"> Remember me
				</label>
			</div>
			
			<div class="container" style="background-color:#f1f1f1">
				<button type="button" class="cancelbtn">Cancel</button>
				<span class="psw">Forgot <a href="#">password?</a></span>
			</div>
		</form>
	</body>
</html>
`
