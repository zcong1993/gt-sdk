package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zcong1993/gt-sdk"
	"net/http"
)

var gtClient = gt.NewCt(*gt.DefaultConfig)

func main() {
	r := gin.Default()

	r.Static("/static", "_example/static")

	r.GET("/gt/register-click", func(c *gin.Context) {
		resp, err := gtClient.Register("", "")
		if err != nil {
			fmt.Printf("%+v\n", err)
			c.Status(http.StatusInternalServerError)
		}

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

		ok, err := gtClient.Validate(&f, false)
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
