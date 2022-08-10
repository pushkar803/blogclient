package main

import (
	"context"
	"encoding/json"
	"fmt"
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

type MintRequest struct {
	TokenURI string "json:tokenURI"
}

type BurnRequest struct {
	NftId string "json:nftId"
}

type TransferRequest struct {
	NftId     string "json:nftId"
	ToAddress string "json:toAddress"
}

type DistributeRequest struct {
	NftIds      string "json:nftIds"
	ToAddresses string "json:toAddresses"
}

type NftItem struct {
	Owner    string "json:owner"
	Id       string "json:id"
	TokenURI string "json:tokenURI"
}

type JsonOP struct {
	NftItem []NftItem "json:NftItem"
}

type SingleJsonOP struct {
	NftItem NftItem "json:NftItem"
}
type NftIDLists struct {
	NftIDList  []string `json:"nftIdList"`
	Pagination struct {
		NextKey interface{} `json:"next_key"`
		Total   string      `json:"total"`
	} `json:"pagination"`
}

type UserTotalNFTs struct {
	Total string `json:"total"`
}

type TxJsonOutput struct {
	Height    string `json:"height"`
	Txhash    string `json:"txhash"`
	Codespace string `json:"codespace"`
	Code      int    `json:"code"`
	Data      string `json:"data"`
	RawLog    string `json:"raw_log"`
	Logs      []struct {
		MsgIndex int    `json:"msg_index"`
		Log      string `json:"log"`
		Events   []struct {
			Type       string `json:"type"`
			Attributes []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"attributes"`
		} `json:"events"`
	} `json:"logs"`
	Info      string      `json:"info"`
	GasWanted string      `json:"gas_wanted"`
	GasUsed   string      `json:"gas_used"`
	Tx        interface{} `json:"tx"`
	Timestamp string      `json:"timestamp"`
	Events    []struct {
		Type       string `json:"type"`
		Attributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
			Index bool   `json:"index"`
		} `json:"attributes"`
	} `json:"events"`
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

	//NFT routes
	router.POST("/mint", validateUser(cosmos), mint)
	router.POST("/burn", validateUser(cosmos), burn)
	router.POST("/transfer", validateUser(cosmos), transfer)
	router.POST("/distribute", validateUser(cosmos), distribute)

	router.GET("/list_nft_item", validateUser(cosmos), listAllNfts)
	router.GET("/show_nft_item/:nftId", validateUser(cosmos), showNftById)
	router.GET("/list_nft_id_of_owner/:ownerAddress", validateUser(cosmos), listNftsOfOwner)
	router.GET("/nft_count_of/:ownerAddress", validateUser(cosmos), GetCountOfNfts)

	router.Run("localhost:8089")

}

func listAllNfts(c *gin.Context) {

	cmd := exec.Command("blogd", "q", "sagenft", "list-nft-item", "--output", "json")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "listAllNfts Failed"})
		return
	}
	op := string(buf)

	var x JsonOP
	err = json.Unmarshal([]byte(op), &x)
	fmt.Println(err)
	if err != nil {
		c.JSON(http.StatusOK, "failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": x})
}

func showNftById(c *gin.Context) {

	nftId := c.Param("nftId")
	cmd := exec.Command("blogd", "q", "sagenft", "show-nft-item", nftId, "--output", "json")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "showNftById Failed"})
		return
	}
	op := string(buf)
	var x SingleJsonOP
	err = json.Unmarshal([]byte(op), &x)
	if err != nil {
		c.JSON(http.StatusOK, "failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": x})
}

func listNftsOfOwner(c *gin.Context) {

	ownerAddress := c.Param("ownerAddress")
	cmd := exec.Command("blogd", "q", "sagenft", "list-nft-id-of-owner", ownerAddress, "--output", "json")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "listNftsOfOwner Failed"})
		return
	}
	op := string(buf)
	var x NftIDLists
	err = json.Unmarshal([]byte(op), &x)
	if err != nil {
		c.JSON(http.StatusOK, "failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": x})
}

func GetCountOfNfts(c *gin.Context) {

	ownerAddress := c.Param("ownerAddress")
	cmd := exec.Command("blogd", "q", "sagenft", "nft-count-of", ownerAddress, "--output", "json")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "GetCountOfNfts Failed"})
		return
	}
	op := string(buf)
	var x UserTotalNFTs
	err = json.Unmarshal([]byte(op), &x)
	if err != nil {
		c.JSON(http.StatusOK, "failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": x})
}

func mint(c *gin.Context) {

	var mintRequest MintRequest
	c.BindJSON(&mintRequest)

	cmd := exec.Command("blogd", "tx", "sagenft", "mint", mintRequest.TokenURI, "--output", "json", "--from", c.GetString("authToken"), "-y")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Mint Failed"})
		return
	}
	op := string(buf)
	var x TxJsonOutput
	err = json.Unmarshal([]byte(op), &x)
	if err != nil {
		c.JSON(http.StatusOK, "failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": x})
}

func burn(c *gin.Context) {

	var burnRequest BurnRequest
	c.BindJSON(&burnRequest)
	cmd := exec.Command("blogd", "tx", "sagenft", "burn", burnRequest.NftId, "--output", "json", "--from", c.GetString("authToken"), "-y")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Burn Failed"})
		return
	}
	op := string(buf)
	var x TxJsonOutput
	err = json.Unmarshal([]byte(op), &x)
	if err != nil {
		c.JSON(http.StatusOK, "failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": x})
}

func transfer(c *gin.Context) {

	var transferRequest TransferRequest
	c.BindJSON(&transferRequest)
	cmd := exec.Command("blogd", "tx", "sagenft", "transfer", transferRequest.NftId, transferRequest.ToAddress, "--output", "json", "--from", c.GetString("authToken"), "-y")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Transfer Failed"})
		return
	}
	op := string(buf)
	var x TxJsonOutput
	err = json.Unmarshal([]byte(op), &x)
	if err != nil {
		c.JSON(http.StatusOK, "failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": x})
}

func distribute(c *gin.Context) {

	var distributeRequest DistributeRequest
	c.BindJSON(&distributeRequest)
	cmd := exec.Command("blogd", "tx", "sagenft", "distribute", distributeRequest.NftIds, distributeRequest.ToAddresses, "--output", "json", "--from", c.GetString("authToken"), "-y")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Distribute Failed"})
		return
	}
	op := string(buf)
	var x TxJsonOutput
	err = json.Unmarshal([]byte(op), &x)
	if err != nil {
		c.JSON(http.StatusOK, "failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": x})
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
