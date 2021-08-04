all: testjs testgo

js/node_modules:
	cd js && npm install

testjs: js/node_modules
	cd js && npm test

testgo:
	cd go && go test