package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/gin-gonic/gin"
	blogtypes "github.com/username/blog/x/blog/types"

	"github.com/ignite-hq/cli/ignite/pkg/cosmosclient"
)

var aliceAddress = ""

type UserCreateResponse struct {
	Name     string "json:name"
	Address  string "json:address"
	Mnemonic string "json:mnemonic"
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
	router.POST("/addReward", validateUser(cosmos), addReward)

	router.Run("localhost:8080")

}

func validateUser(cosmos cosmosclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := c.Request.Header["Authorization"]

		if len(authToken) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}

		authTokenStr := string(authToken[0])
		address, err := cosmos.Address(authTokenStr)

		if err != nil || address.String() == "" {
			//c.Header("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
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
		c.JSON(http.StatusOK, gin.H{"message": "Failed To Get Balance"})
		return
	}
	b, err := getBalance(c, cosmos)
	c.JSON(http.StatusOK, gin.H{"Balance": b})
}

func addReward(c *gin.Context) {

	toAddr := c.GetString("userAddress")

	cmd := exec.Command("blogd", "keys", "show", "alice", "-a")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "addReward Failed: failed to construct fromAddr"})
		return
	}
	fromAddr := string(buf)
	fromAddr = strings.TrimSuffix(fromAddr, "\n")

	cmd = exec.Command("blogd", "tx", "bank", "send", fromAddr, toAddr, "10stake", "--chain-id", "blog", "-y")
	buf, err = cmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "addReward Failed: failed to construct send transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "addReward Success", "data": fromAddr + "00" + toAddr})
}

func userLogin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Login Sucess"})
}

func userRegister(c *gin.Context) {

	var newUser NewUser
	c.BindJSON(&newUser)
	cmd := exec.Command("blogd", "keys", "add", newUser.Name, "--output", "json")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Register Failed: failed to add key (use different name)"})
		return
	}
	op := string(buf)

	var x UserCreateResponse
	err = json.Unmarshal([]byte(op), &x)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Register Failed: failed to unmarshal add key output"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Register Success", "data": x})
}

func createPost(c *gin.Context) {

	cosmos, err := cosmosclient.New(context.Background())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "createPost Failed: failed to construct cosmosclient"})
		return
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
		c.JSON(http.StatusOK, gin.H{"message": "createPost Failed: failed to construct BroadcastTx"})
		return
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
