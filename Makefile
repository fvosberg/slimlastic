default: install

install:
	go install github.com/fvosberg/slimlastic/cmds/generate && mv $(GOPATH)/bin/generate $(GOPATH)/bin/slimlastic
