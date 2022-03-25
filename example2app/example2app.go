package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

//	将客户端放在 test-network 目录下执行

//	contract 智能合约对象
var contract *gateway.Contract
//	ChannelName 通道名
const ChannelName = "mychannel1"
//	ChaincodeName 链码名
const ChaincodeName = "example2"

//	initContract 初始化合约对象
func initContract() {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")

	// 1：创建钱包
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("钱包创建失败: %s\n", err)
		os.Exit(1)
	}

	if !wallet.Exists("appUser") {
		// 钱包文件不存在，新建钱包
		err = populateWallet(wallet)
		if err != nil {
			fmt.Printf("populateWallet 钱包创建失败: %s\n", err)
			os.Exit(1)
		}
	}

	//	2：连接网络
	//	连接网络的配置文件 connection-org1.yaml，文件中有节点信息和与节点通信所需 tls 证书
	ccpPath := filepath.Join(
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		fmt.Printf("连接 fabric 网络失败: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()

	//	3：获取通道网络
	network, err := gw.GetNetwork(ChannelName)
	if err != nil {
		fmt.Printf("获取通道网络失败: %s\n", err)
		os.Exit(1)
	}

	//	4：获取合约对象，可以通过合约对象调用合约
	contract = network.GetContract(ChaincodeName)
}

func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	// 读取组织 user msp 证书
	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	// 读取组织 user msp 私钥
	keyDir := filepath.Join(credPath, "keystore")
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return errors.New("keystore 目录下没有找到私钥文件")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	// 生成 X509 公钥证书
	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	err = wallet.Put("appUser", identity)
	if err != nil {
		return err
	}
	return nil
}

//	add	增加员工
func add(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	form := request.Form
	result, err := contract.SubmitTransaction("Add", form.Get("no"), form.Get("name"), form.Get("age"), form.Get("salary"), form.Get("position"))
	if err != nil {
		r := fmt.Sprintf("交易失败: %s\n", err)
		fmt.Printf(r)
		writer.Write([]byte(r))
	}
	writer.Write(result)
}

//	findByNo 根据编号 No 查询员工
func findByNo(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	result, err := contract.EvaluateTransaction("FindByNo", request.Form.Get("no"))
	if err != nil {
		r := fmt.Sprintf("交易失败: %s\n", err)
		fmt.Printf(r)
		writer.Write([]byte(r))
	}
	writer.Write(result)
}

//	queryAll 查询所有员工
func queryAll(writer http.ResponseWriter, request *http.Request) {
	result, err := contract.EvaluateTransaction("QueryAll")
	if err != nil {
		r := fmt.Sprintf("交易失败: %s\n", err)
		fmt.Printf(r)
		writer.Write([]byte(r))
	}
	writer.Write(result)
}

//	salaryIncrease 为员工加薪
func salaryIncrease(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	_, err := contract.SubmitTransaction("SalaryIncrease", request.Form.Get("salary"))
	if err != nil {
		r := fmt.Sprintf("交易失败: %s\n", err)
		fmt.Printf(r)
		writer.Write([]byte(r))
	}
	writer.Write([]byte("交易成功"))
}


func main() {
	initContract()
	http.HandleFunc("/add", add)
	http.HandleFunc("/findByNo", findByNo)
	http.HandleFunc("/queryAll", queryAll)
	http.HandleFunc("/salaryIncrease", salaryIncrease)
	http.ListenAndServe("0.0.0.0:10810",nil)
}
