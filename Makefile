all: server appjs

server: server.go model.go
	@go build -o server server.go model.go
	@touch $@

appjs: app.go model.go
	@gopherjs build app.go model.go -m -o static/app.js
	@touch static/app.js
	@touch static/app.js.map

clean:
	@rm -rf server static/app.js static/app.js.map

.PHONY: clean appjs
