package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"flag"
	"net"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/influxdata/kapacitor/udf/agent"
)

type myState struct {
	Counter int64
}

// Update the moving average with the next data point.
func (s *myState) update() {
	s.Counter = s.Counter + 1
}

// An Agent.Handler that ...
type handler struct {
	agent *agent.Agent
	state map[string]*myState
}

func newHandler(a *agent.Agent) *handler {
	return &handler{
		state: make(map[string]*myState),
		agent: a,
	}
}

// Return the InfoResponse to describe this UDF agent.
//
// Note that it does not have any option.
func (h *handler) Info() (*agent.InfoResponse, error) {
	info := &agent.InfoResponse{
		Wants:    agent.EdgeType_STREAM,
		Provides: agent.EdgeType_STREAM,
		Options:  map[string]*agent.OptionInfo{},
	}
	return info, nil
}

// Initialze the handler.
func (h *handler) Init(r *agent.InitRequest) (*agent.InitResponse, error) {
	init := &agent.InitResponse{
		Success: true,
		Error:   "",
	}
	return init, nil
}

// This handler does not do batching.
func (h *handler) BeginBatch(*agent.BeginBatch) error {
	return errors.New("batching not supported")
}

// This handler does not do batching.
func (h *handler) EndBatch(*agent.EndBatch) error {
	return errors.New("batching not supported")
}

// Stop the handler gracefully.
func (h *handler) Stop() {
	close(h.agent.Responses)
}

// Create a snapshot of the running state of the process.
func (h *handler) Snapshot() (*agent.SnapshotResponse, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(h.state)

	return &agent.SnapshotResponse{
		Snapshot: buf.Bytes(),
	}, nil
}

// Restore a previous snapshot.
func (h *handler) Restore(req *agent.RestoreRequest) (*agent.RestoreResponse, error) {
	buf := bytes.NewReader(req.Snapshot)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&h.state)
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return &agent.RestoreResponse{
		Success: err == nil,
		Error:   msg,
	}, nil
}

func mapSubexpNames(m, n []string) map[string]string {
	m, n = m[1:], n[1:]
	r := make(map[string]string, len(m))
	for i, _ := range n {
		r[n[i]] = m[i]
	}
	return r
}

// Receive a point and do something with it.
// Send a response with the average value.
func (h *handler) Point(p *agent.Point) error {
	var r = regexp.MustCompile(`(?m).*Kill process (?P<pid>\d+) (?P<proc>\(.*\)) score (?P<score>\d+)`)
	message, ok := p.FieldsString["message"]
	if ok {
		m := r.FindStringSubmatch(message)
		data := mapSubexpNames(m, r.SubexpNames())

		proc := strings.Trim(data["proc"], "()")
		state := h.state[proc]
		if state == nil {
			state := &myState{Counter: 0}
			h.state[proc] = state
		}
		h.state[proc].update()

		newpoint := &agent.Point{
			Time: time.Now().UnixNano(),
			Tags: map[string]string{
				"proc": string(proc),
				"pid":  string(data["pid"]),
			},
			FieldsInt: map[string]int64{
				"count": h.state[proc].Counter,
			},
		}

		// Send point
		h.agent.Responses <- &agent.Response{
			Message: &agent.Response_Point{
				Point: newpoint,
			},
		}
	}

	return nil
}

type accepter struct {
	count int64
}

// Create a new agent/handler for each new connection.
// Count and log each new connection and termination.
func (acc *accepter) Accept(conn net.Conn) {
	count := acc.count
	acc.count++
	a := agent.New(conn, conn)
	h := newHandler(a)
	a.Handler = h

	log.WithField("connections", count).Info("Starting agent for connection")
	a.Start()
	go func() {
		err := a.Wait()
		if err != nil {
			log.Fatal(err)
		}
		log.WithField("connections", count).Info("Agent for connection finished")
	}()
}

var socketPath = flag.String("socket", "/tmp/example.sock", "Where to create the unix socket.")

func main() {
	addr, err := net.ResolveUnixAddr("unix", *socketPath)
	if err != nil {
		log.Fatal(err)
	}

	syscall.Unlink(*socketPath)

	l, err := net.ListenUnix("unix", addr)
	if err != nil {
		log.Fatal(err)
	}

	s := agent.NewServer(l, &accepter{})

	s.StopOnSignals(os.Interrupt, syscall.SIGTERM)

	log.WithField("address", addr.String()).Infoln("Server listening")
	err = s.Serve()
	if err != nil {
		log.Fatal(err)
	}
	log.WithField("address", addr.String()).Infoln("Server stopped")
}
