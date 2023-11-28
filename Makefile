all: testjs testgo testrust

js/node_modules:
	cd js && npm install

testjs: js/node_modules
	cd js && npm test

testgo:
	cd go && go test

testrust:
	cd rust && cargo test -- --nocapture

_build:
	cd js && npm run build

car: js/node_modules
	cd js && npm run car

build: js/node_modules _build car
