/*
 * Copyright (C) 2024 Nethesis S.r.l.
 * http://www.nethesis.it - info@nethesis.it
 *
 * SPDX-License-Identifier: GPL-2.0-only
 *
 * author: Edoardo Spadoni <edoardo.spadoni@nethesis.it>
 */

package main

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/NethServer/nethsecurity-api/response"
	"github.com/NethServer/nethsecurity-controller/api/cache"
	"github.com/NethServer/nethsecurity-controller/api/configuration"
	"github.com/NethServer/nethsecurity-controller/api/logs"
	"github.com/NethServer/nethsecurity-controller/api/methods"
	"github.com/NethServer/nethsecurity-controller/api/middleware"
	"github.com/NethServer/nethsecurity-controller/api/socket"
	"github.com/NethServer/nethsecurity-controller/api/storage"
)

// @title NethSecurity Controller API Server
// @version 1.0
// @description NethSecurity Controller API Server is used to manage multiple stand-alone NethSecurity instances
// @termsOfService https://nethserver.org/terms/

// @contact.name NethServer Developer Team
// @contact.url https://nethserver.org/support

// @license.name GNU GENERAL PUBLIC LICENSE

// @host localhost:5000
// @schemes http
// @BasePath /api

func refreshCacheLoop() {
	ticker := time.NewTicker(60 * time.Minute)
	for range ticker.C {
		// load all units info into cache
		units, err := methods.ListUnits()
		if err != nil {
			return
		}

		for _, unit := range units {
			unitInfo, err := methods.GetRemoteInfo(unit)
			if err == nil {
				cache.SetUnitInfo(unit, unitInfo)
			}
		}
	}
}

func main() {
	// init logs with syslog
	logs.Init("nethsecurity_controller")

	// init configuration
	configuration.Init()

	// init storage
	storage.Init()

	// init socket connection
	socket.Init()

	// init cache
	cache.Init()

	go refreshCacheLoop() // starts cache refresh loop

	// disable log to stdout when running in release mode
	if gin.Mode() == gin.ReleaseMode {
		gin.DefaultWriter = ioutil.Discard
	}

	// init routers
	router := gin.Default()

	// add default compression
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// cors configuration only in debug mode GIN_MODE=debug (default)
	if gin.Mode() == gin.DebugMode {
		// gin gonic cors conf
		corsConf := cors.DefaultConfig()
		corsConf.AllowHeaders = []string{"Authorization", "Content-Type", "Accept"}
		corsConf.AllowAllOrigins = true
		router.Use(cors.New(corsConf))
	}

	// define api group
	api := router.Group("/")

	// define login and logout endpoint
	api.POST("/login", middleware.InstanceJWT().LoginHandler)
	api.POST("/logout", middleware.InstanceJWT().LogoutHandler)

	// define server registration
	api.POST("/units/register", methods.RegisterUnit)

	// define JWT middleware
	api.Use(middleware.InstanceJWT().MiddlewareFunc())
	{
		// refresh handler
		api.GET("/refresh", middleware.InstanceJWT().RefreshHandler)

		// accounts APIs
		accounts := api.Group("/accounts")
		{
			// accounts CRUD
			accounts.GET("", methods.GetAccounts)
			accounts.GET("/:account_id", methods.GetAccount)
			accounts.POST("", methods.AddAccount)
			accounts.PUT("/:account_id", methods.UpdateAccount)
			accounts.DELETE("/:account_id", methods.DeleteAccount)

			// account password change
			accounts.PUT("/password", methods.UpdatePassword)

			// ssh keys read and write
			accounts.GET("/ssh-keys", methods.GetSSHKeys)
			accounts.POST("/ssh-keys", methods.AddSSHKeys)
			accounts.DELETE("/ssh-keys", methods.DeleteSSHKeys)
		}

		// default APIs
		defaults := api.Group("/defaults")
		{
			defaults.GET("", methods.GetDefaults)
		}

		// units APIs
		units := api.Group("/units")
		{
			units.GET("", methods.GetUnits)
			units.GET("/:unit_id", methods.GetUnit)
			units.GET("/:unit_id/token", methods.GetToken)
			units.POST("", methods.AddUnit)
			units.DELETE("/:unit_id", methods.DeleteUnit)
		}
	}

	// handle missing endpoint
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, structs.Map(response.StatusNotFound{
			Code:    404,
			Message: "API not found",
			Data:    nil,
		}))
	})

	// run server
	router.Run(configuration.Config.ListenAddress)
}
