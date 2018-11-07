package util

/**
 * IP Scheme constants
 */
const	ReservedIps uint32				= 	3
const	ServerBits uint32				=	8 
const	ClusterBits uint32				=	14
const	NodeBits uint32					= 	2
var		NodesPerCluster uint32			=   (1 << NodeBits) - ReservedIps
/**
 * Concurrency Constants
 */
const 	ThreadLimit int64				=	10

/**
 * Switch Constants 
 */

const	Vyos	int						=   1
const	Hp		int						=	2
