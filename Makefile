all: testjs testgo testrust

js/node_modules:
	cd js && npm install

testjs: js/node_modules
	cd js && npm test

testgo:
	cd go && go test

testrust:
	cd rust && cargo test -- --nocapture
