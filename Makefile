JS_DIR=js
GO_DIR=go
RUST_DIR=rust
PYTHON_DIR=python

.PHONY: all testjs testgo testrust testpy _build car build clean

all: testjs testgo testrust

$(JS_DIR)/node_modules:
	cd $(JS_DIR) && npm install

testjs: $(JS_DIR)/node_modules
	cd $(JS_DIR) && npm test

testgo:
	cd $(GO_DIR) && go test

testrust:
	cd $(RUST_DIR) && cargo test -- --nocapture

testpy:
	cd $(PYTHON_DIR) && pytest

_build:
	cd $(JS_DIR) && npm run build

car: $(JS_DIR)/node_modules
	cd $(JS_DIR) && npm run car

build: $(JS_DIR)/node_modules _build car

clean:
	rm -rf $(JS_DIR)/node_modules
