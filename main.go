package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ignite-hq/cli/ignite/pkg/cosmosclient"
	"github.com/username/blog/x/blog/types"
)

type BlogPost struct {
	Creator string "json:creator"
	Title   string "json:title"
	Body    string "json:body"
}

func main() {

	router := gin.Default()

	cosmos, err := cosmosclient.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	router.Use(validateUser(cosmos))

	router.GET("/posts", getAllPosts)
	router.POST("/post", createPost)

	router.Run("localhost:8080")

}

func validateUser(cosmos cosmosclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := c.Request.Header["Authorization"]

		if len(authToken) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"Message": "Unauthorized"})
			c.Abort()
			return
		}

		authTokenStr := string(authToken[0])
		address, err := cosmos.Address(authTokenStr)

		if err != nil || address.String() == "" {
			//c.Header("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
			c.JSON(http.StatusUnauthorized, gin.H{"Message": "Unauthorized"})
			c.Abort()
			return
		}

		c.Set("authToken", authTokenStr)
		c.Set("userAddress", address.String())
		c.Next()
	}
}

func createPost(c *gin.Context) {

	cosmos, err := cosmosclient.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	accountName := c.GetString("authToken")

	var iPost BlogPost
	c.BindJSON(&iPost)

	msg := &types.MsgCreatePost{
		Creator: c.GetString("userAddress"),
		Title:   iPost.Title,
		Body:    iPost.Body,
	}

	txResp, err := cosmos.BroadcastTx(accountName, msg)
	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, txResp)

}

func getAllPosts(c *gin.Context) {

	cosmos, err := cosmosclient.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	queryClient := types.NewQueryClient(cosmos.Context)
	queryResp, err := queryClient.Posts(context.Background(), &types.QueryPostsRequest{})
	if err != nil {
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusOK, queryResp)

}
