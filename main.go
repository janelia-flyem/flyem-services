// flyem-services API
//
// REST interface to access flyem services authentication server.  This
// service can create token for all registered applications.
//
//     Version: 0.1.0
//     Contact: Stephen Plaza<plazas@janelia.hhmi.org>
//
//     SecurityDefinitions:
//     Bearer:
//         type: apiKey
//         name: Authorization
//         in: header
//         scopes:
//           admin: Admin scope
//           user: User scope
//     Security:
//     - Bearer:
//
// swagger:meta
//go:generate swagger generate spec -o ./swaggerdocs/swagger.yaml
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	secure "github.com/janelia-flyem/echo-secure"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func customUsage() {
	fmt.Printf("Usage: %s [OPTIONS] CONFIG.json\n", os.Args[0])
	flag.PrintDefaults()
}

type ErrorInfo struct {
	Error string `json:"error"`
}

type AppAuth struct {
	config Config
}

func (aa AppAuth) listApps(c echo.Context) error {
	keys := make([]string, len(aa.config.ApplicationsSecrets))
	i := 0
	for k := range aa.config.ApplicationsSecrets {
		keys[i] = k
		i++
	}
	return c.JSON(http.StatusOK, keys)
}

type jwtCustomClaims struct {
	Email    string      `json:"email"`
	Level    interface{} `json:"level"`
	ImageUrl string      `json:"image-url"`
	jwt.StandardClaims
}

func (aa AppAuth) getAppToken(c echo.Context) error {
	email, ok := c.Get("email").(string)
	imageurl, ok2 := c.Get("imageurl").(string)
	if !ok || email == "" || !ok2 || imageurl == "" {
		errJSON := ErrorInfo{"cannot get authenticated email"}
		return c.JSON(http.StatusBadRequest, errJSON)
	}

	appname := c.Param("app")
	secret, ok := aa.config.ApplicationsSecrets[appname]
	if !ok {
		errJSON := ErrorInfo{"cannot find provided app"}
		return c.JSON(http.StatusBadRequest, errJSON)
	}

	// set level to noauth unless auth is specified
	level := interface{}("noauth")

	authdata := make(map[string]interface{})
	if authFile, ok := aa.config.ApplicationsAuth[appname]; ok {
		// open json file
		jsonFile, err := os.Open(authFile)
		if err != nil {
			errJSON := ErrorInfo{"authorization file cannnot be read for application"}
			return c.JSON(http.StatusBadRequest, errJSON)
		}
		byteData, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal(byteData, &authdata)
		if err != nil {
			errJSON := ErrorInfo{"error reading authorization file"}
			return c.JSON(http.StatusBadRequest, errJSON)
		}

		if userauth, ok := authdata[email]; ok {
			level = userauth
		}

	}

	claims := &jwtCustomClaims{
		email,
		level,
		imageurl,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 50000).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
	})
}

func main() {
	var port = 15000
	flag.Usage = customUsage
	flag.IntVar(&port, "port", 15000, "port to start server")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		return
	}

	// parse options
	options, err := LoadConfig(flag.Args()[0])
	if err != nil {
		fmt.Print(err)
		return
	}

	// create echo web framework
	e := echo.New()

	// setup logger
	logger, err := GetLogger(port, options)

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "{\"uri\": \"${uri}\", \"status\": ${status}, \"bytes_in\": ${bytes_in}, \"bytes_out\": ${bytes_out}, \"duration\": ${latency}, \"time\": ${time_unix}}\n",
		Output: logger,
	}))

	e.Use(middleware.Recover())
	e.Pre(middleware.NonWWWRedirect())

	var authorizer secure.Authorizer

	sconfig := secure.SecureConfig{
		SSLCert:          options.CertPEM,
		SSLKey:           options.KeyPEM,
		ClientID:         options.ClientID,
		ClientSecret:     options.ClientSecret,
		AuthorizeChecker: authorizer,
		Hostname:         options.Hostname,
	}
	secureAPI, err := secure.InitializeEchoSecure(e, sconfig, []byte(options.Secret), "flyem-services")
	if err != nil {
		fmt.Println(err)
		return
	}

	// create read only group
	readGrp := e.Group("/api")
	readGrp.Use(secureAPI.AuthMiddleware(secure.NOAUTH))

	// setup default page
	e.GET("/", secureAPI.AuthMiddleware(secure.NOAUTH)(func(c echo.Context) error {
		return c.HTML(http.StatusOK, "<html><title>flyem-services</title><body><a href='/api/help'>Documentation</a><form action='/logout' method='post'><input type='submit' value='Logout' /></form></body></html>")
	}))

	// swagger:operation GET /api/help apimeta helpyaml
	//
	// swagger REST documentation
	//
	// YAML file containing swagger API documentation
	//
	// ---
	// responses:
	//   200:
	//     description: "successful operation"
	if options.SwaggerFile != "" {
		e.File("/api/help", options.SwaggerFile)
	}

	aa := AppAuth{options}

	// swagger:operation GET /api/applications application listApps
	//
	// Gets applications supported
	//
	// List of applications that registered secrets with this service
	//
	// ---
	// responses:
	//   200:
	//     description: "successful operation"
	//     schema:
	//       type: "array"
	//       items:
	//         type: "string"
	//         description: "application names"
	// security:
	// - Bearer: []
	readGrp.GET("/applications", aa.listApps)

	// swagger:operation GET /token/{app} application getAppToken
	//
	// Returns JWT for given application
	//
	// If no authorization files are available, "noauth" is default for the user.
	//
	// ---
	// parameters:
	// - in: "path"
	//   name: "app"
	//   schema:
	//     type: "string"
	//   required: true
	//   description: "application name"
	// - in: "body"
	//   name: "body"
	//   required: true
	//   schema:
	//     type: "object"
	//     required: ["token"]
	//     properties:
	//       token:
	//         type: "string"
	//         description: "JWT token"
	// responses:
	//   200:
	//     description: "successful operation"
	// security:
	// - Bearer: []
	readGrp.GET("/token/:app", aa.getAppToken)

	// ?! add yaml documentation

	// start server
	secureAPI.StartEchoSecure(port)
}
