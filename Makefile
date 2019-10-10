.PHONY: sync build extra price dump full

sync:
	rsync -azvh ./data/ s11993@h6.modhost.pro:/home/s11993/www/sds/
build:
	go build -o ./bin/dumper dumper.go
extra:
	go build -o ./bin/extra extra.go
price:
	go build -o ./bin/pricer pricer.go
dump:
	./bin/dumper
full:
	./bin/dumper && ./bin/extra && ./bin/pricer