package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"reflect"
)

func DoCoordination(view int, startIndex int, endIndex int, entries []msgs.Entry, io *msgs.Io, config Config, prepare bool) bool {
	// PHASE 2: prepare
	if prepare {

		// check that committed is not set
		for i := 0; i < endIndex - startIndex; i++ {
			entries[i].Committed = false
		}

		prepare := msgs.PrepareRequest{config.ID, view, startIndex, endIndex, entries}
		glog.Info("Starting prepare phase", prepare)
		(*io).OutgoingBroadcast.Requests.Prepare <- prepare

		// collect responses
		glog.Info("Waiting for ", config.Quorum.ReplicateSize, " prepare responses")
		for replied := make([]bool,config.N); !config.Quorum.checkReplicationQuorum(replied); {
			msg := <-(*io).Incoming.Responses.Prepare
			// check msg replies to the msg we just sent
			if reflect.DeepEqual(msg.Request, prepare) {
				glog.Info("Received ", msg)
				if !msg.Response.Success {
					glog.Warning("Coordinator is stepping down")
					return false
				}
				replied[msg.Response.SenderID] = true
				glog.Info("Successful response received, waiting for more")
			}
		}
	}

	// PHASE 3: commit
	// set committed so requests will be applied to state machines
	for i := 0; i < endIndex - startIndex; i++ {
		entries[i].Committed = true
	}
	// dispatch commit requests to all
	commit := msgs.CommitRequest{config.ID, startIndex, endIndex, entries}
	glog.Info("Starting commit phase", commit)
	(*io).OutgoingBroadcast.Requests.Commit <- commit

	// TODO: handle replies properly
	go func() {
		for replied := make([]bool,config.N); !config.Quorum.checkReplicationQuorum(replied); {
			msg := <-(*io).Incoming.Responses.Commit
			// check msg replies to the msg we just sent
			if reflect.DeepEqual(msg.Request, commit) {
				glog.Info("Received ", msg)
			}
			replied[msg.Response.SenderID] = true
		}
	}()

	return true
}

// returns true if successful
func RunCoordinator(state *State, io *msgs.Io, config Config) {

	for {
		glog.Info("Coordinator is ready to handle request")
		req := <-(*io).Incoming.Requests.Coordinate
		success := DoCoordination(req.View, req.StartIndex, req.EndIndex, req.Entries, io, config, req.Prepare)
		reply := msgs.CoordinateResponse{config.ID, success}
		(*io).OutgoingUnicast[req.SenderID].Responses.Coordinate <- msgs.Coordinate{req, reply}
		glog.Info("Coordinator is finished handling request")
		// TOD0: handle failure
	}
}
