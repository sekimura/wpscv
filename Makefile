all: server client

server: server/main.go common/types.go
	@(cd server && go build -o server main.go)
	@touch server/server

client: client/app.go common/types.go
	@(cd client && gopherjs build app.go -m -o ../static/app.js)
	@touch static/app.js
	@touch static/app.js.map

clean:
	@rm -rf server/server static/app.js static/app.js.map

.PHONY: clean server client
