package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/harmonicinc-com/joebot/client"
	joebot_middleware "github.com/harmonicinc-com/joebot/middleware"
	"github.com/harmonicinc-com/joebot/models"
	"github.com/harmonicinc-com/joebot/server"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("joebot", "Command & Control Server/Client For Managing Machines Via Web Interface")

	serverCommand = app.Command("server", "Server Mode")
	serverPort    = serverCommand.Flag("port", "Port For Listening Slave Machine, Default = 13579").Default("13579").Short('p').Int()
	webPortalPort = serverCommand.Flag("web-portal-port", "Port For The Web Portal, Default = 8080").Default("8080").Short('w').Int()

	clientCommand                = app.Command("client", "Client Mode")
	cServerIP                    = clientCommand.Arg("ip", "Server IP").Required().String()
	cServerPort                  = clientCommand.Flag("port", "Server Port, Default=13579").Default("13579").Short('p').Int()
	cAllowedPortRangeLBound      = clientCommand.Flag("allowed-port-lower-bound", "Lower Bound Of Allowed Port Range").Default("0").Short('l').Int()
	cAllowedPortRangeUBound      = clientCommand.Flag("allowed-port-upper-bound", "Upper Bound Of Allowed Port Range").Default("65535").Short('u').Int()
	cTags                        = clientCommand.Flag("tag", "Tags").Strings()
	cFilebrowserDefaultDirectory = clientCommand.Flag("dir", "Filebrowser Default Directory, Default=/").Default("/").Short('f').String()
)

func main() {
	defer func() {
		fmt.Println("Ended")
	}()

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case serverCommand.FullCommand():
		s := server.NewServer(nil)
		s.Start(*serverPort)

		e := echo.New()
		v1 := e.Group("/api")

		webPortalAssetsFS := WebPortalAssetsFS()

		v1.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		}))

		v1.Use(joebot_middleware.IPWhitelist(s.UserRepo))

		e.GET("/", func(c echo.Context) error {
			f, err := webPortalAssetsFS.Open("index.html")
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			b, err := io.ReadAll(f)
			if err != nil {
				log.Fatal(err)
			}
			return c.HTML(200, string(b))
		})
		// e.GET("/*", echo.WrapHandler(joebot_html.Handler))
		e.GET("/*", echo.WrapHandler(http.FileServer(http.FS(webPortalAssetsFS))))
		v1.GET("/clients", func(c echo.Context) error {
			return c.JSON(http.StatusOK, s.GetClientsList())
		})
		v1.GET("/users", func(c echo.Context) error {
			res, err := s.GetAllUser()
			if err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}

			return c.JSON(http.StatusOK, res)
		})
		v1.POST("/users", func(c echo.Context) error {
			json := models.UserInfo{}
			if err := c.Bind(&json); err != nil {
				return err
			}
			if err := s.CreateUser(json); err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}
			return c.JSON(http.StatusOK, json)
		})
		v1.POST("/login", func(c echo.Context) error {
			username, password := c.FormValue("username"), c.FormValue("password")
			res, err := s.UserLogin(username, password)
			if err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}
			return c.JSON(http.StatusOK, res)
		})
		v1.POST("/client/:id", func(c echo.Context) error {
			type msg struct {
				Message string `json:"message"`
			}

			client, err := s.GetClientById(c.Param("id"))
			if err != nil {
				return c.JSON(http.StatusNotFound, msg{err.Error()})
			}
			portStr := c.FormValue("target_client_port")
			port, err := strconv.Atoi(portStr)
			if err != nil || port <= 0 {
				return c.JSON(http.StatusBadRequest, msg{"Invalid target_client_port"})
			}

			portTunnelInfo, err := client.CreateTunnel(port)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, msg{err.Error()})
			}

			return c.JSON(http.StatusOK, portTunnelInfo)
		})
		v1.POST("/bulk-install", func(c echo.Context) error {
			json := models.BulkInstallInfo{}

			if err := c.Bind(&json); err != nil {
				return err
			}
			result, err := s.BulkInstallJoebot(json)
			if err != nil {
				return err
			}

			return c.String(http.StatusOK, result)
		})
		e.Start(":" + strconv.Itoa(*webPortalPort))
	case clientCommand.FullCommand():
		wg := &sync.WaitGroup{}
		wg.Add(1)
		c := client.NewClient(*cServerIP, *cServerPort, *cAllowedPortRangeLBound, *cAllowedPortRangeUBound, *cTags, nil)
		c.FilebrowserDefaultDir = *cFilebrowserDefaultDirectory
		c.Start()
		wg.Wait()
	}
}
