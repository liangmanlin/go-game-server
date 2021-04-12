SHELL := /bin/bash
# print-% : ;@echo $* = $($*)
.PHONY: all dialyzer clean
MAKEFLAGS+= --no-print-director
SERVER_ROOT := .
ifndef OUT_DIR
	OUT_DIR := $(SERVER_ROOT)/bin
endif

NOTMain := $(wildcard tool/*/*.go)

HotFix := $(sort $(wildcard hotfix/*))
HotFix := $(word $(words $(HotFix)),$(HotFix))
HotFix := $(HotFix:%=%/*.go)
HotFix := $(wildcard $(HotFix))

SOURCES := $(wildcard *.go */*.go */*/*.go)

HotFixOut := $(HotFix:%.go=$(OUT_DIR)/%.so)

SOURCES := $(filter-out $(NOTMain) $(HotFix),$(SOURCES))

PROTO_TOOL := $(SERVER_ROOT)/tool/bin/pbBuild
PROTO_SRC := $(wildcard $(SERVER_ROOT)/tool/pbBuild/*.go $(SERVER_ROOT)/tool/pbBuild/*/*.go)

all: mk_dir $(PROTO_TOOL) $(SERVER_ROOT)/proto/pb_auto.go $(OUT_DIR)/main $(HotFixOut) $(GMPD)

$(PROTO_TOOL): $(PROTO_SRC)
	go build -o $@ $(SERVER_ROOT)/tool/pbBuild/*.go

$(SERVER_ROOT)/proto/pb_auto.go: $(SERVER_ROOT)/global/pb_def.go $(PROTO_TOOL)
	@($(PROTO_TOOL) -f $< -o $(SERVER_ROOT)/proto)

$(OUT_DIR)/main: $(SOURCES)
	go build -o $@ main.go

$(OUT_DIR)/%.so: %.go
	go build -buildmode=plugin -o $@ $<

mk_dir:
	@(mkdir -p $(OUT_DIR))
	@(mkdir -p $(OUT_DIR)/hotfix)

clean:
	@(rm -rf $(OUT_DIR)/main)
	@(rm -rf $(OUT_DIR)/gmpd/*)
	@(rm -rf $(OUT_DIR)/hotfix/*)
	@(echo ****clean****)