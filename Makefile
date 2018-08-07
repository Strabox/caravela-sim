################ CARAVELA's SIMULATOR MAKEFILE ###############
GOCMD=go

######### Builtin GO tools #########
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

############ Output Files ###########
EXE=.exe
BINARY_NAME=caravela_sim$(EXE)

########################## COMMANDS ############################

all: test build

build:
	@echo Building for the current machine settings...
	$(GOBUILD) -o $(BINARY_NAME) -v

clean:
	@echo Cleaning project...
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

install:
	@echo Installing CARAVELA-SIM in the local GO environment...
	$(GOINSTALL) -v -gcflags "-N -l" .

test:
	@echo Testing...
	$(GOTEST) -v ./...

simulate:
	@echo Executing default/example simulation
	$(BINARY_NAME) start