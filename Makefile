################ CARAVELA's SIMULATOR MAKEFILE ###############
GOCMD=go

######### Builtin GO tools #########
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean

############ Output Files ###########
EXE=.exe
BINARY_NAME=caravela_sim$(EXE)

########################## COMMANDS ############################

all: build

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
	@echo Testing CARAVELA-SIM...
	$(BINARY_NAME) start

