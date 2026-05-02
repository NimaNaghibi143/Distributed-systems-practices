package goraft

import (
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type StateMachine interface {
	Apply(cmd []byte) ([]byte, error)
}

type ApplyResult struct {
	Result []byte
	Error  error
}

type Entry struct {
	Command []byte
	Term    uint64

	// Set by the primary so it can learn about the result of
	// applying this command to the state machine
	result chan ApplyResult
}

type ClusterMember struct {
	Id      uint64
	Address string

	// Index of the next log entry to send
	nextIndex uint64
	// Hghest log entry known to be replicated
	matchIndex uint64

	// Whos was voted for in the most recent term
	votedFor uint64

	// TCP connection
	rcpClient *rpc.Client
}

type ServerState string

const (
	leaderState    ServerState = "leader"
	followerState  ServerState = "follower"
	candidateStare ServerState = "candidate"
)

type Server struct {
	// These variables are for shutting down.
	done   bool
	server *http.Server
	debug  bool
	mu     sync.Mutex

	// ------ Persistent state ------

	// The current term
	currentTerm uint64
	log         []Entry

	// votedFor is stored in `cluster []ClusterMember` below,
	// mapped by `clusterIndex`

	// ------ Readonly state ------

	// unique identifier for this server
	id uint64

	// The TCP address for RPC
	address string

	// When to start elections after no append entry messages
	electionTimeout time.Time

	// How often to send empty messages
	heartBeatMs int

	// When to next send empty message
	heartBeatTimeout time.Time

	// User-provided StateMachine
	statemachine StateMachine

	// Metadata directory
	metadataDir string

	// Metadata store

	fd *os.File

	// ------ Volatile state ------

	// Index of highest log entry known to be committed
	commitIndex uint64

	// Index of highest log entry applied to state machine
	lastApplied uint64

	// Candidate, follower, or leader
	state ServerState

	// Servers in the cluster, including this one
	cluster []ClusterMember

	// Index of this server
	clusterIndex int
}
