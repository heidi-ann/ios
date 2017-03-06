// Package store provides a simple key value store
// Not safe for concurreny access
// TODO: replace map with https://github.com/orcaman/concurrent-map
package app

import (
	"github.com/golang/glog"
	"strconv"
	"strings"
)

type Store map[string]string

func newStore() *Store {
	var s Store
	s = map[string]string{}
	return &s
}

func (s *Store) execute(req string) string {
	request := strings.Split(req, " ")

	switch request[0] {
	case "update":
		if len(request) != 3 {
			return "not recognised"
		}
		glog.Infof("Updating %s to %s", request[1], request[2])
		(*s)[request[1]] = request[2]
		return "OK"
	case "exists":
		if len(request) != 2 {
			return "request not recognised"
		}
		glog.Infof("Checking if %s exists", request[1])
		_, exists := (*s)[request[1]]
		return strconv.FormatBool(exists)
	case "get":
		if len(request) != 2 {
			return "request not recognised"
		}
		glog.Infof("Getting %s", request[1])
		value, ok := (*s)[request[1]]
		if ok {
			return value
		} else {
			return "key not found"
		}
	case "delete":
		if len(request) != 2 {
			return "request not recognised"
		}
		glog.Infof("Deleting %s", request[1])
		delete(*s, request[1])
		return "OK"
	case "count":
		if len(request) != 1 {
			return "request not recognised"
		}
		glog.Infof("Counting size of key-value store")
		return strconv.Itoa(len(*s))
	case "print":
		if len(request) != 1 {
			return "request not recognised"
		}
		glog.Infof("Printing key-value store")
		return s.print()
	default:
		return "request not recognised"
	}
}

func (s *Store) process(req string) string {
	reqs := strings.Split(strings.Trim(req, "\n"), "; ")
	var reply string

	for i := range reqs {
		if i == 0 {
			reply = s.execute(reqs[i])
		} else {
			reply = reply + "; " + s.execute(reqs[i])
		}
	}
	return reply
}

func (s *Store) print() string {
	str := ""
	for key, value := range *s {
		str += key + ", " + value + "\n"
	}
	return str
}