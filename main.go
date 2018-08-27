package main

import "flag"


/**
 * Arguments
 * -n Number of Nodes (Docker Containers) in Cluster 
 * -i <image>
 * 
 * --test 		Run tests
 * --ethereum 	Set up Ethereum
 * --syscoin 	Set up Syscoin
 * --eos		Set up EOS
 * --servers <server1,server1...>
 * --verbose
 */
var (
	VERBOSE bool 
)

 func main(){
 	config := new(Config)

	var test bool
	var optBuildEthereum bool
	var optBuildSyscoin bool
	var optBuildEos bool

	var servers string
	flag.StringVar(&servers,"servers","charlie","The servers to be used")

	flag.IntVar(&config.nodes,"n",30,"Number of Nodes (Docker Containers) in Cluster")
	flag.StringVar(&config.image,"i","whiteblock:latest","The build image to be used")

	flag.BoolVar(&test,"test",false,"Test instead of run")

	flag.BoolVar(&optBuildEthereum,"ethereum",false,"Start up ethereum")
	flag.BoolVar(&optBuildSyscoin,"syscoin",false,"Start up syscoin")
	flag.BoolVar(&optBuildEos,"eos",false,"Start up eos")

	flag.BoolVar(&VERBOSE,"verbose",false,"Make the script extremely verbose")

	flag.Parse()
	
 	//Arguments are now parsed
	switch {
		case test:
			runTests()
 		case optBuildEthereum:
			ethereum(4000000,15468,15468,config.nodes,build(config,getServers(servers)))
		case optBuildSyscoin:
			syscoinRegTest(config.nodes,build(config,getServers(servers)))
		case optBuildEos:
			eos(config.nodes,build(config,getServers(servers)))
		default:
			build(config,getServers(servers))
 	}
}