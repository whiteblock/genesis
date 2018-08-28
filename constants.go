package main

/**
 * IP Scheme constants
 */
const	NUMBER_OF_RESERVED_IPS uint32	= 	3
const	SERVER_BITS uint32				=	8 
const	CLUSTER_BITS uint32				=	14
const	NODE_BITS uint32				= 	2
var		NODES_PER_CLUSTER uint32		=   (1 << NODE_BITS) - NUMBER_OF_RESERVED_IPS
/**
 * Concurrency Constants
 */
const 	THREAD_LIMIT int64				=	10

/**
 * Switch Constants 
 */

const	VYOS	int						=   1
const	HP		int						=	2

/**
 * Data Constants
 */

const 	DATA_LOC	string				= 	"~/.dddata"

/**
 * Schema Constants
 */

const 	SWITCH_TABLE	string			= 	"switches"
const 	SERVER_TABLE	string			= 	"servers"
const	TEST_TABLE		string			= 	"testnets"
const	NODES_TABLE		string			= 	"nodes"