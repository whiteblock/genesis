package main

import (
	"./util"
	"fmt"
)

func runTests() {
	fmt.Println("++++++++++++++IP Scheme Tests++++++++++++++")
	testIP()
	/*fmt.Println("++++++++++++++Ethereum Tests++++++++++++++")
	testEthereum()*/
}

func iPTestPair(server int, node int) {
	ip := util.GetNodeIP(server, node)
	gw := util.GetGateway(server, node)
	fmt.Printf("Server %d, Node %d:\nIP:%s\nGateway:%s\n\n", server, node, ip, gw)
}

func testIP() {

	fmt.Println("---------------Node IP AND Gateway IP----------------")

	iPTestPair(1, 0)
	iPTestPair(1, 1)
	iPTestPair(1, 2)
	iPTestPair(1, 3)
	iPTestPair(1, 5)
	iPTestPair(3, 41)
	iPTestPair(2000, 5)

	fmt.Println("---------------Get Gateways----------------")

	gateways1 := util.GetGateways(1, 200)
	gateways2 := util.GetGateways(10, 20)
	gateways3 := util.GetGateways(100, 2)
	fmt.Printf("Gateways for Server 1		with 200 nodes :\n %+v \n", gateways1)
	fmt.Printf("Gateways for Server 10		with 20 nodes :\n %+v \n", gateways2)
	fmt.Printf("Gateways for Server 100		with 2 nodes :\n %+v \n", gateways3)

}
