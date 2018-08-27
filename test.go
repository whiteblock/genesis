package main

import (
	"fmt"
)

func runTests(){
	fmt.Println("++++++++++++++IP Scheme Tests++++++++++++++")
	testIP()
	fmt.Println("++++++++++++++Ethereum Tests++++++++++++++")
	testEthereum()
}

func iPTestPair(server int, node int){
	ip := getNodeIP(server,node)
	gw := getGateway(server,node)
	fmt.Printf("Server %d, Node %d:\nIP:%s\nGateway:%s\n\n",server,node,ip,gw)
}

func testIP(){
	powTest1 := pow(2,8)
	powTest2 := pow(10,3)
	powTest3 := pow(2,4)
	fmt.Println("--------------Power--------------")
	fmt.Printf("%d == 256\n",powTest1)
	fmt.Printf("%d == 1000\n",powTest2)
	fmt.Printf("%d == 16\n",powTest3)
	
	fmt.Println("---------------Node IP AND Gateway IP----------------")

	iPTestPair(1,0)
	iPTestPair(1,1)
	iPTestPair(1,2)
	iPTestPair(1,3)
	iPTestPair(1,5)
	iPTestPair(3,41)
	iPTestPair(2000,5)


	fmt.Println("---------------Get Gateways----------------")

	gateways1 := getGateways(1, 200)
	gateways2 := getGateways(10, 20)
	gateways3 := getGateways(100, 2)
	fmt.Printf("Gateways for Server 1		with 200 nodes :\n %+v \n",gateways1)
	fmt.Printf("Gateways for Server 10		with 20 nodes :\n %+v \n",gateways2)
	fmt.Printf("Gateways for Server 100		with 2 nodes :\n %+v \n",gateways3)

}


func testEthereum(){
	fmt.Println("---------------Init Node----------------")
	fmt.Printf("%s\n",eth_initNode(2, 1755,"10.2.0.4"))
	fmt.Printf("%s\n",eth_initNode(1, 1755,"10.2.0.3"))
	fmt.Println("---------------Get Wallet----------------")
	fmt.Printf("%s\n",eth_getWallet(2))
	fmt.Printf("%s\n",eth_getWallet(1))
}