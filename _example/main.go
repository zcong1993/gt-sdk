package main

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/zcong1993/gt-sdk"
	"net/http"
)


func main() {
	config := gt.DefaultConfig
	//config.ApiServer = "test.com"
	gtClient := gt.NewCt(*config)

	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("gtsession", store))

	r.Static("/static", "_example/static")

	r.GET("/gt/register-click", func(c *gin.Context) {
		session := sessions.Default(c)
		resp, err := gtClient.Register("", "")
		if err != nil {
			session.Set("fallback", true)
			fmt.Printf("%+v, use fallback\n", err)
		} else {
			session.Set("fallback", false)
		}

		session.Save()
		c.JSON(http.StatusOK, resp)
	})

	r.POST("/gt/validate-click", func(c *gin.Context) {
		var f gt.ValidateForm
		err := c.ShouldBind(&f)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		fmt.Printf("%+v\n", f)

		session := sessions.Default(c)
		var fallback bool
		v := session.Get("fallback")
		if v == nil {
			fallback = false
		} else {
			fallback = v.(bool)
		}

		ok, err := gtClient.Validate(&f, fallback)
		resp := map[string]interface{}{}

		if err != nil {
			resp["status"] = "error"
			resp["info"] = err

			fmt.Printf("%+v\n", err)
		} else {
			if ok {
				resp["status"] = "success"
				resp["info"] = "success"
			} else {
				resp["status"] = "failed"
				resp["info"] = "failed"
			}
		}

		c.JSON(http.StatusOK, resp)
	})

	r.Run()
}
