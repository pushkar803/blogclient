package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/gin-gonic/gin"
	blogtypes "github.com/username/blog/x/blog/types"

	"github.com/ignite-hq/cli/ignite/pkg/cosmosclient"
)

type UserCreateResponse struct {
	Name     string "json:creator"
	Address  string "json:title"
	Mnemonic string "json:body"
}

type BlogPost struct {
	Creator string "json:creator"
	Title   string "json:title"
	Body    string "json:body"
}

type NewUser struct {
	Name string "json:name"
}

func main() {

	router := gin.Default()

	cosmos, err := cosmosclient.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	//router.Use(validateUser(cosmos))
	router.POST("/login", userLogin)
	router.POST("/register", userRegister)

	router.GET("/posts", validateUser(cosmos), getAllPosts)
	router.POST("/post", validateUser(cosmos), createPost)
	router.GET("/myBalance", validateUser(cosmos), fetchBalance)
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

func fetchBalance(c *gin.Context) {
	cosmos, err := cosmosclient.New(context.Background())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Message": "Failed To Get Balance"})
		return
	}
	b, err := getBalance(c, cosmos)
	c.JSON(http.StatusOK, gin.H{"Balance": b})
}

func userLogin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Message": "Login Sucess"})
}

func userRegister(c *gin.Context) {

	var newUser NewUser
	c.BindJSON(&newUser)
	cmd := exec.Command("blogd", "keys", "add", newUser.Name, "--output", "json")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Message": "Register Failed"})
		return
	}
	op := string(buf)
	fmt.Println(op)

	c.JSON(http.StatusOK, gin.H{"Message": "Register Failed", "op": op})
}

func createPost(c *gin.Context) {

	cosmos, err := cosmosclient.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	accountName := c.GetString("authToken")

	var iPost BlogPost
	c.BindJSON(&iPost)

	msg := &blogtypes.MsgCreatePost{
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

	queryClient := blogtypes.NewQueryClient(cosmos.Context)
	queryResp, err := queryClient.Posts(context.Background(), &blogtypes.QueryPostsRequest{})
	if err != nil {
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusOK, queryResp)

}

func getBalance(c *gin.Context, cosmos cosmosclient.Client) (string, error) {
	resp, err := banktypes.NewQueryClient(cosmos.Context).Balance(c, &banktypes.QueryBalanceRequest{
		Address: c.GetString("userAddress"),
		Denom:   "stake",
	})
	return resp.Balance.String(), err
}
